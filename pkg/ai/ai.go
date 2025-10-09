package ai

import (
	"ai-team/pkg/errors"
	"ai-team/pkg/tools"
	"ai-team/pkg/types"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ai-team/pkg/logger"

	"github.com/sirupsen/logrus"
)

// AIClient abstracts provider-specific logic for chat and embedding.
type AIClient interface {
	ChatCompletion(task string) (string, error)
	// Add more methods as needed, e.g. Embedding, Image, etc.
}

// OpenAIClient implements AIClient for OpenAI.
type OpenAIClient struct {
	Client *http.Client
	APIURL string
	APIKey string
	Model  string
}

func (c *OpenAIClient) ChatCompletion(task string) (string, error) {
	return CallOpenAI(c.Client, task, c.APIURL, c.APIKey)
}

// GeminiClient implements AIClient for Gemini.
type GeminiClient struct {
	Client            *http.Client
	APIURL            string
	APIKey            string
	Model             string
	ConfigurableTools []types.ConfigurableTool
}

func (c *GeminiClient) ChatCompletion(task string) (string, error) {
	return CallGemini(c.Client, task, c.Model, c.APIURL, c.APIKey, c.ConfigurableTools)
}

// OllamaClient implements AIClient for Ollama.
type OllamaClient struct {
	Client            *http.Client
	APIURL            string
	Model             string
	ConfigurableTools []types.ConfigurableTool
}

func (c *OllamaClient) ChatCompletion(task string) (string, error) {
	return CallOllama(c.Client, task, c.APIURL, c.Model, c.ConfigurableTools)
}

// CallGeminiFunc allows mocking of CallGemini in tests
var CallGeminiFunc = CallGemini

func CallOpenAI(client *http.Client, task string, apiURL string, apiKey string) (string, error) {
	logrus.Info("Calling OpenAI API...")

	requestBody := strings.NewReader(`{
		"model": "text-davinci-003",
		"prompt": "` + task + `",
		"max_tokens": 100
	}`)

	req, err := http.NewRequest("POST", apiURL, requestBody)
	if err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to create openai request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to send openai request", err)
	}
	defer resp.Body.Close()

	var openAIResp types.OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to decode openai response", err)
	}

	if len(openAIResp.Choices) > 0 {
		return openAIResp.Choices[0].Text, nil
	}

	return "", errors.New(errors.ErrCodeAPI, "no response from openai", nil)
}

