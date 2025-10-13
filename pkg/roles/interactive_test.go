package roles

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// MockUI is a mock implementation of the UI interface.
type MockUI struct {
	ConfirmFunc      func(prompt string) (bool, error)
	PromptSelectFunc func(options []string) (string, error)
	OpenEditorFunc   func(content string) (string, error)
	PagerFunc        func(content string) error
	PrettyJSONFunc   func(obj interface{}) error
}

func (m *MockUI) Confirm(prompt string) (bool, error) {
	if m.ConfirmFunc != nil {
		return m.ConfirmFunc(prompt)
	}
	return false, nil
}

func (m *MockUI) PromptSelect(options []string) (string, error) {
	if m.PromptSelectFunc != nil {
		return m.PromptSelectFunc(options)
	}
	return "", nil
}

func (m *MockUI) OpenEditor(content string) (string, error) {
	if m.OpenEditorFunc != nil {
		return m.OpenEditorFunc(content)
	}
	return "", nil
}

func (m *MockUI) Pager(content string) error {
	if m.PagerFunc != nil {
		return m.PagerFunc(content)
	}
	return nil
}

func (m *MockUI) PrettyJSON(obj interface{}) error {
	if m.PrettyJSONFunc != nil {
		return m.PrettyJSONFunc(obj)
	}
	return nil
}

func TestStartSession_Abort(t *testing.T) {
	// Create a mock UI
	mockUI := &MockUI{
		ConfirmFunc: func(prompt string) (bool, error) {
			return false, nil
		},
	}

	// Create a session
	session := &Session{
		UI: mockUI,
	}

	// Capture stdout
	output := captureOutput(func() {
		StartSession(session)
	})

	// Check that the session was aborted
	expected := "Session aborted."
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, output)
	}
}

// captureOutput captures stdout and returns it as a string.
func captureOutput(f func()) string {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old // restoring the real stdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}