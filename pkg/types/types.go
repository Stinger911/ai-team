package types

// OpenAIResponse represents the JSON response from the OpenAI API.
type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// GeminiResponse represents the JSON response from the Gemini API.
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// OllamaResponse represents the JSON response from the Ollama API.
type OllamaResponse struct {
	Response string `json:"response"`
}
