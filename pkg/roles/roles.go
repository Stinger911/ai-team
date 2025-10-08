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

	"ai-team/pkg/logger"

	"github.com/sirupsen/logrus"
)

// executeConfiguredTool executes a tool by name using the arguments and the list of configurable tools.
// It returns the tool response (if any) as interface{} for use in prompt context.
func executeConfiguredTool(toolName string, args map[string]interface{}, configurableTools []types.ConfigurableTool) interface{} {
	switch toolName {
	case "write_file":
		filePath, _ := args["file_path"].(string)
		content, _ := args["content"].(string)
		if filePath != "" {
			_, _ = tools.WriteFile(filePath, content)
			return map[string]interface{}{"file_path": filePath, "content": content}
		} else {
			logrus.Warn("[ToolCall] file_path is empty, skipping file write")
			return map[string]interface{}{"error": "file_path is empty"}
		}
	case "list_dir":
		path, _ := args["path"].(string)
		if path != "" {
			for _, t := range configurableTools {
				if t.Name == "list_dir" {
					cmd := strings.ReplaceAll(t.CommandTemplate, "{{.path}}", path)
					out, err := tools.RunCommand(cmd)
					if err != nil {
						logrus.Errorf("[ToolCall] list_dir failed: %v", err)
						return map[string]interface{}{"error": err.Error()}
					} else {
						logrus.Infof("[ToolCall] list_dir output:\n%s", out)
						return map[string]interface{}{"output": out}
					}
				}
			}
			return map[string]interface{}{"error": "list_dir tool config not found"}
		} else {
			logrus.Warn("[ToolCall] path is empty, skipping list_dir")
			return map[string]interface{}{"error": "path is empty"}
		}
	case "read_file":
		filePath, _ := args["file_path"].(string)
		if filePath != "" {
			for _, t := range configurableTools {
				if t.Name == "read_file" {
					cmd := strings.ReplaceAll(t.CommandTemplate, "{{.file_path}}", filePath)
					out, err := tools.RunCommand(cmd)
					if err != nil {
						logrus.Errorf("[ToolCall] read_file failed: %v", err)
						return map[string]interface{}{"error": err.Error()}
					} else {
						logrus.Infof("[ToolCall] read_file output:\n%s", out)
						return map[string]interface{}{"output": out}
					}
				}
			}
			return map[string]interface{}{"error": "read_file tool config not found"}
		} else {
			logrus.Warn("[ToolCall] file_path is empty, skipping read_file")
			return map[string]interface{}{"error": "file_path is empty"}
		}
	default:
		logrus.Warnf("[ToolCall] Tool '%s' not implemented for execution", toolName)
		return map[string]interface{}{"error": fmt.Sprintf("Tool '%s' not implemented", toolName)}
	}
	// Should not reach here
	// return nil
}

