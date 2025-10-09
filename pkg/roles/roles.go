package roles

import (
	"ai-team/config"
	"ai-team/pkg/ai"
	"ai-team/pkg/errors"
	"ai-team/pkg/tools"
	"ai-team/pkg/types"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"ai-team/pkg/logger"
)

// ExecuteRole executes a single AI role.
func ExecuteRole(
	role types.Role,
	input map[string]interface{},
	cfg config.Config,
	logFilePath string, // Add logFilePath parameter
) (string, error) {
	// Render the prompt with the provided input
	tmpl, err := template.New("prompt").Parse(role.Prompt)
	if err != nil {
		return "", errors.New(errors.ErrCodeRole, "failed to parse role prompt template", err)
	}

	var processedPrompt bytes.Buffer
	if err := tmpl.Execute(&processedPrompt, input); err != nil {
		return "", errors.New(errors.ErrCodeRole, "failed to execute role prompt template", err)
	}

	// Call the AI model based on the role's model
	// Currently only Gemini is supported for roles
	// (Future: Add cases for OpenAI, Ollama, etc.)
	client := &http.Client{}

	// Determine provider and model config
	var response string
	var roleErr error

	// Debug: print available Gemini model keys, full map, and requested model name
	logger.DebugPrintf("Gemini models available: %v", keys(cfg.Gemini.Models))
	logger.DebugPrintf("Gemini models map: %#v", cfg.Gemini.Models)
	logger.DebugPrintf("Requested Gemini model: '%s'", role.Model)
	// Try Gemini
	if modelCfg, ok := cfg.Gemini.Models[role.Model]; ok {
		apiKey := modelCfg.Apikey
		if apiKey == "" {
			apiKey = cfg.Gemini.Apikey
		}
		apiURL := modelCfg.Apiurl
		if apiURL == "" {
			apiURL = cfg.Gemini.Apiurl
		}
		response, roleErr = ai.CallGeminiFunc(
			client,
			processedPrompt.String(),
			modelCfg.Model,
			apiURL,
			apiKey,
			cfg.Tools,
		)
	} else if modelCfg, ok := cfg.OpenAI.Models[role.Model]; ok {
		apiKey := modelCfg.Apikey
		if apiKey == "" {
			apiKey = cfg.OpenAI.Apikey
		}
		apiURL := modelCfg.Apiurl
		if apiURL == "" {
			apiURL = cfg.OpenAI.DefaultApiurl
		}
		response, roleErr = ai.CallOpenAI(
			client,
			processedPrompt.String(),
			apiURL,
			apiKey,
		)
	} else if modelCfg, ok := cfg.Ollama.Models[role.Model]; ok {
		apiURL := modelCfg.Apiurl
		if apiURL == "" {
			apiURL = cfg.Ollama.Apiurl
		}
		response, roleErr = ai.CallOllama(
			client,
			processedPrompt.String(),
			apiURL,
			modelCfg.Model,
			cfg.Tools,
		)
	} else {
		return "", errors.New(errors.ErrCodeRole, fmt.Sprintf("unsupported or undefined model '%s' in role", role.Model), nil)
	}

	// Log the role call
	logEntry := types.RoleCallLogEntry{
		RoleName: role.Model, // Use model name as identifier
		Input:    input,
		Output:   response,
	}
	if roleErr != nil {
		logEntry.Error = roleErr.Error()
	}
	if logFilePath != "" {
		if logErr := logger.LogRoleCall(logFilePath, logEntry); logErr != nil {
			logger.DebugPrintf("Failed to log role call: %v", logErr)
		}
	}

	// Always sanitize output for downstream use
	cleanResponse := extractFirstJSON(response)
	return cleanResponse, roleErr
}

