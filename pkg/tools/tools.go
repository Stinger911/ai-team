package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"ai-team/pkg/errors"
)

// ToolExecutor executes ToolCalls using a ToolRegistry.
type ToolExecutor struct {
	Registry *ToolRegistry
	Logger   *logrus.Logger
	// MetricsHook can be set to send metrics/events (stub for future integration)
	MetricsHook func(event string, fields map[string]interface{})
	RetryCount  int
	Timeout     time.Duration
}

// Execute runs a ToolCall with validation, logging, error handling, and retry/timeout logic.
func (te *ToolExecutor) Execute(call ToolCall) (interface{}, error) {
	if te.Logger == nil {
		te.Logger = logrus.New()
	}
	logger := te.Logger.WithFields(logrus.Fields{"tool": call.Name, "args": call.Arguments})
	logger.Infof("ToolExecutor: Executing tool call: %s", call.Name)
	if te.MetricsHook != nil {
		te.MetricsHook("tool_call_start", map[string]interface{}{"tool": call.Name, "args": call.Arguments})
	}

	// Validate tool call
	if err := te.Registry.ValidateToolCall(call); err != nil {
		logger.Errorf("Validation failed: %v", err)
		if te.MetricsHook != nil {
			te.MetricsHook("tool_call_validation_failed", map[string]interface{}{"tool": call.Name, "error": err.Error()})
		}
		return nil, err
	}

	toolImpl, ok := te.Registry.GetToolImpl(call.Name)
	if !ok {
		err := fmt.Errorf("tool implementation not found: %s", call.Name)
		logger.Error(err)
		if te.MetricsHook != nil {
			te.MetricsHook("tool_call_impl_not_found", map[string]interface{}{"tool": call.Name})
		}
		return nil, err
	}

	var lastErr error
	retries := te.RetryCount
	if retries < 1 {
		retries = 1
	}
	for attempt := 1; attempt <= retries; attempt++ {
		if te.MetricsHook != nil {
			te.MetricsHook("tool_call_attempt", map[string]interface{}{"tool": call.Name, "attempt": attempt})
		}
		ctx := context.Background()
		if te.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, te.Timeout)
			defer cancel()
		}
		done := make(chan struct{})
		var result interface{}
		go func() {
			result, lastErr = toolImpl.Execute(call.Arguments)
			close(done)
		}()
		select {
		case <-done:
			if lastErr == nil {
				logger.Infof("Tool %s succeeded on attempt %d", call.Name, attempt)
				if te.MetricsHook != nil {
					te.MetricsHook("tool_call_success", map[string]interface{}{"tool": call.Name, "attempt": attempt})
				}
				return result, nil
			}
			logger.Warnf("Tool %s failed on attempt %d: %v", call.Name, attempt, lastErr)
			if te.MetricsHook != nil {
				te.MetricsHook("tool_call_failure", map[string]interface{}{"tool": call.Name, "attempt": attempt, "error": lastErr.Error()})
			}
		case <-ctx.Done():
			lastErr = fmt.Errorf("tool %s timed out after %s", call.Name, te.Timeout)
			logger.Error(lastErr)
			if te.MetricsHook != nil {
				te.MetricsHook("tool_call_timeout", map[string]interface{}{"tool": call.Name, "timeout": te.Timeout.String()})
			}
		}
	}
	logger.Errorf("Tool %s failed after %d attempts: %v", call.Name, retries, lastErr)
	if te.MetricsHook != nil {
		te.MetricsHook("tool_call_final_failure", map[string]interface{}{"tool": call.Name, "retries": retries, "error": lastErr.Error()})
	}
	return nil, lastErr
}

// ToolRegistry holds all registered tools and their schemas.
type ToolRegistry struct {
	tools map[string]ToolSchema
	impls map[string]Tool // tool name to implementation
}

// NewToolRegistry creates a new ToolRegistry instance.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolSchema),
		impls: make(map[string]Tool),
	}
}

