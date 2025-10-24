package roles

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"ai-team/config"
	"ai-team/pkg/ai"
	"ai-team/pkg/cli"
	"ai-team/pkg/tools"
	"ai-team/pkg/types"
)

// Session represents an interactive role-playing session.
type Session struct {
	DryRun         bool
	Model          string
	MaxIterations  int
	ContextFile    string
	UI             cli.UI
	Config         *config.Config
	Transcript     *types.Transcript
	TranscriptPath string
	Yes            bool
}

// ExecuteRoleFunc is a variable that holds the function to execute a role.
// It can be replaced in tests for mocking.
var ExecuteRoleFunc = ExecuteRole

// NewToolCallExtractorFunc is a variable that holds the function to create a new tool call extractor.
// It can be replaced in tests for mocking.
var NewToolCallExtractorFunc = ai.NewDefaultToolCallExtractor

// StartSession starts a new interactive session.
func StartSession(session *Session) {
	fmt.Printf("Interactive session starting with options: %+v\n", session)

	confirm, err := session.UI.Confirm("Start session?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if !confirm {
		fmt.Println("Session aborted.")
		return
	}

	// Create a new tool registry
	toolRegistry := tools.NewToolRegistry()

	tools.RegisterDefaultTools(toolRegistry)

	// Get the role from the user
	selectedRole, err := getRole(session)
	if err != nil {
		fmt.Printf("Error getting role: %v\n", err)
		return
	}

	role := session.Config.Roles[selectedRole]

	session.Transcript = &types.Transcript{
		Role:      selectedRole,
		StartedAt: time.Now(),
		Steps:     []types.Step{},
	}

	// Get the inputs from the user
	inputs, err := getInputs(session, &role)
	if err != nil {
		fmt.Printf("Error getting inputs: %v\n", err)
		return
	}

	// Execute the role
	output, err := ExecuteRoleFunc(role, inputs, session.Config, "")
	if err != nil {
		fmt.Printf("Error executing role: %v\n", err)
		return	
	}

	// Extract the tool call from the output
	toolCall, _, err := NewToolCallExtractorFunc(toolRegistry).ExtractToolCall(output)
	if err != nil {
		fmt.Println("Role output:")
		session.UI.Pager(output)
		return
	}

	// Handle the tool call
	handleToolCall(session, toolRegistry, toolCall, &role, inputs)

	// Write transcript if path is provided
	if session.TranscriptPath != "" {
		err := writeTranscript(session.TranscriptPath, session.Transcript)
		if err != nil {
			fmt.Printf("Error writing transcript: %v\n", err)
		} else {
			fmt.Printf("Transcript written to: %s\n", session.TranscriptPath)
		}
	}
}

func writeTranscript(filePath string, transcript *types.Transcript) error {
	data, err := json.MarshalIndent(transcript, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func handleToolCall(session *Session, toolRegistry *tools.ToolRegistry, toolCall *types.ToolCall, role *types.Role, inputs map[string]interface{}) {
	for i := 0; i < session.MaxIterations; i++ {
		// Pretty-print the tool call
		session.UI.PrettyJSON(toolCall)

		step := types.Step{
			ToolCall:  toolCall,
			Approved:  false,
			Result:    nil,
		}

		var selectedOption string
		if session.Yes {
			selectedOption = "Approve & execute"
		} else {
			options := []string{"Approve & execute", "Edit tool_call JSON", "Reject", "Ask LLM to re-plan"}
			var err error
			selectedOption, err = session.UI.PromptSelect(options)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				session.Transcript.Steps = append(session.Transcript.Steps, step) // Record step before returning
				return
			}
		}

		switch selectedOption {
		case "Approve & execute":
			result, continueLoop := approveAndExecute(session, toolRegistry, toolCall, session.DryRun)
			step.Approved = true
			step.Result = result
			if !continueLoop {
				session.Transcript.Steps = append(session.Transcript.Steps, step)
				return
			}
			inputs["tool_output"] = result
		case "Edit tool_call JSON":
			toolCall = editToolCall(session, toolCall)
			session.Transcript.Steps = append(session.Transcript.Steps, step) // Record step after edit
			continue
		case "Reject":
			fmt.Println("Tool call rejected.")
			session.Transcript.Steps = append(session.Transcript.Steps, step)
			return
		case "Ask LLM to re-plan":
			// Get the new instruction from the user
			fmt.Println("Enter new instruction:")
			newInstruction, err := session.UI.OpenEditor("")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				session.Transcript.Steps = append(session.Transcript.Steps, step)
				return
			}

			// Execute the role again with the new instruction
			inputs["instruction"] = newInstruction
			output, err := ExecuteRoleFunc(*role, inputs, session.Config, "")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				session.Transcript.Steps = append(session.Transcript.Steps, step)
				return
			}
			step.LlmOutput = output

			// Extract the tool call from the output
			newToolCall, _, err := NewToolCallExtractorFunc(toolRegistry).ExtractToolCall(output)
			if err != nil {
				fmt.Println("Role output:")
				session.UI.Pager(output)
				session.Transcript.Steps = append(session.Transcript.Steps, step)
				return
			}
			toolCall = newToolCall
			session.Transcript.Steps = append(session.Transcript.Steps, step)
			continue
		}

		// If we approved and executed, now get the next LLM output
		output, err := ExecuteRoleFunc(*role, inputs, session.Config, "")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			session.Transcript.Steps = append(session.Transcript.Steps, step)
			return
		}
		step.LlmOutput = output

		newToolCall, _, err := NewToolCallExtractorFunc(toolRegistry).ExtractToolCall(output)
		if err != nil {
			fmt.Println("Role output:")
			session.UI.Pager(output)
			session.Transcript.Steps = append(session.Transcript.Steps, step)
			return
		}
		toolCall = newToolCall
		session.Transcript.Steps = append(session.Transcript.Steps, step)
	}
}

