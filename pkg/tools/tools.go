package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"ai-team/pkg/errors"
)

// WriteFile writes content to a specified file.
func WriteFile(filePath string, content string) (string, error) {
	absPath, absErr := os.Getwd()
	if absErr == nil {
		fmt.Printf("[WriteFile] Current working directory: %s\n", absPath)
	}
	fmt.Printf("[WriteFile] Attempting to write file: %s\n", filePath)

	// Ensure parent directory exists
	dir := filePath
	if idx := strings.LastIndex(filePath, "/"); idx != -1 {
		dir = filePath[:idx]
		if dir != "" {
			if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
				fmt.Printf("[WriteFile] Failed to create parent directory %s: %v\n", dir, mkErr)
				return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("failed to create parent directory %s", dir), mkErr)
			}
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		fmt.Printf("[WriteFile] Failed to write file %s: %v\n", filePath, err)
		return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("failed to write file %s", filePath), err)
	}
	fmt.Printf("[WriteFile] Successfully wrote to file: %s\n", filePath)
	return fmt.Sprintf("Successfully wrote to file: %s", filePath), nil
}

// RunCommand executes a shell command.
func RunCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("failed to run command: %s", command), err)
	}
	return string(output), nil
}

// ApplyPatch applies a patch to a file.
func ApplyPatch(filePath string, patchContent string) (string, error) {
	// Create a temporary patch file
	tmpPatchFile, err := os.CreateTemp("", "patch-*.patch")
	if err != nil {
		return "", errors.New(errors.ErrCodeTool, "failed to create temporary patch file", err)
	}
	defer os.Remove(tmpPatchFile.Name()) // Clean up the temporary file

	_, err = tmpPatchFile.WriteString(patchContent)
	if err != nil {
		return "", errors.New(errors.ErrCodeTool, "failed to write patch content to temporary file", err)
	}
	tmpPatchFile.Close()

	// Apply the patch using the 'patch' command
	cmd := exec.Command("patch", filePath, tmpPatchFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("failed to apply patch to %s", filePath), err)
	}

	return fmt.Sprintf("Successfully applied patch to %s:\n%s", filePath, string(output)), nil
}
