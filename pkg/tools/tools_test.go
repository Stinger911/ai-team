package tools

import (
	"os"
	"testing"
)

func TestWriteFile_Success(t *testing.T) {
	filePath := "test_writefile.txt"
	content := "hello, world!"
	defer os.Remove(filePath)
	msg, err := WriteFile(filePath, content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if msg == "" {
		t.Error("expected success message, got empty string")
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content mismatch: got %q, want %q", string(data), content)
	}
}

func TestWriteFile_Fail(t *testing.T) {
	// Try to write to a directory that doesn't exist and can't be created
	filePath := "/root/should_fail.txt"
	_, err := WriteFile(filePath, "fail")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestRunCommand_Success(t *testing.T) {
	out, err := RunCommand("echo hi")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out == "" {
		t.Error("expected output, got empty string")
	}
}

func TestRunCommand_Fail(t *testing.T) {
	_, err := RunCommand("nonexistentcommand1234")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestApplyPatch_Fail(t *testing.T) {
	_, err := ApplyPatch("/no/such/file.txt", "bad patch")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
