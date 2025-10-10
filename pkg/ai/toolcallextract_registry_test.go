package ai

import (
	"testing"

	"ai-team/pkg/tools"
)

func TestToolCallExtractor_UsesRegistryValidation(t *testing.T) {
	reg := tools.NewToolRegistry()
	tools.RegisterDefaultTools(reg)

	extractor := NewDefaultToolCallExtractor(reg)

	// model returns snake_case tool call inside JSON
	resp := "{\"tool_call\": {\"name\": \"write_file\", \"arguments\": {\"file_path\": \"test_out.txt\", \"content\": \"ok\"}}}"
	tc, handler, err := extractor.ExtractToolCall(resp)
	if err != nil || tc == nil {
		t.Fatalf("expected tool-call, got err=%v, tc=%v", err, tc)
	}
	if tc.Name != "write_file" {
		t.Fatalf("expected write_file, got %s", tc.Name)
	}
	if handler != "json_recursive" && handler != "inline_json" {
		t.Fatalf("expected handler json_recursive or inline_json, got %s", handler)
	}
}
