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

type ToolCallExtractorInterface interface {
	ExtractToolCall(s string) (*types.ToolCall, string, error)
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

	// 1. Try to parse the whole response as JSON (object or array)
	var raw interface{}
	if err := json.Unmarshal([]byte(s), &raw); err == nil {
		// Recursively search for tool-call JSON in all string fields
		if _, found := findToolCallInJSON(raw); found != nil {
			log.Infof("Found tool-call in parsed JSON structure: tool=%s", found.Name)
			// Normalize payload before validation to canonical names (snake_case lower)
			norm := normalizeToolCall(found)
			log.Debugf("Original tool-call: name=%s args=%+v", found.Name, found.Arguments)
			log.Debugf("Normalized tool-call: name=%s args=%+v", norm.Name, norm.Arguments)
			if e.Registry != nil {
				if err := e.Registry.ValidateToolCall(tools.ToolCall{Name: norm.Name, Arguments: norm.Arguments}); err != nil {
					log.Warnf("Schema validation failed for tool-call: %s: %v", norm.Name, err)
				} else {
					// return original shape for downstream but indicate we used normalized name for validation
					return found, "json_recursive", nil
				}
			} else {
				return found, "json_recursive", nil
			}
		}
	}

	// 2. Try all format handlers (legacy)
	for _, h := range e.Handlers {
		log.Debugf("Trying handler: %s", h.Name())
		tc, err := h.Extract(s)
		if err == nil && tc != nil {
			log.Infof("Handler '%s' succeeded: tool=%s", h.Name(), tc.Name)
			// Normalize before validation
			norm := normalizeToolCall(tc)
			log.Debugf("Original tool-call (handler %s): name=%s args=%+v", h.Name(), tc.Name, tc.Arguments)
			log.Debugf("Normalized tool-call (handler %s): name=%s args=%+v", h.Name(), norm.Name, norm.Arguments)
			if e.Registry != nil {
				if errVal := e.Registry.ValidateToolCall(tools.ToolCall{Name: norm.Name, Arguments: norm.Arguments}); errVal != nil {
					log.Warnf("Schema validation failed for tool-call: %s: %v", norm.Name, errVal)
					continue // schema validation failed
				}
			}
			return tc, h.Name(), nil
		} else if err != nil {
			log.Debugf("Handler '%s' failed: %v", h.Name(), err)
		}
	}
	log.Warn("No valid tool-call found in response")
	return nil, "", fmt.Errorf("no valid tool-call found")
}

// findToolCallInJSON recursively searches for a tool-call JSON string in all string fields of a JSON object/array.
func findToolCallInJSON(v interface{}) (*types.ToolCall, *types.ToolCall) {
	switch val := v.(type) {
	case map[string]interface{}:
		for _, v2 := range val {
			// If string, try to parse as tool-call
			if s, ok := v2.(string); ok {
				if tc, err := parseToolCallJSON(s); err == nil && tc != nil {
					return tc, tc
				}
			} else {
				if tc, found := findToolCallInJSON(v2); found != nil {
					return tc, found
				}
			}
		}
	case []interface{}:
		for _, item := range val {
			if tc, found := findToolCallInJSON(item); found != nil {
				return tc, found
			}
		}
	}
	return nil, nil
}

// NewDefaultToolCallExtractor returns a ToolCallExtractor with default handlers.
func NewDefaultToolCallExtractor(reg *tools.ToolRegistry) ToolCallExtractorInterface {
	return &ToolCallExtractor{
		Handlers: []ToolCallFormatHandler{
			&JSONCodeBlockHandler{},
			&InlineJSONHandler{},
			&YAMLHandler{}, // Pluggable for future
		},
		Registry: reg,
	}
}

// normalizeToolCall produces a normalized copy of a ToolCall with the tool name
// converted to snake_case lowercase and argument keys converted to snake_case
// lowercase. This helps matching tool names/args against registry schemas which
// may be registered in snake_case or camelCase variants.
func normalizeToolCall(tc *types.ToolCall) *types.ToolCall {
	if tc == nil {
		return nil
	}
	norm := &types.ToolCall{
		Name:      toSnakeCaseLocal(tc.Name),
		Arguments: make(map[string]interface{}),
	}
	for k, v := range tc.Arguments {
		nk := toSnakeCaseLocal(k)
		// lower-case the key for extra tolerance
		nk = strings.ToLower(nk)
		// if key already set, prefer existing (do not overwrite)
		if _, exists := norm.Arguments[nk]; !exists {
			norm.Arguments[nk] = v
		}
	}
	return norm
}

// toSnakeCaseLocal converts CamelCase or camelCase or snake_case to snake_case.
func toSnakeCaseLocal(s string) string {
	var out []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, '_')
		}
		out = append(out, r)
	}
	return strings.ToLower(string(out))
}