// ExecuteChain executes a chain of AI roles.
func ExecuteChain(
	chain types.RoleChain,
	initialInput map[string]interface{},
	cfg config.Config,
	logFilePath string, // Add logFilePath parameter
) (map[string]interface{}, error) {
	roles := cfg.Roles
	logger.DebugPrintf("Executing chain (steps): %+v", chain.Steps)
	logger.DebugPrintf("Roles: %v", roles)
	// Initialize ToolRegistry and ToolExecutor for the chain
	toolRegistry := tools.NewToolRegistry()
	tools.RegisterDefaultTools(toolRegistry)
	toolExecutor := &tools.ToolExecutor{
		Registry:   toolRegistry,
		Logger:     nil, // Use default logger or inject as needed
		RetryCount: 1,
		Timeout:    10 * time.Second,
	}

	context := make(map[string]interface{})
	for k, v := range initialInput {
		context[k] = v
	}

	var lastToolResponse interface{} = nil
	for _, chainRole := range chain.Steps {
		loopCount := 1
		maxLoop := 100 // Prevent infinite loops
		if chainRole.Loop {
			if chainRole.LoopCount > 0 {
				loopCount = chainRole.LoopCount
			} else if chainRole.LoopCondition != "" {
				loopCount = maxLoop // Use maxLoop if only LoopCondition is set
			} else {
				loopCount = 1 // Default to 1 if not specified
			}
		}
		for i := 0; i < loopCount; i++ {
			// Look up the role by key from the map, prefer 'Role' field (YAML 'role')
			roleKey := chainRole.Role
			if roleKey == "" {
				roleKey = chainRole.Name
			}
			roleDef, ok := roles[roleKey]
			if !ok {
				return nil, errors.New(errors.ErrCodeRole, fmt.Sprintf("role '%s' not found in config", roleKey), nil)
			}
			logger.DebugPrintf("Found role: %s with model: %s", roleKey, roleDef.Model)

			// Prepare input for the current role
			roleInput := make(map[string]interface{})
			for k, v := range chainRole.Input {
				// Resolve input from context if it's a template
				if strVal, ok := v.(string); ok && strings.HasPrefix(strVal, "{{") && strings.HasSuffix(strVal, "}}") {
					tmpl, err := template.New("input").Parse(strVal)
					if err != nil {
						return nil, errors.New(errors.ErrCodeRole, fmt.Sprintf("failed to parse input template for role %s in chain", roleKey), err)
					}
					var resolvedInput bytes.Buffer
					if err := tmpl.Execute(&resolvedInput, context); err != nil {
						return nil, errors.New(errors.ErrCodeRole, fmt.Sprintf("failed to execute input template for role %s in chain", roleKey), err)
					}
					roleInput[k] = resolvedInput.String()
				} else {
					roleInput[k] = v
				}
			}
			// Inject most recent tool response
			if lastToolResponse != nil {
				roleInput["lastToolResponse"] = lastToolResponse
			}

			logger.DebugPrintf("Executing role: %s (loop %d/%d) with input: %v", roleKey, i+1, loopCount, roleInput)
			rawOutput, err := ExecuteRole(roleDef, roleInput, cfg, logFilePath)
			output := extractFirstJSON(rawOutput)
			if err != nil {
				logger.DebugPrintf("Failed to execute role %s in chain: %v", roleKey, err)
				// Log the failed role call
				if logFilePath != "" {
					logEntry := types.RoleCallLogEntry{
						RoleName: roleKey,
						Input:    roleInput,
						Output:   output,
						Error:    err.Error(),
					}
					_ = logger.LogRoleCall(logFilePath, logEntry)
				}
				return nil, errors.New(errors.ErrCodeRole, "failed to execute role "+roleKey+" in chain", err)
			}

			// Try to execute a normal tool call if present (use rawOutput)
			var toolCallObj struct {
				ToolName  string                 `json:"tool_name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			toolCallErr := json.Unmarshal([]byte(rawOutput), &toolCallObj)
			if toolCallErr == nil && toolCallObj.ToolName != "" {
				logger.DebugPrintf("[ToolCall] Executing tool: %s", toolCallObj.ToolName)
				lastToolResponse = executeConfiguredTool(toolCallObj.ToolName, toolCallObj.Arguments, toolRegistry, toolExecutor)
			} else {
				// Try tool_call (Gemini style, use rawOutput)
				var toolCallWrap struct {
					ToolCall struct {
						Name      string                 `json:"name"`
						Arguments map[string]interface{} `json:"arguments"`
					} `json:"tool_call"`
				}
				if err := json.Unmarshal([]byte(rawOutput), &toolCallWrap); err == nil && toolCallWrap.ToolCall.Name != "" {
					logger.DebugPrintf("[ToolCallWrap] Executing tool: %s", toolCallWrap.ToolCall.Name)
					lastToolResponse = executeConfiguredTool(toolCallWrap.ToolCall.Name, toolCallWrap.ToolCall.Arguments, toolRegistry, toolExecutor)
				} else {
					// Fallback: if output is a JSON object with file_path and content, write the file (use extractFirstJSON output)
					var fileObj struct {
						FilePath string `json:"file_path"`
						Content  string `json:"content"`
					}
					if err := json.Unmarshal([]byte(output), &fileObj); err == nil && fileObj.FilePath != "" {
						logger.DebugPrintf("[Fallback] fileObj: file_path=%s, content-len=%d", fileObj.FilePath, len(fileObj.Content))
						logger.DebugPrintf("[Fallback] Writing file: %s", fileObj.FilePath)
						_, _ = tools.WriteFile(fileObj.FilePath, fileObj.Content)
						lastToolResponse = map[string]interface{}{"file_path": fileObj.FilePath, "content": fileObj.Content}
					}
				}
			}
			// Store output in context if OutputKey is set (immediately after output is set)
			if chainRole.OutputKey != "" {
				context[chainRole.OutputKey] = output
			}
		}
	}
	return context, nil
}

// extractFirstJSON extracts the first JSON object from a string, handling markdown code blocks.
func extractFirstJSON(s string) string {
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSpace(s)
	}
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		return s[start : end+1]
	}
	return s
}

// executeConfiguredTool executes a tool by name from the list of configurable tools.
// Refactored: executeConfiguredTool uses ToolRegistry and ToolExecutor for robust orchestration.
func executeConfiguredTool(toolName string, args map[string]interface{}, reg *tools.ToolRegistry, exec *tools.ToolExecutor) interface{} {
	call := tools.ToolCall{
		Name:      toolName,
		Arguments: args,
	}
	if err := reg.ValidateToolCall(call); err != nil {
		return map[string]interface{}{
			"error":            "tool call validation failed",
			"tool":             toolName,
			"validation_error": err.Error(),
		}
	}
	result, err := exec.Execute(call)
	if err != nil {
		return map[string]interface{}{
			"error":      "tool execution failed",
			"tool":       toolName,
			"exec_error": err.Error(),
		}
	}
	return result
}

// keys returns the keys of a map[string]T as a []string
func keys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
