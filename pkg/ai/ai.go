package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ai-team/pkg/errors"
	"ai-team/pkg/types"
)

func CallOpenAI(client *http.Client, task string, apiURL string, apiKey string) (string, error) {
	fmt.Println("Calling OpenAI API...")

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

func CallGemini(client *http.Client, task string, apiURL string, apiKey string) (string, error) {
	fmt.Println("Calling Gemini API...")

	requestBody := strings.NewReader(`{
		"contents": [{
			"parts":[
				{"text": "` + task + `"}
			]
		}]
	}`)

	req, err := http.NewRequest("POST", apiURL, requestBody)
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

	var geminiResp types.GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to decode gemini response", err)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", errors.New(errors.ErrCodeAPI, "no response from gemini", nil)
}

func CallOllama(client *http.Client, task string, apiURL string) (string, error) {
	fmt.Println("Calling Ollama API...")

	requestBody := strings.NewReader(`{
		"model": "llama2",
		"prompt": "` + task + `"
	}`)

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

	var ollamaResp types.OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", errors.New(errors.ErrCodeAPI, "failed to decode ollama response", err)
	}

	return ollamaResp.Response, nil
}