// RegisterTool registers a tool schema and its implementation.
func (r *ToolRegistry) RegisterTool(schema ToolSchema, impl Tool) {
	r.tools[schema.Name] = schema
	r.impls[schema.Name] = impl
}

// GetToolSchema returns the schema for a tool by name.
func (r *ToolRegistry) GetToolSchema(name string) (ToolSchema, bool) {
	schema, ok := r.tools[name]
	return schema, ok
}

// GetToolImpl returns the implementation for a tool by name.
func (r *ToolRegistry) GetToolImpl(name string) (Tool, bool) {
	impl, ok := r.impls[name]
	return impl, ok
}

// ListTools returns all registered tool schemas.
func (r *ToolRegistry) ListTools() []ToolSchema {
	schemas := make([]ToolSchema, 0, len(r.tools))
	for _, s := range r.tools {
		schemas = append(schemas, s)
	}
	return schemas
}

// Tool is the interface all tools must implement.
type Tool interface {
	Execute(args map[string]interface{}) (interface{}, error)
}

// WriteFileTool implements the Tool interface for writing files.
type WriteFileTool struct{}

func (t *WriteFileTool) Execute(args map[string]interface{}) (interface{}, error) {
	filePath, ok1 := args["filePath"].(string)
	content, ok2 := args["content"].(string)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("invalid arguments for WriteFile: filePath and content required")
	}
	return WriteFile(filePath, content)
}

// RunCommandTool implements the Tool interface for running shell commands.
type RunCommandTool struct{}

func (t *RunCommandTool) Execute(args map[string]interface{}) (interface{}, error) {
	command, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid arguments for RunCommand: command required")
	}
	return RunCommand(command)
}

// ApplyPatchTool implements the Tool interface for applying patches.
type ApplyPatchTool struct{}

func (t *ApplyPatchTool) Execute(args map[string]interface{}) (interface{}, error) {
	filePath, ok1 := args["filePath"].(string)
	patchContent, ok2 := args["patchContent"].(string)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("invalid arguments for ApplyPatch: filePath and patchContent required")
	}
	return ApplyPatch(filePath, patchContent)
}

// RegisterDefaultTools registers the built-in tools in the given registry.
func RegisterDefaultTools(reg *ToolRegistry) {
	reg.RegisterTool(ToolSchema{
		Name:        "WriteFile",
		Description: "Writes content to a specified file.",
		Arguments: []ToolArgument{
			{Name: "filePath", Type: "string", Required: true, Description: "Path to the file to write."},
			{Name: "content", Type: "string", Required: true, Description: "Content to write."},
		},
	}, &WriteFileTool{})

	reg.RegisterTool(ToolSchema{
		Name:        "RunCommand",
		Description: "Executes a shell command.",
		Arguments: []ToolArgument{
			{Name: "command", Type: "string", Required: true, Description: "Shell command to execute."},
		},
	}, &RunCommandTool{})

	reg.RegisterTool(ToolSchema{
		Name:        "ApplyPatch",
		Description: "Applies a patch to a file.",
		Arguments: []ToolArgument{
			{Name: "filePath", Type: "string", Required: true, Description: "Path to the file to patch."},
			{Name: "patchContent", Type: "string", Required: true, Description: "Patch content."},
		},
	}, &ApplyPatchTool{})
}

// ToolCall represents a validated tool invocation.
type ToolCall struct {
	Name      string
	Arguments map[string]interface{}
	Context   map[string]interface{} // optional
}

// ValidateToolCall checks if the ToolCall matches the ToolSchema in the registry.
func (r *ToolRegistry) ValidateToolCall(call ToolCall) error {
	schema, ok := r.GetToolSchema(call.Name)
	if !ok {
		return fmt.Errorf("tool '%s' not found in registry", call.Name)
	}
	// Check required arguments and types
	for _, arg := range schema.Arguments {
		val, exists := call.Arguments[arg.Name]
		if arg.Required && !exists {
			return fmt.Errorf("missing required argument '%s' for tool '%s'", arg.Name, call.Name)
		}
		if exists {
			switch arg.Type {
			case "string":
				if _, ok := val.(string); !ok {
					return fmt.Errorf("argument '%s' for tool '%s' must be string", arg.Name, call.Name)
				}
			case "int":
				if _, ok := val.(int); !ok {
					return fmt.Errorf("argument '%s' for tool '%s' must be int", arg.Name, call.Name)
				}
			case "bool":
				if _, ok := val.(bool); !ok {
					return fmt.Errorf("argument '%s' for tool '%s' must be bool", arg.Name, call.Name)
				}
				// Add more types as needed
			}
		}
	}
	return nil
}