func CallGemini(client *http.Client, task string, model string, apiURL string, apiKey string, configurableTools []types.ConfigurableTool) (string, error) {
	logrus.Infof("Calling Gemini API with model: %s", model)

	// Construct the full API URL with the model
	fullAPIURL := fmt.Sprintf("%s/v1/models/%s:generateContent", apiURL, model)

	// Escape the task string for JSON
	request := types.GeminiRequest{
		Contents: []types.GeminiContent{
			{
				Parts: []types.GeminiPart{
					{Text: task},
				},
			},
		},
	}
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to marshal gemini request body", err)
	}
	requestBody := strings.NewReader(string(bodyBytes))
	logger.DebugPrintf("Gemini request body: %s", string(bodyBytes))

	req, err := http.NewRequest("POST", fullAPIURL, requestBody)
	if err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to create gemini request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.URL.RawQuery = "key=" + apiKey

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to send gemini request", err)
	}
	defer resp.Body.Close()

	// Read the response body once to allow for multiple decodes
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to read gemini response body", readErr)
	}

	// Check for API errors first (e.g., non-200 status code with error message)
	if resp.StatusCode != http.StatusOK {
		var apiError struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&apiError); err == nil && apiError.Error.Message != "" {
			return "", errors.New(errors.ErrCodeAPI, fmt.Sprintf("Gemini API error: %s", apiError.Error.Message), nil)
		}
		return "", errors.New(errors.ErrCodeAPI, fmt.Sprintf("Gemini API returned status %d", resp.StatusCode), nil)
	}

	// Attempt to decode as a tool call
	var toolCallReq types.ToolCallRequest
	bodyString := string(bodyBytes)

	// Debug: Log the raw response for troubleshooting
	logger.DebugPrintf("Raw Gemini response: %s\n", bodyString)

	// Try to extract JSON from a markdown code block with "tool_code" language specifier
	toolCodeBlockPrefix := "```tool_code\n"
	toolCodeBlockSuffix := "```"
	if strings.HasPrefix(bodyString, toolCodeBlockPrefix) && strings.HasSuffix(bodyString, toolCodeBlockSuffix) {
		jsonString := strings.TrimPrefix(bodyString, toolCodeBlockPrefix)
		jsonString = strings.TrimSuffix(jsonString, toolCodeBlockSuffix)
		bodyString = jsonString // Use the extracted JSON string for further processing
		logger.DebugPrintf("Extracted tool_code JSON: %s\n", bodyString)
	}

	// Now, try to extract content between <__AI_AGENT_CONTENT__> tags
	contentStartTag := "<__AI_AGENT_CONTENT__>"
	contentEndTag := "<__AI_AGENT_CONTENT__>"
	if strings.Contains(bodyString, contentStartTag) && strings.Contains(bodyString, contentEndTag) {
		startIndex := strings.Index(bodyString, contentStartTag) + len(contentStartTag)
		endIndex := strings.LastIndex(bodyString, contentEndTag)
		if startIndex < endIndex {
			extractedContent := bodyString[startIndex:endIndex]
			logger.DebugPrintf("Extracted content between tags: %s\n", extractedContent)
			// Now, try to unmarshal the extracted content as a tool call
			if err := json.NewDecoder(bytes.NewReader([]byte(extractedContent))).Decode(&toolCallReq); err == nil && toolCallReq.ToolCall.Name != "" {
				logger.DebugPrintf("Successfully parsed tool call from content tags: %s\n", toolCallReq.ToolCall.Name)
				// It's a tool call!
				toolOutput, toolErr := executeToolCall(toolCallReq.ToolCall, configurableTools)
				if toolErr != nil {
					return "", toolErr // Return the tool execution error
				}
				return toolOutput, nil // Return the tool's output
			} else {
				logger.DebugPrintf("Failed to parse tool call from content tags, error: %v\n", err)
			}
		}
	}

	// If not a tool call in a markdown block or between content tags, try direct JSON decode
	if err := json.NewDecoder(bytes.NewReader([]byte(bodyString))).Decode(&toolCallReq); err == nil && toolCallReq.ToolCall.Name != "" {
		logger.DebugPrintf("Successfully parsed tool call from direct JSON: %s\n", toolCallReq.ToolCall.Name)
		// It's a tool call!
		toolOutput, toolErr := executeToolCall(toolCallReq.ToolCall, configurableTools)
		if toolErr != nil {
			return "", toolErr // Return the tool execution error
		}
		return toolOutput, nil // Return the tool's output
	} else {
		logger.DebugPrintf("Failed to parse tool call from direct JSON, error: %v\n", err)
	}

	// If not a tool call, attempt to decode as a regular Gemini response
	var geminiResp types.GeminiResponse
	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&geminiResp); err == nil {
		if len(geminiResp.Candidates) > 0 {
			candidate := geminiResp.Candidates[0]
			// Handle UNEXPECTED_TOOL_CALL
			if candidate.FinishReason == "UNEXPECTED_TOOL_CALL" && candidate.ToolCall != nil {
				logger.DebugPrintf("Gemini returned UNEXPECTED_TOOL_CALL, executing tool: %s\n", candidate.ToolCall.Name)
				toolOutput, toolErr := executeToolCall(*candidate.ToolCall, configurableTools)
				if toolErr != nil {
					return "", toolErr
				}
				return toolOutput, nil
			}
			// Normal text response
			if len(candidate.Content.Parts) > 0 {
				return candidate.Content.Parts[0].Text, nil
			}
		}
		return "", errors.New(errors.ErrCodeAPI, "no content in Gemini response", nil)
	}

	return "", errors.New(errors.ErrCodeAPI, "malformed Gemini response or unrecognized format", nil)
}

var (
	// For testing purposes, these can be overridden
	WriteFileFunc  = tools.WriteFile
	RunCommandFunc = tools.RunCommand
	ApplyPatchFunc = tools.ApplyPatch
)

