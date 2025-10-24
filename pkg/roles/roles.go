package roles

import (
	"ai-team/config"
	ai "ai-team/pkg/ai"
	"ai-team/pkg/errors"
	"ai-team/pkg/tools"
	"ai-team/pkg/types"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"ai-team/pkg/logger"
)

// ExecuteRole executes a single AI role.
func ExecuteRole(
	role types.Role,
	input map[string]interface{},
	cfg *config.Config,
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

	switch role.Provider {
	case "gemini":
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
		} else {
			return "", errors.New(errors.ErrCodeRole, fmt.Sprintf("Gemini model '%s' not found in config", role.Model), nil)
		}
	case "openai":
		logger.DebugPrintf("Looking for OpenAI model %q in map with keys: %q", role.Model, keys(cfg.OpenAI.Models))
		if modelCfg, ok := cfg.OpenAI.Models[role.Model]; ok {
			logger.DebugPrintf("OpenAI model '%s' found: %t", role.Model, ok)
			apiKey := modelCfg.Apikey
			if apiKey == "" {
				apiKey = cfg.OpenAI.Apikey
			}
			apiURL := modelCfg.Apiurl
			if apiURL == "" {
				apiURL = cfg.OpenAI.DefaultApiurl
			}
			response, roleErr = ai.CallOpenAIFunc(
				client,
				processedPrompt.String(),
				apiURL,
				apiKey,
			)
		} else {
			return "", errors.New(errors.ErrCodeRole, fmt.Sprintf("OpenAI model '%s' not found in config", role.Model), nil)
		}
	case "ollama":
		if modelCfg, ok := cfg.Ollama.Models[role.Model]; ok {
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
			return "", errors.New(errors.ErrCodeRole, fmt.Sprintf("Ollama model '%s' not found in config", role.Model), nil)
		}
	default:
		return "", errors.New(errors.ErrCodeRole, fmt.Sprintf("unsupported or undefined provider '%s' for model '%s'", role.Provider, role.Model), nil)
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

	// Use ToolCallExtractor for robust extraction with schema validation
	toolRegistry := tools.NewToolRegistry()
	tools.RegisterDefaultTools(toolRegistry)
	extractor := ai.NewDefaultToolCallExtractor(toolRegistry)
	tc, _, err := extractor.ExtractToolCall(response)
	if err == nil && tc != nil {
		// If a tool-call is found, return its JSON
		b, _ := json.Marshal(tc)
		return string(b), roleErr
	}
	// Fallback: extract first JSON object (legacy)
	cleanResponse := response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start != -1 && end != -1 && end > start {
		cleanResponse = response[start : end+1]
	}
	return cleanResponse, roleErr
}

