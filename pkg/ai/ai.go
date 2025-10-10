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
	// Construct a simple request body (keep it flexible -- callers can pass a provider-specific apiURL)
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

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to read openai response body", readErr)
	}

	// If non-200, try to surface an API error message
	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&apiErr); err == nil && apiErr.Error.Message != "" {
			return "", errors.New(errors.ErrCodeAPI, fmt.Sprintf("OpenAI API error: %s", apiErr.Error.Message), nil)
		}
		return "", errors.New(errors.ErrCodeAPI, fmt.Sprintf("OpenAI API returned status %d", resp.StatusCode), nil)
	}

	bodyString := string(bodyBytes)
	logger.DebugPrintf("Raw OpenAI response: %s", bodyString)
	return bodyString, nil
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

	bodyString := string(bodyBytes)
	logger.DebugPrintf("Raw Gemini response: %s\n", bodyString)

	// Do not extract or execute tool calls here; just return the raw model response
	return bodyString, nil
}

var (
	// For testing purposes, these can be overridden
	WriteFileFunc  = tools.WriteFile
	RunCommandFunc = tools.RunCommand
	ApplyPatchFunc = tools.ApplyPatch
)

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

	if resp.StatusCode != http.StatusOK {
		// Try to decode a possible structured error
		var apiErr struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&apiErr); err == nil && apiErr.Error != "" {
			return "", errors.New(errors.ErrCodeAPI, fmt.Sprintf("Ollama API error: %s", apiErr.Error), nil)
		}
		return "", errors.New(errors.ErrCodeAPI, fmt.Sprintf("Ollama API returned status %d", resp.StatusCode), nil)
	}

	return string(bodyBytes), nil
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