// executeToolCall executes the requested tool and returns its output.
func executeToolCall(tc types.ToolCall, configurableTools []types.ConfigurableTool) (string, error) {
	logger.DebugPrintf("Executing tool call: %s with args: %v\n", tc.Name, tc.Arguments)

	// Check for hardcoded tools first
	switch tc.Name {
	case "write_file":
		filePath, ok := tc.Arguments["file_path"].(string)
		if !ok {
			return "", errors.New(errors.ErrCodeTool, "missing or invalid 'file_path' for write_file", nil)
		}
		content, ok := tc.Arguments["content"].(string)
		if !ok {
			return "", errors.New(errors.ErrCodeTool, "missing or invalid 'content' for write_file", nil)
		}
		logger.DebugPrintf("Calling WriteFileFunc with filePath: %s, content length: %d\n", filePath, len(content)) // Debug print
		result, err := WriteFileFunc(filePath, content)
		if err != nil {
			logger.DebugPrintf("WriteFileFunc failed: %v\n", err)
			return "", err
		}
		logger.DebugPrintf("WriteFileFunc succeeded: %s\n", result)
		return result, nil
	case "run_command":
		command, ok := tc.Arguments["command"].(string)
		if !ok {
			return "", errors.New(errors.ErrCodeTool, "missing or invalid 'command' for run_command", nil)
		}
		logger.DebugPrintf("Calling RunCommandFunc with command: %s\n", command) // Debug print
		return RunCommandFunc(command)
	case "apply_patch":
		filePath, ok := tc.Arguments["file_path"].(string)
		if !ok {
			return "", errors.New(errors.ErrCodeTool, "missing or invalid 'file_path' for apply_patch", nil)
		}
		patchContent, ok := tc.Arguments["patch_content"].(string)
		if !ok {
			return "", errors.New(errors.ErrCodeTool, "missing or invalid 'patch_content' for apply_patch", nil)
		}
		logger.DebugPrintf("Calling ApplyPatchFunc with filePath: %s, patchContent length: %d\n", filePath, len(patchContent)) // Debug print
		return ApplyPatchFunc(filePath, patchContent)
	}

	// Check for configurable tools
	for _, ct := range configurableTools {
		if ct.Name == tc.Name {
			logger.DebugPrintf("Using configurable tool: %s\n", ct.Name)
			// Construct the command from the template
			command := ct.CommandTemplate
			for _, arg := range ct.Arguments {
				if val, ok := tc.Arguments[arg.Name].(string); ok {
					command = strings.ReplaceAll(command, fmt.Sprintf("{{.%s}}", arg.Name), val)
				} else {
					return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("missing or invalid argument '%s' for configurable tool '%s'", arg.Name, ct.Name), nil)
				}
			}
			logger.DebugPrintf("Executing configurable tool command: %s\n", command)
			result, err := RunCommandFunc(command)
			if err != nil {
				logger.DebugPrintf("Configurable tool failed: %v\n", err)
				return "", err
			}
			logger.DebugPrintf("Configurable tool succeeded: %s\n", result)
			return result, nil
		}
	}

	return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("unknown tool: %s", tc.Name), nil)
}

func CallOllama(client *http.Client, task string, apiURL string, model string, tools []types.ConfigurableTool) (string, error) {
	logrus.Info("Calling Ollama API...")
	var reqBody = types.OllamaRequest{
		Model: model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: task,
			},
		},
	}
	bodyStr, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to marshal ollama request body", err)
	}
	requestBody := strings.NewReader(string(bodyStr))
	logger.DebugPrintf("Ollama request body: %s", string(bodyStr))
	req, err := http.NewRequest("POST", apiURL, requestBody)
	if err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to create ollama request", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to send ollama request", err)
	}
	defer resp.Body.Close()

	logger.DebugPrintf("Ollama response status: %s", resp.Status)
	var bodyBytes, readErr = io.ReadAll(resp.Body)
	if readErr != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to read ollama response body", readErr)
	}
	logger.DebugPrintf("Ollama response body: %s", string(bodyBytes))
	var ollamaResp types.OllamaResponse
	if err := json.Unmarshal(bodyBytes, &ollamaResp); err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to decode ollama response", err)
	}

	return ollamaResp.Response, nil
}

// ListGeminiModels lists available Gemini models.
func ListGeminiModels(client *http.Client, apiURL string, apiKey string) ([]string, error) {
	logrus.Info("Listing Gemini models...")

	fullAPIURL := fmt.Sprintf("%s/v1/models", apiURL)

	req, err := http.NewRequest("GET", fullAPIURL, nil)
	if err != nil {
		return nil, errors.New(errors.ErrCodeAPI, "failed to create gemini list models request", err)
	}

	req.URL.RawQuery = "key=" + apiKey

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New(errors.ErrCodeAPI, "failed to send gemini list models request", err)
	}
	defer resp.Body.Close()

	var modelListResp types.GeminiModelListResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelListResp); err != nil {
		return nil, errors.New(errors.ErrCodeAPI, "failed to decode gemini list models response", err)
	}

	var models []string
	for _, model := range modelListResp.Models {
		models = append(models, model.Name)
	}

	return models, nil
}