func approveAndExecute(session *Session, toolRegistry *tools.ToolRegistry, toolCall *types.ToolCall, dryRun bool) (interface{}, bool) {
	if dryRun {
		fmt.Println("DRY RUN: Tool call would be:")
		session.UI.PrettyJSON(toolCall)

		if toolCall.Name == "write_file" || toolCall.Name == "WriteFile" {
			filePath, ok := toolCall.Arguments["file_path"].(string)
			if !ok {
				fmt.Printf("Error: Missing or invalid 'file_path' argument for write_file tool.\n")
				return nil, false
			}
			content, ok := toolCall.Arguments["content"].(string)
			if !ok {
				fmt.Printf("Error: Missing or invalid 'content' argument for write_file tool.\n")
				return nil, false
			}
			oldContent := tools.ReadFileOrEmpty(filePath)
			diff := tools.GenerateUnifiedDiff(filePath, oldContent, content)
			fmt.Println("DRY RUN: Diff:")
			fmt.Println(diff)
		}

		return nil, true
	}

	if toolCall.Name == "write_file" || toolCall.Name == "WriteFile" {
		filePath, ok := toolCall.Arguments["file_path"].(string)
		if !ok {
			fmt.Printf("Error: Missing or invalid 'file_path' argument for write_file tool.\n")
			return nil, false
		}
		content, ok := toolCall.Arguments["content"].(string)
		if !ok {
			fmt.Printf("Error: Missing or invalid 'content' argument for write_file tool.\n")
			return nil, false
		}
		oldContent := tools.ReadFileOrEmpty(filePath)
		diff := tools.GenerateUnifiedDiff(filePath, oldContent, content)
		fmt.Println("Diff:")
		fmt.Println(diff)

		confirm, err := session.UI.Confirm("Apply this change?")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return nil, false
		}
		if !confirm {
			fmt.Println("Change rejected.")
			return nil, false
		}

		backupPath, err := tools.BackupFile(filePath)
		if err != nil {
			fmt.Printf("Error creating backup: %v\n", err)
			return nil, false
		}
		if backupPath != "" {
			fmt.Printf("Backup created at: %s\n", backupPath)
		}
	}

	if toolCall.Name == "run_command" || toolCall.Name == "RunCommand" {
		command, ok := toolCall.Arguments["command"].(string)
		if !ok {
			fmt.Printf("Error: Missing or invalid 'command' argument for run_command tool.\n")
			return nil, false
		}
		fmt.Printf("Command to execute: %s\n", command)

		confirm, err := session.UI.Confirm("Execute this command?")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return nil, false
		}
		if !confirm {
			fmt.Println("Command rejected.")
			return nil, false
		}
	}

	// Execute the tool call
	toolExecutor := &tools.ToolExecutor{Registry: toolRegistry}
	result, err := toolExecutor.Execute(tools.ToolCall{Name: toolCall.Name, Arguments: toolCall.Arguments})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil, false
	}

	fmt.Println("Tool output:")
	session.UI.Pager(fmt.Sprintf("%v", result))
	return result, true
}

func editToolCall(session *Session, toolCall *types.ToolCall) *types.ToolCall {
	// Open the editor to edit the tool call JSON
	jsonBytes, err := json.MarshalIndent(toolCall, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return toolCall
	}

	editedJSON, err := session.UI.OpenEditor(string(jsonBytes))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return toolCall
	}

	// Parse the edited JSON
	var editedToolCall types.ToolCall
	if err := json.Unmarshal([]byte(editedJSON), &editedToolCall); err != nil {
		fmt.Printf("Error: %v\n", err)
		return toolCall
	}

	return &editedToolCall
}

func getRole(session *Session) (string, error) {
	// Get the available roles from the configuration
	var roleNames []string
	for name := range session.Config.Roles {
		roleNames = append(roleNames, name)
	}

	// Prompt the user to select a role
	selectedRole, err := session.UI.PromptSelect(roleNames)
	if err != nil {
		return "", err
	}

	return selectedRole, nil
}

func getInputs(session *Session, role *types.Role) (map[string]interface{}, error) {
	inputs := make(map[string]interface{})

	// Get the inputs required by the role by parsing the prompt
	re := regexp.MustCompile(`{{\.(.*?)}}`)
	matches := re.FindAllStringSubmatch(role.Prompt, -1)

	for _, match := range matches {
		inputName := match[1]

		// Prompt the user for the input
		fmt.Printf("Enter value for input '%s': ", inputName)
		value, err := session.UI.OpenEditor("")
		if err != nil {
			return nil, err
		}
		inputs[inputName] = value
	}

	return inputs, nil
}

func askLLMToReplan(session *Session, toolRegistry *tools.ToolRegistry, role *types.Role, inputs map[string]interface{}) *types.ToolCall {
	// Get the new instruction from the user
	fmt.Println("Enter new instruction:")
	newInstruction, err := session.UI.OpenEditor("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil
	}

	// Execute the role again with the new instruction
	inputs["instruction"] = newInstruction
	output, err := ExecuteRole(*role, inputs, session.Config, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil
	}

	// Extract the tool call from the output
	newToolCall, _, err := ai.NewDefaultToolCallExtractor(toolRegistry).ExtractToolCall(output)
	if err != nil {
		fmt.Println("Role output:")
		session.UI.Pager(output)
		return nil
	}

	return newToolCall
}