// extractFirstJSON extracts the first valid JSON object from a string, even if surrounded by markdown/code blocks or extra text.
func extractFirstJSON(s string) string {
	// Remove code block markers if present
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSpace(s)
	}
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
	}
	// Find the first '{' and the last '}'
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		return s[start : end+1]
	}
	return s
}

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

	var response string
	var roleErr error

	// Call the AI model based on the role's model
	// Currently only Gemini is supported for roles
	// (Future: Add cases for OpenAI, Ollama, etc.)
	client := &http.Client{}
	switch role.Model {
	case "Gemini":
		response, roleErr = ai.CallGemini(
			client,
			processedPrompt.String(),
			cfg.Gemini.Model, // Use the model specified in the role
			cfg.Gemini.APIURL,
			cfg.Gemini.APIKey,
			cfg.Tools,
		)
	case "Ollama":
		response, roleErr = ai.CallOllama(
			client,
			processedPrompt.String(),
			cfg.Ollama.APIURL,
			cfg.Ollama.Model,
			cfg.Tools,
		)
	default:
		return "", errors.New(errors.ErrCodeRole, fmt.Sprintf("unsupported model '%s' in role '%s'", role.Model, role.Name), nil)
	}

	// Log the role call
	logEntry := types.RoleCallLogEntry{
		RoleName: role.Name,
		Input:    input,
		Output:   response,
	}
	if roleErr != nil {
		logEntry.Error = roleErr.Error()
	}
	if logFilePath != "" {
		if logErr := logger.LogRoleCall(logFilePath, logEntry); logErr != nil {
			logrus.WithError(logErr).Warn("Failed to log role call")
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
	var roles []types.Role = cfg.Roles
	logrus.Debugf("Executing chain: %s", chain.Name)
	logrus.Debugf("Roles: %v", roles)
	var configurableTools []types.ConfigurableTool = cfg.Tools

	context := make(map[string]interface{})
	for k, v := range initialInput {
		context[k] = v
	}

	var lastToolResponse interface{} = nil
	for _, chainRole := range chain.Roles {
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
			// Find the role definition
			var currentRole types.Role
			found := false
			for _, r := range roles {
				if r.Name == chainRole.Name {
					currentRole = r
					if config.IsModelDefined(r.Model, cfg) {
						found = true
					}
					logrus.Debugf("Found role: %s with model: %s", r.Name, r.Model)
					break
				}
			}
			if !found {
				return nil, errors.New(errors.ErrCodeRole, fmt.Sprintf("role %s not found in chain %s", chainRole.Name, chain.Name), nil)
			}

			// Prepare input for the current role
			roleInput := make(map[string]interface{})
			for k, v := range chainRole.Input {
				// Resolve input from context if it's a template
				if strVal, ok := v.(string); ok && strings.HasPrefix(strVal, "{{") && strings.HasSuffix(strVal, "}}") {
					tmpl, err := template.New("input").Parse(strVal)
					if err != nil {
						return nil, errors.New(errors.ErrCodeRole, fmt.Sprintf("failed to parse input template for role %s in chain %s", chainRole.Name, chain.Name), err)
					}
					var resolvedInput bytes.Buffer
					if err := tmpl.Execute(&resolvedInput, context); err != nil {
						return nil, errors.New(errors.ErrCodeRole, fmt.Sprintf("failed to execute input template for role %s in chain %s", chainRole.Name, chain.Name), err)
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

			logrus.Infof("Executing role: %s (loop %d/%d) with input: %v", currentRole.Name, i+1, loopCount, roleInput)
			rawOutput, err := ExecuteRole(currentRole, roleInput, cfg, logFilePath)
			output := extractFirstJSON(rawOutput)
			if err != nil {
				logrus.WithError(err).Errorf("Failed to execute role %s in chain %s", currentRole.Name, chain.Name)
				// Log the failed role call
				if logFilePath != "" {
					logEntry := types.RoleCallLogEntry{
						RoleName: currentRole.Name,
						Input:    roleInput,
						Output:   output,
						Error:    err.Error(),
					}
					_ = logger.LogRoleCall(logFilePath, logEntry)
				}
				return nil, errors.New(errors.ErrCodeRole, "failed to execute role "+currentRole.Name+" in chain "+chain.Name, err)
			}

			// Try to execute a normal tool call if present (use rawOutput)
			var toolCallObj struct {
				ToolName  string                 `json:"tool_name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			toolCallErr := json.Unmarshal([]byte(rawOutput), &toolCallObj)
			if toolCallErr == nil && toolCallObj.ToolName != "" {
				logrus.Infof("[ToolCall] Executing tool: %s", toolCallObj.ToolName)
				lastToolResponse = executeConfiguredTool(toolCallObj.ToolName, toolCallObj.Arguments, configurableTools)
			} else {
				// Try tool_call (Gemini style, use rawOutput)
				var toolCallWrap struct {
					ToolCall struct {
						Name      string                 `json:"name"`
						Arguments map[string]interface{} `json:"arguments"`
					} `json:"tool_call"`
				}
				if err := json.Unmarshal([]byte(rawOutput), &toolCallWrap); err == nil && toolCallWrap.ToolCall.Name != "" {
					logrus.Infof("[ToolCallWrap] Executing tool: %s", toolCallWrap.ToolCall.Name)
					lastToolResponse = executeConfiguredTool(toolCallWrap.ToolCall.Name, toolCallWrap.ToolCall.Arguments, configurableTools)
				} else {
					// Fallback: if output is a JSON object with file_path and content, write the file (use extractFirstJSON output)
					var fileObj struct {
						FilePath string `json:"file_path"`
						Content  string `json:"content"`
					}
					if err := json.Unmarshal([]byte(output), &fileObj); err == nil && fileObj.FilePath != "" {
						logrus.Debugf("[Fallback] fileObj: file_path=%s, content-len=%d", fileObj.FilePath, len(fileObj.Content))
						logrus.Infof("[Fallback] Writing file: %s", fileObj.FilePath)
						_, _ = tools.WriteFile(fileObj.FilePath, fileObj.Content)
						lastToolResponse = map[string]interface{}{"file_path": fileObj.FilePath, "content": fileObj.Content}
					}
				}
			}
		}
	}

	return context, nil
}
