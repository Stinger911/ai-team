package ai

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"ai-team/pkg/tools"
	"ai-team/pkg/types"

	"github.com/sirupsen/logrus"
)

// YAMLHandler extracts tool-calls from YAML blocks (future extensibility).
type YAMLHandler struct{}

func (h *YAMLHandler) Name() string { return "yaml_block" }

func (h *YAMLHandler) Extract(s string) (*types.ToolCall, error) {
	// Example: look for ```yaml ... ``` blocks (not implemented, placeholder)
	return nil, fmt.Errorf("YAML handler not implemented")
}

// ToolCallExtractor provides robust extraction of tool-calls from AI responses.
type ToolCallExtractor struct {
	Handlers []ToolCallFormatHandler
	Registry *tools.ToolRegistry
}

// ToolCallFormatHandler attempts to extract a tool-call from a string.
type ToolCallFormatHandler interface {
	Extract(s string) (*types.ToolCall, error)
	Name() string
}

// JSONCodeBlockHandler extracts JSON from markdown code blocks.
type JSONCodeBlockHandler struct{}

func (h *JSONCodeBlockHandler) Name() string { return "json_code_block" }

func (h *JSONCodeBlockHandler) Extract(s string) (*types.ToolCall, error) {
	re := regexp.MustCompile("(?s)```json\\s*(\\{.*?\\})\\s*```")
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return nil, fmt.Errorf("no json code block found")
	}
	return parseToolCallJSON(matches[1])
}

// InlineJSONHandler extracts the first JSON object from text.
type InlineJSONHandler struct{}

func (h *InlineJSONHandler) Name() string { return "inline_json" }

func (h *InlineJSONHandler) Extract(s string) (*types.ToolCall, error) {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no inline json found")
	}
	jsonStr := s[start : end+1]
	return parseToolCallJSON(jsonStr)
}

// parseToolCallJSON tries to parse a tool-call from a JSON string.
func parseToolCallJSON(jsonStr string) (*types.ToolCall, error) {
	var req types.ToolCallRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err == nil && req.ToolCall.Name != "" {
		return &req.ToolCall, nil
	}
	// Try direct ToolCall
	var tc types.ToolCall
	if err := json.Unmarshal([]byte(jsonStr), &tc); err == nil && tc.Name != "" {
		return &tc, nil
	}
	return nil, fmt.Errorf("no valid tool_call in json")
}

// ExtractToolCall runs all handlers and returns the first valid tool-call.
func (e *ToolCallExtractor) ExtractToolCall(s string) (*types.ToolCall, string, error) {
	log := logrus.WithField("component", "ToolCallExtractor")
	for _, h := range e.Handlers {
		log.Debugf("Trying handler: %s", h.Name())
		tc, err := h.Extract(s)
		if err == nil && tc != nil {
			log.Infof("Handler '%s' succeeded: tool=%s", h.Name(), tc.Name)
			if e.Registry != nil && e.Registry.ValidateToolCall(tools.ToolCall{Name: tc.Name, Arguments: tc.Arguments}) != nil {
				log.Warnf("Schema validation failed for tool-call: %s", tc.Name)
				continue // schema validation failed
			}
			return tc, h.Name(), nil
		} else if err != nil {
			log.Debugf("Handler '%s' failed: %v", h.Name(), err)
		}
	}
	log.Warn("No valid tool-call found in response")
	return nil, "", fmt.Errorf("no valid tool-call found")
}

// NewDefaultToolCallExtractor returns a ToolCallExtractor with default handlers.
func NewDefaultToolCallExtractor(reg *tools.ToolRegistry) *ToolCallExtractor {
	return &ToolCallExtractor{
		Handlers: []ToolCallFormatHandler{
			&JSONCodeBlockHandler{},
			&InlineJSONHandler{},
			&YAMLHandler{}, // Pluggable for future
		},
		Registry: reg,
	}
}
