package roles

import (
	"ai-team/pkg/ai"
	"ai-team/pkg/tools"
	"ai-team/pkg/types"
	"fmt"
)

// MockExecuteRoleFunc is a mock for the ExecuteRole function.
var MockExecuteRoleFunc func(role types.Role, inputs map[string]interface{}, cfg types.Config, logFilePath string) (string, error)

// MockExecuteRole is a wrapper around MockExecuteRoleFunc.
func MockExecuteRole(role types.Role, inputs map[string]interface{}, cfg types.Config, logFilePath string) (string, error) {
	if MockExecuteRoleFunc != nil {
		return MockExecuteRoleFunc(role, inputs, cfg, logFilePath)
	}
	return "", fmt.Errorf("ExecuteRole not mocked")
}

// MockToolCallExtractor is a mock for ai.ToolCallExtractor.
type MockToolCallExtractor struct {
	ExtractToolCallFunc func(llmOutput string) (*types.ToolCall, string, error)
}

func (m *MockToolCallExtractor) ExtractToolCall(llmOutput string) (*types.ToolCall, string, error) {
	if m.ExtractToolCallFunc != nil {
		return m.ExtractToolCallFunc(llmOutput)
	}
	return nil, llmOutput, fmt.Errorf("ExtractToolCall not mocked")
}

// NewDefaultToolCallExtractor is a mock for ai.NewDefaultToolCallExtractor.
var NewDefaultToolCallExtractorFunc func(registry *tools.ToolRegistry) ai.ToolCallExtractorInterface

func NewDefaultToolCallExtractor(registry *tools.ToolRegistry) ai.ToolCallExtractorInterface {
	if NewDefaultToolCallExtractorFunc != nil {
		return NewDefaultToolCallExtractorFunc(registry)
	}
	return &MockToolCallExtractor{}
}

// MockTool is a mock for the tools.Tool interface.
type MockTool struct {
	ExecuteFunc func(args map[string]interface{}) (interface{}, error)
}

func (m *MockTool) Execute(args map[string]interface{}) (interface{}, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(args)
	}
	return nil, fmt.Errorf("Tool Execute not mocked")
}

// MockToolRegistry is a mock for tools.ToolRegistry.
type MockToolRegistry struct {
	tools.ToolRegistry
	ValidateToolCallFunc func(call tools.ToolCall) error
	GetToolImplFunc      func(name string) (tools.Tool, bool)
}

func (m *MockToolRegistry) ValidateToolCall(call tools.ToolCall) error {
	if m.ValidateToolCallFunc != nil {
		return m.ValidateToolCallFunc(call)
	}
	return nil
}

func (m *MockToolRegistry) GetToolImpl(name string) (tools.Tool, bool) {
	if m.GetToolImplFunc != nil {
		return m.GetToolImplFunc(name)
	}
	return nil, false
}