// ToolSchema defines the schema for a tool, including its name, description, and arguments.
type ToolSchema struct {
	Name        string
	Description string
	Arguments   []ToolArgument
}

// ToolArgument defines a single argument for a tool.
type ToolArgument struct {
	Name        string
	Type        string // e.g. "string", "int", "bool"
	Required    bool
	Description string
}

// WriteFile writes content to a specified file.
func WriteFile(filePath string, content string) (string, error) {
	log := logrus.WithFields(logrus.Fields{
		"tool":        "WriteFile",
		"filePath":    filePath,
		"content_len": len(content),
	})
	log.Infof("Starting WriteFile with filePath=%s, content_len=%d", filePath, len(content))

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic during WriteFile: %v", r)
		}
	}()

	absPath, absErr := os.Getwd()
	if absErr == nil {
		log.Debugf("[WriteFile] Current working directory: %s", absPath)
	} else {
		log.Warnf("[WriteFile] Could not get current working directory: %v", absErr)
	}

	// Ensure parent directory exists
	dir := filePath
	if idx := strings.LastIndex(filePath, "/"); idx != -1 {
		dir = filePath[:idx]
		if dir != "" {
			log.Debugf("[WriteFile] Ensuring parent directory exists: %s", dir)
			if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
				log.Errorf("Failed to create parent directory %s: %v", dir, mkErr)
				return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("failed to create parent directory %s", dir), mkErr)
			}
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Errorf("Failed to write file %s: %v", filePath, err)
		return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("failed to write file %s (cwd=%s)", filePath, absPath), err)
	}
	log.Infof("Successfully wrote to file: %s %d bytes", filePath, len(content))
	log.Infof("Finished WriteFile")
	return fmt.Sprintf("Successfully wrote to file: %s", filePath), nil
}

// RunCommand executes a shell command.
func RunCommand(command string) (string, error) {
	log := logrus.WithFields(logrus.Fields{
		"tool":    "RunCommand",
		"command": command,
	})
	log.Infof("Starting RunCommand: %s", command)
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic during RunCommand: %v", r)
		}
	}()

	absPath, absErr := os.Getwd()
	if absErr == nil {
		log.Debugf("[RunCommand] Current working directory: %s", absPath)
	} else {
		log.Warnf("[RunCommand] Could not get current working directory: %v", absErr)
	}

	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to run command: %s, output: %s, err: %v", command, string(output), err)
		return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("failed to run command: %s (cwd=%s)", command, absPath), err)
	}
	log.Infof("Finished RunCommand: %s", command)
	return string(output), nil
}

