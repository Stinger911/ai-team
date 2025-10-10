package roles

import (
	"testing"
)

func TestEvaluateLoopCondition_BasicTrueFalse(t *testing.T) {
	ok, err := evaluateLoopCondition("true", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected true for literal 'true'")
	}

	ok, err = evaluateLoopCondition("false", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected false for literal 'false'")
	}
}

func TestEvaluateLoopCondition_EqualityAndInequality(t *testing.T) {
	ctx := map[string]interface{}{"foo": "bar"}
	ok, err := evaluateLoopCondition("{{.foo}} == 'bar'", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected equality to be true")
	}

	ok, err = evaluateLoopCondition("{{.foo}} != 'bar'", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected inequality to be false")
	}

	ctx2 := map[string]interface{}{"foo": "baz"}
	ok, err = evaluateLoopCondition("{{.foo}} != 'bar'", ctx2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected inequality to be true when values differ")
	}
}

func TestEvaluateLoopCondition_TemplateWithToolCall(t *testing.T) {
	ctx := map[string]interface{}{"tool_call": map[string]interface{}{"name": "write_file"}}
	ok, err := evaluateLoopCondition("{{.tool_call.name}} == 'write_file'", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected condition to be true for tool_call.name == write_file")
	}
}

func TestEvaluateLoopCondition_UnrecognizedAndInvalid(t *testing.T) {
	// Unrecognized expression should return false without error
	ok, err := evaluateLoopCondition("some random text", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error for unrecognized expression: %v", err)
	}
	if ok {
		t.Fatalf("expected unrecognized expression to evaluate to false")
	}

	// Invalid template should return an error
	_, err = evaluateLoopCondition("{{.foo", map[string]interface{}{})
	if err == nil {
		t.Fatalf("expected error for invalid template")
	}
}
