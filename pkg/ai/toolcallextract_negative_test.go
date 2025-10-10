package ai

import "testing"

func TestToolCallExtractor_MalformedJSONDoesNotPanic(t *testing.T) {
	extractor := NewDefaultToolCallExtractor(nil)
	resp := "Here is malformed JSON: {\"tool_call\": {\"name\": \"write_file\", \"arguments\": {\"file_path\": \"foo.txt\", \"content\": \"unterminated }}"
	tc, handler, err := extractor.ExtractToolCall(resp)
	if err == nil && tc != nil {
		t.Fatalf("expected no tool-call for malformed JSON, got tc=%v handler=%s", tc, handler)
	}
}

func TestToolCallExtractor_NoToolCallWhenNotPresent(t *testing.T) {
	extractor := NewDefaultToolCallExtractor(nil)
	resp := "Just some text and numbers: 123 { not a tool }"
	tc, handler, err := extractor.ExtractToolCall(resp)
	if err == nil || tc != nil {
		t.Fatalf("expected no tool-call, got err=%v, tc=%v, handler=%s", err, tc, handler)
	}
}
