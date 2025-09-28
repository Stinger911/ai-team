package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ai-team/pkg/types"
)

// LogRoleCall appends a role call log entry to a specified log file.
func LogRoleCall(logFilePath string, entry types.RoleCallLogEntry) error {
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logFilePath, err)
	}
	defer file.Close()

	entry.Timestamp = time.Now().Format(time.RFC3339)

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	_, err = file.Write(jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to write log entry to file: %w", err)
	}
	_, err = file.WriteString("\n") // Add a newline for readability
	if err != nil {
		return fmt.Errorf("failed to write newline to log file: %w", err)
	}

	return nil
}
