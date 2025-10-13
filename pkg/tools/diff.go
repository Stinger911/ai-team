package tools

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
)

// GenerateUnifiedDiff returns a unified diff string between old and new content.
func GenerateUnifiedDiff(filePath, oldContent, newContent string) string {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")
	var diff bytes.Buffer
	diff.WriteString(fmt.Sprintf("--- %s\n+++ %s\n", filePath, filePath))
	for i := 0; i < len(oldLines) || i < len(newLines); i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}
		if oldLine != newLine {
			if oldLine != "" {
				diff.WriteString(fmt.Sprintf("-%s\n", oldLine))
			}
			if newLine != "" {
				diff.WriteString(fmt.Sprintf("+%s\n", newLine))
			}
		}
	}
	return diff.String()
}

// ReadFileOrEmpty returns the file content or empty string if not found.
func ReadFileOrEmpty(filePath string) string {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return string(b)
}