// ExecuteChain executes a chain of AI roles.
func ExecuteChain(
	chain types.RoleChain,
	initialInput map[string]interface{},
	cfg *config.Config,
	logFilePath string, // Add logFilePath parameter
) (map[string]interface{}, error) {
	roles := cfg.Roles
	logger.DebugPrintf("Executing chain (steps): %+v", chain.Steps)
	logger.DebugPrintf("Roles: %v", roles)
	// Initialize ToolRegistry and ToolExecutor for the chain
	toolRegistry := tools.NewToolRegistry()
	tools.RegisterDefaultTools(toolRegistry)
	// toolExecutor removed (was unused)

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

			logger.DebugPrintf("Preparing to execute role: %s (loop %d/%d) with input: %v", roleKey, i+1, loopCount, roleInput)
			// Inject lastToolResponse just before role execution, after any tool execution from previous step
			roleInput["lastToolResponse"] = lastToolResponse
			// Also provide a JSON-stringified version for easy templating in prompts
			if lastToolResponse != nil {
				if b, err := json.Marshal(lastToolResponse); err == nil {
					roleInput["lastToolResponse_json"] = string(b)
				} else {
					roleInput["lastToolResponse_json"] = fmt.Sprintf("%v", lastToolResponse)
				}
			} else {
				roleInput["lastToolResponse_json"] = ""
			}

			logger.DebugPrintf("Executing role: %s (loop %d/%d) with input: %v", roleKey, i+1, loopCount, roleInput)
			rawOutput, _ := ExecuteRole(roleDef, roleInput, cfg, logFilePath)
			// Try to extract tool call from Gemini response's text field if present
			var toolCallText string
			var output string
			// Try to parse as Gemini response
			type geminiPart struct {
				Text string `json:"text"`
			}
			type geminiContent struct {
				Parts []geminiPart `json:"parts"`
			}
			type geminiCandidate struct {
				Content geminiContent `json:"content"`
			}
			type geminiResponse struct {
				Candidates []geminiCandidate `json:"candidates"`
			}
			var gemResp geminiResponse
			if err := json.Unmarshal([]byte(rawOutput), &gemResp); err == nil && len(gemResp.Candidates) > 0 && len(gemResp.Candidates[0].Content.Parts) > 0 {
				toolCallText = gemResp.Candidates[0].Content.Parts[0].Text
			} else {
				toolCallText = rawOutput
			}
			extractor := ai.NewDefaultToolCallExtractor(toolRegistry)
			tc, _, errExtract := extractor.ExtractToolCall(toolCallText)
			if errExtract == nil && tc != nil {
				b, _ := json.Marshal(tc)
				output = string(b)
				// expose the parsed tool_call in the context for loop_condition templates
				context["tool_call"] = map[string]interface{}{"name": tc.Name, "arguments": tc.Arguments}
				// Inline tool execution logic
				toolExecutor := &tools.ToolExecutor{
					Registry:   toolRegistry,
					Logger:     nil,
					RetryCount: 1,
					Timeout:    0,
				}
				call := tools.ToolCall{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				}
				result, err := toolExecutor.Execute(call)
				if err != nil {
					lastToolResponse = map[string]interface{}{
						"error":      "tool execution failed",
						"tool":       tc.Name,
						"exec_error": err.Error(),
					}
				} else {
					lastToolResponse = result
				}
				logger.DebugPrintf("[Chain] lastToolResponse after executing tool %s: %v", tc.Name, lastToolResponse)
			} else {
				// Fallback: extract first JSON object (legacy)
				output = toolCallText
				start := strings.Index(toolCallText, "{")
				end := strings.LastIndex(toolCallText, "}")
				if start != -1 && end != -1 && end > start {
					output = toolCallText[start : end+1]
				}
				// Try to parse as a legacy tool call (file_path/content)
				var fileObj struct {
					FilePath string `json:"file_path"`
					Content  string `json:"content"`
				}
				if err := json.Unmarshal([]byte(output), &fileObj); err == nil && fileObj.FilePath != "" {
					logger.DebugPrintf("[Fallback] fileObj: file_path=%s, content-len=%d", fileObj.FilePath, len(fileObj.Content))
					logger.DebugPrintf("[Fallback] Writing file: %s", fileObj.FilePath)
					_, _ = tools.WriteFile(fileObj.FilePath, fileObj.Content)
					lastToolResponse = map[string]interface{}{"file_path": fileObj.FilePath, "content": fileObj.Content}
				} else {
					lastToolResponse = nil
					// clear any tool_call context when no tool was found
					delete(context, "tool_call")
				}
			}
			// Store output in context if OutputKey is set (immediately after output is set)
			if chainRole.OutputKey != "" {
				// If lastToolResponse is from write_file and has content, store the content directly
				if lastToolResponse != nil {
					if respMap, ok := lastToolResponse.(map[string]interface{}); ok {
						if content, ok := respMap["content"]; ok {
							if strContent, ok := content.(string); ok && strContent != "" {
								context[chainRole.OutputKey] = strContent
							} else {
								context[chainRole.OutputKey] = output
							}
						} else {
							context[chainRole.OutputKey] = output
						}
					} else {
						context[chainRole.OutputKey] = output
					}
				} else {
					context[chainRole.OutputKey] = output
				}
			}
			logger.DebugPrintf("[Chain] lastToolResponse after executing tool %s: %v", roleKey, lastToolResponse)

			// If a loop condition is provided on the chain role, evaluate it now. If it evaluates
			// to true, break out of the inner loop early.
			if chainRole.LoopCondition != "" {
				ok, err := evaluateLoopCondition(chainRole.LoopCondition, context)
				if err != nil {
					logger.DebugPrintf("Failed to evaluate loop_condition '%s': %v", chainRole.LoopCondition, err)
				} else if ok {
					logger.DebugPrintf("Loop condition evaluated true, breaking loop for role %s", roleKey)
					break
				}
			}
		}
	}
	return context, nil
}

// keys returns the keys of a map[string]T as a []string
func keys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// evaluateLoopCondition renders the loop_condition template using the provided context
// and evaluates simple expressions. Supported forms after rendering:
//   - "true" / "false" (case-insensitive)
//   - "<left> == '<right>'" or "<left> != '<right>'"
//
// For equality checks, surrounding quotes are optional for the right-hand side.
func evaluateLoopCondition(condTemplate string, context map[string]interface{}) (bool, error) {
	// Render template
	tmpl, err := template.New("loop_condition").Parse(condTemplate)
	if err != nil {
		return false, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return false, err
	}
	rendered := strings.TrimSpace(buf.String())
	lower := strings.ToLower(rendered)
	if lower == "true" {
		return true, nil
	}
	if lower == "false" || rendered == "" {
		return false, nil
	}
	// try equality / inequality
	if strings.Contains(rendered, "==") {
		parts := strings.SplitN(rendered, "==", 2)
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		right = strings.Trim(right, " \"'")
		return left == right, nil
	}
	if strings.Contains(rendered, "!=") {
		parts := strings.SplitN(rendered, "!=", 2)
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		right = strings.Trim(right, " \"'")
		return left != right, nil
	}
	// not recognized -> false
	return false, nil
}
