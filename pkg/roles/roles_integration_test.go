package roles

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"ai-team/pkg/ai"
	"net/http"

	"ai-team/pkg/types"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)

	// Build the ai-team binary before running tests
	projectRoot := getProjectRoot()
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(projectRoot, "ai-team"), filepath.Join(projectRoot, "main.go"))
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		logrus.Fatalf("Failed to build ai-team binary: %v", err)
	}

	// Explicitly set config file for viper
	configPath := filepath.Join(projectRoot, "config.yaml")
	viper.SetConfigFile(configPath)

	code := m.Run()
	os.Exit(code)
}

// Helper to simulate user input
func simulateInput(input string) *bytes.Buffer {
	return bytes.NewBufferString(input)
}

func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)                // This is pkg/roles
	return filepath.Join(basepath, "..", "..") // This should be the project root
}

func TestRoleCommand_CLI(t *testing.T) {
	if _, err := os.Stat(filepath.Join(getProjectRoot(), "ai-team")); os.IsNotExist(err) {
		t.Skip("ai-team binary not found; skipping integration test")
	}

	// Mock ai.CallGeminiFunc
	oldCallGeminiFunc := ai.CallGeminiFunc
	ai.CallGeminiFunc = func(client *http.Client, prompt, model, apiURL, apiKey string, tools []types.ConfigurableTool) (string, error) {
		return "Mocked Gemini Response", nil
	}
	defer func() {
		ai.CallGeminiFunc = oldCallGeminiFunc
	}()

	projectRoot := getProjectRoot()
	configPath := filepath.Join(projectRoot, "config.yaml")
	t.Logf("Config path: %s", configPath)
	cmd := exec.Command(filepath.Join(projectRoot, "ai-team"), "role", "architect", "--config", configPath, "problem=add two numbers")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "AI_TEAM_DEBUG=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("role command failed: %v\nOutput: %s", err, string(output))
	}
	if len(output) == 0 {
		t.Error("expected output from role command")
	}
}

func TestRunChainCommand_CLI(t *testing.T) {
	if _, err := os.Stat("../../ai-team"); os.IsNotExist(err) {
		t.Skip("ai-team binary not found; skipping integration test")
	}
	projectRoot := getProjectRoot()
	configPath := filepath.Join(projectRoot, "config.yaml")
	cmd := exec.Command(filepath.Join(projectRoot, "ai-team"), "run-chain", "design-code-test", "--config", configPath)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run-chain command failed: %v\nOutput: %s", err, string(output))
	}
	if len(output) == 0 {
		t.Error("expected output from run-chain command")
	}
}

func TestRoleCommand_InteractiveCLI_Abort(t *testing.T) {
	if _, err := os.Stat("../../ai-team"); os.IsNotExist(err) {
		t.Skip("ai-team binary not found; skipping integration test")
	}

	// Simulate "n" for "Start session?"
	input := simulateInput("n\n")

	projectRoot := getProjectRoot()
	configPath := filepath.Join(projectRoot, "config.yaml")
	cmd := exec.Command(filepath.Join(projectRoot, "ai-team"), "role", "--interactive", "--config", configPath)
	cmd.Dir = projectRoot
	cmd.Stdin = input
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		t.Fatalf("interactive role command failed: %v\nOutput: %s", err, out.String())
	}

	expectedOutput := "Session aborted."
	if !strings.Contains(out.String(), expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, out.String())
	}
}
