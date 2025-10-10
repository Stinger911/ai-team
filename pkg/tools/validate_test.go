package tools

import (
	"testing"
)

func TestValidateToolCall_ListDirSnakeCase(t *testing.T) {
	reg := NewToolRegistry()
	RegisterDefaultTools(reg)

	call := ToolCall{Name: "list_dir", Arguments: map[string]interface{}{"path": "."}}
	if err := reg.ValidateToolCall(call); err != nil {
		t.Fatalf("expected valid tool call, got error: %v", err)
	}
}

func TestValidateToolCall_WriteFileVariants(t *testing.T) {
	reg := NewToolRegistry()
	RegisterDefaultTools(reg)

	// snake_case registration
	call1 := ToolCall{Name: "write_file", Arguments: map[string]interface{}{"file_path": "a.txt", "content": "x"}}
	if err := reg.ValidateToolCall(call1); err != nil {
		t.Fatalf("expected valid write_file call, got error: %v", err)
	}

	// camelCase registration
	call2 := ToolCall{Name: "WriteFile", Arguments: map[string]interface{}{"filePath": "b.txt", "content": "y"}}
	if err := reg.ValidateToolCall(call2); err != nil {
		t.Fatalf("expected valid WriteFile call, got error: %v", err)
	}

	// missing required argument should fail
	call3 := ToolCall{Name: "write_file", Arguments: map[string]interface{}{"file_path": "c.txt"}}
	if err := reg.ValidateToolCall(call3); err == nil {
		t.Fatalf("expected error for missing content, got nil")
	}
}

func TestValidateToolCall_IntAcceptance(t *testing.T) {
	reg := NewToolRegistry()
	// register a test tool that requires an int argument
	reg.RegisterTool(ToolSchema{
		Name:        "TestInt",
		Description: "Test int arg",
		Arguments: []ToolArgument{
			{Name: "count", Type: "int", Required: true, Description: "count"},
		},
	}, &ListDirTool{})

	// JSON-unmarshaled numbers are float64; ensure float that is integer-valued is accepted
	call := ToolCall{Name: "TestInt", Arguments: map[string]interface{}{"count": 2.0}}
	if err := reg.ValidateToolCall(call); err != nil {
		t.Fatalf("expected valid int acceptance, got error: %v", err)
	}
}
