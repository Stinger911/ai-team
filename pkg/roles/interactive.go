package roles

import (
	"encoding/json"
	"fmt"
	"regexp"

	"ai-team/config"
	"ai-team/pkg/ai"
	"ai-team/pkg/cli"
	"ai-team/pkg/tools"
	"ai-team/pkg/types"
)

// Session represents an interactive role-playing session.
type Session struct {
	DryRun        bool
	Model         string
	MaxIterations int
	ContextFile   string
	UI            cli.UI
	Config        *config.Config
}

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
		fmt.Printf("Error: %v\n", err)
		return
	}

	role := session.Config.Roles[selectedRole]

	// Get the inputs from the user
	inputs, err := getInputs(session, &role)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Execute the role
	output, err := ExecuteRole(role, inputs, *session.Config, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Extract the tool call from the output
	toolCall, _, err := ai.NewDefaultToolCallExtractor(toolRegistry).ExtractToolCall(output)
	if err != nil {
		fmt.Println("Role output:")
		session.UI.Pager(output)
		return
	}

	// Handle the tool call
	handleToolCall(session, toolRegistry, toolCall, &role, inputs)
}

func handleToolCall(session *Session, toolRegistry *tools.ToolRegistry, toolCall *types.ToolCall, role *types.Role, inputs map[string]interface{}) {
	for {
		// Pretty-print the tool call
		session.UI.PrettyJSON(toolCall)

		// Ask the user for confirmation
		options := []string{"Approve & execute", "Edit tool_call JSON", "Reject", "Ask LLM to re-plan"}
		selectedOption, err := session.UI.PromptSelect(options)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		switch selectedOption {
		case "Approve & execute":
			if approveAndExecute(session, toolRegistry, toolCall) {
				return
			}
		case "Edit tool_call JSON":
			toolCall = editToolCall(session, toolCall)
		case "Reject":
			fmt.Println("Tool call rejected.")
			return
		case "Ask LLM to re-plan":
			toolCall = askLLMToReplan(session, toolRegistry, role, inputs)
		}
	}
}

func approveAndExecute(session *Session, toolRegistry *tools.ToolRegistry, toolCall *types.ToolCall) bool {
	// Execute the tool call
	toolExecutor := &tools.ToolExecutor{Registry: toolRegistry}
	result, err := toolExecutor.Execute(tools.ToolCall{Name: toolCall.Name, Arguments: toolCall.Arguments})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return true
	}

	fmt.Println("Tool output:")
	session.UI.Pager(fmt.Sprintf("%v", result))
	return true
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
	output, err := ExecuteRole(*role, inputs, *session.Config, "")
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