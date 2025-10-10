package ai

import (
	"ai-team/pkg/types"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"choices": [{"text": "Hello, world!"}]}`)
	}))
	defer server.Close()

	client := server.Client()

	resp, err := CallOpenAI(client, "write a hello world program in Go", server.URL, "test_api_key")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Parse raw response and verify the choices text
	var openResp types.OpenAIResponse
	if err := json.Unmarshal([]byte(resp), &openResp); err != nil {
		t.Errorf("failed to parse OpenAI raw response: %v", err)
	} else if len(openResp.Choices) == 0 || openResp.Choices[0].Text != "Hello, world!" {
		t.Errorf("expected choice text 'Hello, world!', got %+v", openResp)
	}
}

func TestCallGemini(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"candidates": [{"content": {"parts": [{"text": "Hello, world!"}]}}]}`)
	}))
	defer server.Close()

	client := server.Client()

	resp, err := CallGemini(client, "write a hello world program in Go", "gemini-pro", server.URL, "test_api_key", []types.ConfigurableTool{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Parse the raw JSON response and check the text field
	var gemResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(resp), &gemResp); err != nil {
		t.Errorf("failed to parse Gemini response: %v", err)
	} else if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		t.Errorf("missing candidates or parts in Gemini response")
	} else if gemResp.Candidates[0].Content.Parts[0].Text != "Hello, world!" {
		t.Errorf("expected text 'Hello, world!', got %q", gemResp.Candidates[0].Content.Parts[0].Text)
	}
}

func TestCallGemini_ToolCall(t *testing.T) {
	expectedFilePath := "test_file.txt"
	expectedContent := "tool call content"
	expectedToolOutput := "file written successfully"

	// Mock the WriteFileFunc
	originalWriteFileFunc := WriteFileFunc
	WriteFileFunc = func(filePath, content string) (string, error) {
		if filePath != expectedFilePath {
			t.Errorf("expected file path %q, got %q", expectedFilePath, filePath)
		}
		if content != expectedContent {
			t.Errorf("expected content %q, got %q", expectedContent, content)
		}
		return expectedToolOutput, nil
	}
	defer func() { WriteFileFunc = originalWriteFileFunc }() // Restore original

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"tool_call": {"name": "write_file", "arguments": {"file_path": %q, "content": %q}}}`, expectedFilePath, expectedContent)
	}))
	defer server.Close()

	client := server.Client()

	resp, err := CallGemini(client, "write a file", "gemini-pro", server.URL, "test_api_key", []types.ConfigurableTool{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Parse the raw JSON response and check the tool_call fields
	var toolCallResp struct {
		ToolCall struct {
			Name      string `json:"name"`
			Arguments struct {
				FilePath string `json:"file_path"`
				Content  string `json:"content"`
			} `json:"arguments"`
		} `json:"tool_call"`
	}
	if err := json.Unmarshal([]byte(resp), &toolCallResp); err != nil {
		t.Errorf("failed to parse tool_call response: %v", err)
	} else {
		if toolCallResp.ToolCall.Name != "write_file" {
			t.Errorf("expected tool name 'write_file', got %q", toolCallResp.ToolCall.Name)
		}
		if toolCallResp.ToolCall.Arguments.FilePath != expectedFilePath {
			t.Errorf("expected file_path %q, got %q", expectedFilePath, toolCallResp.ToolCall.Arguments.FilePath)
		}
		if toolCallResp.ToolCall.Arguments.Content != expectedContent {
			t.Errorf("expected content %q, got %q", expectedContent, toolCallResp.ToolCall.Arguments.Content)
		}
	}
}

func TestCallGemini_ModelSelection(t *testing.T) {
	expectedModel := "gemini-ultra"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, expectedModel) {
			t.Errorf("expected model %q in URL path, got %q", expectedModel, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"candidates": [{"content": {"parts": [{"text": "Model test response"}]}}]}`)
	}))
	defer server.Close()

	client := server.Client()

	_, err := CallGemini(client, "test task", expectedModel, server.URL, "test_api_key", []types.ConfigurableTool{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestListGeminiModels(t *testing.T) {
	expectedModels := []string{"gemini-pro", "gemini-ultra", "gemini-flash"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"models": [{"name": "gemini-pro"}, {"name": "gemini-ultra"}, {"name": "gemini-flash"}]}`)
	}))
	defer server.Close()

	client := server.Client()

	models, err := ListGeminiModels(client, server.URL, "test_api_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(models) != len(expectedModels) {
		t.Errorf("expected %d models, got %d", len(expectedModels), len(models))
	}

	for i, model := range models {
		if model != expectedModels[i] {
			t.Errorf("expected model %q at index %d, got %q", expectedModels[i], i, model)
		}
	}
}

func TestCallGemini_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error": {"message": "Bad Request"}}`)
	}))
	defer server.Close()

	client := server.Client()

	_, err := CallGemini(client, "test task", "gemini-pro", server.URL, "test_api_key", []types.ConfigurableTool{})
	if err == nil {
		t.Error("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "Gemini API error: Bad Request") {
		t.Errorf("expected API error message, got %v", err)
	}
}

func TestCallGemini_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `this is not json`)
	}))
	defer server.Close()

	client := server.Client()

	resp, err := CallGemini(client, "test task", "gemini-pro", server.URL, "test_api_key", []types.ConfigurableTool{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !strings.Contains(resp, "this is not json") {
		t.Errorf("expected raw response to contain 'this is not json', got %q", resp)
	}
}

func TestCallGemini_NetworkError(t *testing.T) {
	// Close the server immediately to simulate a network error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	client := server.Client()

	_, err := CallGemini(client, "test task", "gemini-pro", server.URL, "test_api_key", []types.ConfigurableTool{})
	if err == nil {
		t.Error("expected a network error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to send gemini request") {
		t.Errorf("expected network error, got %v", err)
	}
}

func TestCallOllama(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"response": "Hello, world!"}`)
	}))
	defer server.Close()

	client := server.Client()

	resp, err := CallOllama(client, "write a hello world program in Go", server.URL, "test-model", nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Parse raw response body and check for the 'response' field
	var or struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal([]byte(resp), &or); err != nil {
		t.Errorf("failed to parse Ollama raw response: %v", err)
	} else if or.Response != "Hello, world!" {
		t.Errorf("expected response 'Hello, world!', got %q", or.Response)
	}
}