// ApplyPatch applies a patch to a file.
func ApplyPatch(filePath string, patchContent string) (string, error) {
	log := logrus.WithFields(logrus.Fields{
		"tool":      "ApplyPatch",
		"filePath":  filePath,
		"patch_len": len(patchContent),
	})
	log.Infof("Starting ApplyPatch with filePath=%s, patch_len=%d", filePath, len(patchContent))
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic during ApplyPatch: %v", r)
		}
	}()
	absPath, absErr := os.Getwd()
	if absErr == nil {
		log.Debugf("[ApplyPatch] Current working directory: %s", absPath)
	} else {
		log.Warnf("[ApplyPatch] Could not get current working directory: %v", absErr)
	}
	// Create a temporary patch file
	tmpPatchFile, err := os.CreateTemp("", "patch-*.patch")
	if err != nil {
		log.Errorf("Failed to create temporary patch file: %v", err)
		return "", errors.New(errors.ErrCodeTool, "failed to create temporary patch file", err)
	}
	defer os.Remove(tmpPatchFile.Name()) // Clean up the temporary file

	_, err = tmpPatchFile.WriteString(patchContent)
	if err != nil {
		log.Errorf("Failed to write patch content to temporary file: %v", err)
		return "", errors.New(errors.ErrCodeTool, "failed to write patch content to temporary file", err)
	}
	tmpPatchFile.Close()

	// Apply the patch using the 'patch' command
	cmd := exec.Command("patch", filePath, tmpPatchFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("Failed to apply patch to %s, output: %s, err: %v", filePath, string(output), err)
		return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("failed to apply patch to %s (cwd=%s)", filePath, absPath), err)
	}

	log.Infof("Successfully applied patch to %s:\n%s", filePath, string(output))
	log.Infof("Finished ApplyPatch")
	return fmt.Sprintf("Successfully applied patch to %s:\n%s", filePath, string(output)), nil
}

// ExecuteTool executes the specified tool with the given parameters.
func ExecuteTool(toolName string, params map[string]interface{}) (string, error) {
	start := time.Now()
	log := logrus.WithFields(logrus.Fields{
		"tool":   toolName,
		"params": params,
	})
	log.Infof("Starting ExecuteTool: %s with params: %+v", toolName, params)
	absPath, absErr := os.Getwd()
	if absErr == nil {
		log.Debugf("[ExecuteTool] Current working directory: %s", absPath)
	} else {
		log.Warnf("[ExecuteTool] Could not get current working directory: %v", absErr)
	}

	var result string
	var err error

	switch toolName {
	case "WriteFile":
		filePath, ok1 := params["filePath"].(string)
		content, ok2 := params["content"].(string)
		if !ok1 || !ok2 {
			log.Errorf("Invalid parameters for WriteFile: filePath=%v, content=%v, types: %T, %T", ok1, ok2, params["filePath"], params["content"])
			err = errors.New(errors.ErrCodeTool, "Invalid parameters for WriteFile", nil)
			return "", err
		}
		log.Debugf("[ExecuteTool] WriteFile called with filePath=%s, content_len=%d", filePath, len(content))
		result, err = WriteFile(filePath, content)
	case "RunCommand":
		command, ok := params["command"].(string)
		if !ok {
			log.Errorf("Invalid parameters for RunCommand: command=%v, type: %T", ok, params["command"])
			err = errors.New(errors.ErrCodeTool, "Invalid parameters for RunCommand", nil)
			return "", err
		}
		log.Debugf("[ExecuteTool] RunCommand called with command=%s", command)
		result, err = RunCommand(command)
	case "ApplyPatch":
		filePath, ok1 := params["filePath"].(string)
		patchContent, ok2 := params["patchContent"].(string)
		if !ok1 || !ok2 {
			log.Errorf("Invalid parameters for ApplyPatch: filePath=%v, patchContent=%v, types: %T, %T", ok1, ok2, params["filePath"], params["patchContent"])
			err = errors.New(errors.ErrCodeTool, "Invalid parameters for ApplyPatch", nil)
			return "", err
		}
		log.Debugf("[ExecuteTool] ApplyPatch called with filePath=%s, patch_len=%d", filePath, len(patchContent))
		result, err = ApplyPatch(filePath, patchContent)
	default:
		log.Errorf("Unknown tool: %s", toolName)
		err = errors.New(errors.ErrCodeTool, fmt.Sprintf("Unknown tool: %s", toolName), nil)
		return "", err
	}

	elapsed := time.Since(start)
	if err != nil {
		log.Errorf("Tool %s failed after %s: %v", toolName, elapsed, err)
		return "", errors.New(errors.ErrCodeTool, fmt.Sprintf("tool %s failed (cwd=%s): %v", toolName, absPath, err), err)
	}

	log.Infof("Tool %s succeeded in %s. Result: %s", toolName, elapsed, result)
	log.Infof("Finished ExecuteTool")
	return result, nil
}
