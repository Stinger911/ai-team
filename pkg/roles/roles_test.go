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
		Provider: "gemini",
		Prompt:   "You are a test role. Echo: {{.input}}",
		Model:    "gemini-2.5-flash",
	}
	input := map[string]interface{}{"input": "hello"}
	mockCfg := config.Config{}
	mockCfg.Gemini.Apiurl = "http://mock-gemini"
	mockCfg.Gemini.Models = map[string]config.ModelConfig{
		"gemini-2.5-flash": {
			Model: "gemini-2.5-flash",
		},
	}
	output, err := ExecuteRole(role, input, &mockCfg, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty output")
	}
}

// Add more tests for ExecuteChain, tool call fallback, etc.

func TestExecuteChain_AnalysisDesign_StopsOnWriteFile(t *testing.T) {
	// Mock ai.CallGeminiFunc to return list_dir responses for first two calls
	// and then a write_file tool call on the third call.
	origCallGemini := ai.CallGeminiFunc
	callCount := 0
	ai.CallGeminiFunc = func(_ *http.Client, prompt, model, apiURL, apiKey string, tools []types.ConfigurableTool) (string, error) {
		callCount++
		if callCount < 3 {
			// Return a JSON tool_call for list_dir
			return `{"tool_call": {"name": "list_dir", "arguments": {"path": "."}}}`, nil
		}
		// On third call, return write_file which should cause loop_condition to be true
		return `{"tool_call": {"name": "write_file", "arguments": {"file_path": "ai-team-data/pre-design.md", "content": "# Pre-design"}}}`, nil
	}
	defer func() { ai.CallGeminiFunc = origCallGemini }()

	// Prepare a minimal config matching the chain/roles used in work.nogit.yaml
	mockCfg := config.Config{}
	mockCfg.Gemini.Models = map[string]config.ModelConfig{"gemini-25-flash": {Model: "gemini-2.5-flash"}}
	mockCfg.Gemini.Apikey = "test"
	mockCfg.Gemini.Apiurl = "http://mock"

	// Add roles into config: analist (the looping role) and architect
	mockCfg.Roles = map[string]types.Role{
		"analist": {
			Provider: "gemini",
			Model:    "gemini-25-flash",
			Prompt:   "analist prompt",
		},
		"architect": {
			Provider: "gemini",
			Model:    "gemini-25-flash",
			Prompt:   "architect prompt",
		},
	}

	// Create the chain with analist looping and loop_condition matching write_file
	chain := types.RoleChain{
		Steps: []types.ChainRole{
			{
				Role:          "analist",
				Input:         map[string]interface{}{"problem": "test"},
				Loop:          true,
				LoopCount:     5,
				LoopCondition: "{{.tool_call.name}} == 'write_file'",
				OutputKey:     "pre_design",
			},
		},
	}

	ctx, err := ExecuteChain(chain, map[string]interface{}{"initial_problem": "x"}, &mockCfg, "")
	if err != nil {
		t.Fatalf("ExecuteChain returned error: %v", err)
	}
	// Ensure callCount is 3 (stopped when write_file was produced)
	if callCount != 3 {
		t.Fatalf("expected 3 calls to AI (2 list_dir then 1 write_file), got %d", callCount)
	}
	// Ensure pre_design is set in context
	if _, ok := ctx["pre_design"]; !ok {
		t.Fatalf("expected pre_design in context")
	}
}
