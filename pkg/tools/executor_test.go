package tools

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

type mockTool struct {
	attempts int32
}

func (m *mockTool) Execute(args map[string]interface{}) (interface{}, error) {
	a := atomic.AddInt32(&m.attempts, 1)
	if a < 2 {
		return nil, fmt.Errorf("transient error attempt=%d", a)
	}
	return "success", nil
}

type slowTool struct{}

func (s *slowTool) Execute(args map[string]interface{}) (interface{}, error) {
	// block longer than executor timeout to trigger timeout branch
	time.Sleep(200 * time.Millisecond)
	return "done", nil
}

func TestToolExecutor_RetryAndSuccess(t *testing.T) {
	reg := NewToolRegistry()
	// register a no-arg schema so validation passes
	reg.RegisterTool(ToolSchema{Name: "MockTool", Description: "mock"}, &mockTool{})
	// also register implementation under the same name
	reg.impls["MockTool"] = &mockTool{}

	exec := &ToolExecutor{Registry: reg, RetryCount: 3, Timeout: 1 * time.Second}
	res, err := exec.Execute(ToolCall{Name: "MockTool", Arguments: map[string]interface{}{}})
	if err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}
	if res != "success" {
		t.Fatalf("unexpected result: %v", res)
	}
}

func TestToolExecutor_Timeout(t *testing.T) {
	reg := NewToolRegistry()
	reg.RegisterTool(ToolSchema{Name: "SlowTool", Description: "slow"}, &slowTool{})
	reg.impls["SlowTool"] = &slowTool{}

	exec := &ToolExecutor{Registry: reg, RetryCount: 1, Timeout: 50 * time.Millisecond}
	_, err := exec.Execute(ToolCall{Name: "SlowTool", Arguments: map[string]interface{}{}})
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if err.Error() == "" {
		t.Fatalf("expected non-empty error message on timeout")
	}
}
