package roles

import (
	"os"
	"os/exec"
	"testing"
)

func TestRoleCommand_CLI(t *testing.T) {
	if _, err := os.Stat("../ai-team"); os.IsNotExist(err) {
		t.Skip("ai-team binary not found; skipping integration test")
	}
	cmd := exec.Command("../ai-team", "role", "architect", "--input", "problem=add two numbers")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("role command failed: %v\nOutput: %s", err, string(output))
	}
	if len(output) == 0 {
		t.Error("expected output from role command")
	}
}

func TestRunChainCommand_CLI(t *testing.T) {
	if _, err := os.Stat("../ai-team"); os.IsNotExist(err) {
		t.Skip("ai-team binary not found; skipping integration test")
	}
	cmd := exec.Command("../ai-team", "run-chain", "design-code-test", "--input", "initial_problem=add two numbers")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run-chain command failed: %v\nOutput: %s", err, string(output))
	}
	if len(output) == 0 {
		t.Error("expected output from run-chain command")
	}
}
