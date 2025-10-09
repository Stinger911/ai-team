package ai

import (
	"strings"
	"testing"
)

func TestToolCallExtractor_JSONCodeBlock(t *testing.T) {
	extractor := NewDefaultToolCallExtractor(nil)
	resp := "Here is a tool call:\n```json\n{\"tool_call\": {\"name\": \"write_file\", \"arguments\": {\"file_path\": \"foo.txt\", \"content\": \"bar\"}}}\n```"
	tc, handler, err := extractor.ExtractToolCall(resp)
	if err != nil || tc == nil {
		t.Fatalf("expected tool-call, got err=%v, tc=%v", err, tc)
	}
	if tc.Name != "write_file" {
		t.Errorf("expected name write_file, got %s", tc.Name)
	}
	if handler != "json_code_block" {
		t.Errorf("expected handler json_code_block, got %s", handler)
	}
}

func TestToolCallExtractor_InlineJSON(t *testing.T) {
	extractor := NewDefaultToolCallExtractor(nil)
	resp := "Random text {\"name\": \"run_command\", \"arguments\": {\"command\": \"ls\"}} more text"
	tc, handler, err := extractor.ExtractToolCall(resp)
	if err != nil || tc == nil {
		t.Fatalf("expected tool-call, got err=%v, tc=%v", err, tc)
	}
	if tc.Name != "run_command" {
		t.Errorf("expected name run_command, got %s", tc.Name)
	}
	if handler != "inline_json" {
		t.Errorf("expected handler inline_json, got %s", handler)
	}
}

func TestToolCallExtractor_NoToolCall(t *testing.T) {
	extractor := NewDefaultToolCallExtractor(nil)
	resp := "No tool call here."
	tc, handler, err := extractor.ExtractToolCall(resp)
	if err == nil || tc != nil {
		t.Fatalf("expected no tool-call, got err=%v, tc=%v, handler=%s", err, tc, handler)
	}
}

func TestToolCallExtractor_ArgumentWithMarkdownBlock(t *testing.T) {
	extractor := NewDefaultToolCallExtractor(nil)
	resp := "Here is a tool call:\n```json\n{\"tool_call\": {\"name\": \"write_file\", \"arguments\": {\"file_path\": \"foo.py\", \"content\": \"Here is some code:\\n```python\\ndef foo():\\n    return 42\\n```\\n\"}}}\n```"
	tc, handler, err := extractor.ExtractToolCall(resp)
	if err != nil || tc == nil {
		t.Fatalf("expected tool-call, got err=%v, tc=%v", err, tc)
	}
	if tc.Name != "write_file" {
		t.Errorf("expected name write_file, got %s", tc.Name)
	}
	args := tc.Arguments
	content, ok := args["content"].(string)
	if !ok || content == "" {
		t.Errorf("expected content argument to be present and non-empty")
	}
	if !strings.Contains(content, "def foo()") || !strings.Contains(content, "```python") {
		t.Errorf("expected content to contain python code block, got: %s", content)
	}
	if handler != "json_code_block" {
		t.Errorf("expected handler json_code_block, got %s", handler)
	}
}
