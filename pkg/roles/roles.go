package roles

import (
	"ai-team/pkg/ai"
	"ai-team/pkg/errors"
	"ai-team/pkg/logger"
	"ai-team/pkg/tools"
	"ai-team/pkg/types"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

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
	geminiAPIURL string,
	geminiAPIKey string,
	configurableTools []types.ConfigurableTool,
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
	response, roleErr = ai.CallGeminiFunc(
		client,
		processedPrompt.String(),
		role.Model, // Use the model specified in the role
		geminiAPIURL,
		geminiAPIKey,
		configurableTools,
	)

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
	roles []types.Role,
	initialInput map[string]interface{},
	geminiAPIURL string,
	geminiAPIKey string,
	configurableTools []types.ConfigurableTool,
	logFilePath string, // Add logFilePath parameter
) (map[string]interface{}, error) {
	context := make(map[string]interface{})
	for k, v := range initialInput {
		context[k] = v
	}

	for _, chainRole := range chain.Roles {
		// Find the role definition
		var currentRole types.Role
		found := false
		for _, r := range roles {
			if r.Name == chainRole.Name {
				currentRole = r
				found = true
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

		logrus.Infof("Executing role: %s with input: %v", currentRole.Name, roleInput)
		rawOutput, err := ExecuteRole(currentRole, roleInput, geminiAPIURL, geminiAPIKey, configurableTools, logFilePath)
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
			// Only handle write_file tool for now
			if toolCallObj.ToolName == "write_file" {
				filePath, _ := toolCallObj.Arguments["file_path"].(string)
				content, _ := toolCallObj.Arguments["content"].(string)
				logrus.Debugf("[ToolCall] About to write file: %s (len=%d)", filePath, len(content))
				if filePath != "" {
					_, _ = tools.WriteFile(filePath, content)
				} else {
					logrus.Warn("[ToolCall] file_path is empty, skipping file write")
				}
			}
		} else {
			// Try tool_call (Gemini style, use rawOutput)
			var toolCallWrap struct {
				ToolCall struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments"`
				} `json:"tool_call"`
			}
			if err := json.Unmarshal([]byte(rawOutput), &toolCallWrap); err == nil && toolCallWrap.ToolCall.Name == "write_file" {
				filePath, _ := toolCallWrap.ToolCall.Arguments["file_path"].(string)
				content, _ := toolCallWrap.ToolCall.Arguments["content"].(string)
				logrus.Debugf("[ToolCallWrap] About to write file: %s (len=%d)", filePath, len(content))
				if filePath != "" {
					logrus.Infof("[ToolCallWrap] Writing file: %s", filePath)
					_, _ = tools.WriteFile(filePath, content)
				} else {
					logrus.Warn("[ToolCallWrap] file_path is empty, skipping file write")
				}
			} else {
				// Fallback: if output is a JSON object with file_path and content, write the file (use extractFirstJSON output)
				var fileObj struct {
					FilePath string `json:"file_path"`
					Content  string `json:"content"`
				}
				if err := json.Unmarshal([]byte(output), &fileObj); err == nil {
					logrus.Debugf("[Fallback] fileObj: file_path=%s, content-len=%d", fileObj.FilePath, len(fileObj.Content))
					if fileObj.FilePath != "" {
						logrus.Infof("[Fallback] Writing file: %s", fileObj.FilePath)
						_, _ = tools.WriteFile(fileObj.FilePath, fileObj.Content)
					} else {
						logrus.Warn("[Fallback] file_path is empty, skipping file write")
					}
				}
			}
		}

		if chainRole.OutputKey != "" {
			context[chainRole.OutputKey] = output
			logrus.Infof("Role %s output stored in context key: %s", currentRole.Name, chainRole.OutputKey)
		}
	}

	return context, nil
}
