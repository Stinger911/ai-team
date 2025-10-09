package roles

import (
	"ai-team/config"
	"ai-team/pkg/ai"
	"ai-team/pkg/types"
	"net/http"
	"testing"
)

func TestExecuteRole_Basic(t *testing.T) {
	// Mock ai.CallGeminiFunc to avoid real HTTP
	origCallGemini := ai.CallGeminiFunc
	ai.CallGeminiFunc = func(_ *http.Client, prompt, model, apiURL, apiKey string, tools []types.ConfigurableTool) (string, error) {
		return "mocked-response", nil
	}
	defer func() { ai.CallGeminiFunc = origCallGemini }()

	role := types.Role{
		Prompt: "You are a test role. Echo: {{.input}}",
		Model:  "gemini-2.5-flash",
	}
	input := map[string]interface{}{"input": "hello"}
	mockCfg := config.Config{}
	mockCfg.Gemini.Apiurl = "http://mock-gemini"
	mockCfg.Gemini.Models = map[string]config.ModelConfig{
		"gemini-2.5-flash": {
			Model: "gemini-2.5-flash",
		},
	}
	output, err := ExecuteRole(role, input, mockCfg, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty output")
	}
}

// Add more tests for ExecuteChain, tool call fallback, etc.
