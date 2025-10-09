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

func TestListDir_Success(t *testing.T) {
	dir := t.TempDir()
	f1 := dir + "/file1.txt"
	f2 := dir + "/file2.txt"
	subdir := dir + "/subdir"
	os.WriteFile(f1, []byte("a"), 0644)
	os.WriteFile(f2, []byte("b"), 0644)
	os.Mkdir(subdir, 0755)
	files, err := ListDir(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	found := map[string]bool{"file1.txt": false, "file2.txt": false, "subdir/": false}
	for _, name := range files {
		if _, ok := found[name]; ok {
			found[name] = true
		}
	}
	for k, v := range found {
		if !v {
			t.Errorf("expected to find %s in dir listing", k)
		}
	}
}

func TestListDir_Fail(t *testing.T) {
	_, err := ListDir("/no/such/dir")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestReadFile_Success(t *testing.T) {
	file := t.TempDir() + "/test.txt"
	content := "hello test file"
	os.WriteFile(file, []byte(content), 0644)
	out, err := ReadFile(file)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out != content {
		t.Errorf("file content mismatch: got %q, want %q", out, content)
	}
}

func TestReadFile_Fail(t *testing.T) {
	_, err := ReadFile("/no/such/file.txt")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
