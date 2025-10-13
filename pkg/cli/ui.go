package cli

import (
	"encoding/json"

	"fmt"

	"io/ioutil"

	"os"

	"os/exec"

	"strings"

	"github.com/c-bata/go-prompt"
)

// DefaultUI is the default implementation of the UI interface.

type DefaultUI struct{}

// PromptSelect prompts the user to select an option from a list.

func (ui *DefaultUI) PromptSelect(options []string) (string, error) {

	fmt.Println("Please select an option:")

	completer := func(d prompt.Document) []prompt.Suggest {

		s := []prompt.Suggest{}

		for _, option := range options {

			s = append(s, prompt.Suggest{Text: option})

		}

		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)

	}

	selected := prompt.Input("> ", completer,

		prompt.OptionTitle("Select an option"),

		prompt.OptionPrefixTextColor(prompt.Yellow),

		prompt.OptionSelectedSuggestionBGColor(prompt.Blue),

		prompt.OptionSuggestionBGColor(prompt.DarkGray),
	)

	return selected, nil

}

// Confirm prompts the user for a yes/no confirmation.

func (ui *DefaultUI) Confirm(prompt string) (bool, error) {

	fmt.Printf("%s [y/n]: ", prompt)

	var response string

	_, err := fmt.Scanln(&response)

	if err != nil {

		return false, err

	}

	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y" || response == "yes", nil

}

// OpenEditor opens the user's default editor to edit the given content.

func (ui *DefaultUI) OpenEditor(content string) (string, error) {

	editor := os.Getenv("EDITOR")

	if editor == "" {

		editor = "vim"

	}

	file, err := ioutil.TempFile(os.TempDir(), "ai-team-editor-")

	if err != nil {

		return "", err

	}

	defer os.Remove(file.Name())

	if _, err := file.WriteString(content); err != nil {

		return "", err

	}

	if err := file.Close(); err != nil {

		return "", err

	}

	cmd := exec.Command(editor, file.Name())

	cmd.Stdin = os.Stdin

	cmd.Stdout = os.Stdout

	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {

		return "", err

	}

	newContent, err := ioutil.ReadFile(file.Name())

	if err != nil {

		return "", err

	}

	return string(newContent), nil

}

// Pager displays the given content in a pager.

func (ui *DefaultUI) Pager(content string) error {

	cmd := exec.Command("less")

	cmd.Stdin = strings.NewReader(content)

	cmd.Stdout = os.Stdout

	return cmd.Run()

}

// PrettyJSON prints the given object as pretty-printed JSON.

func (ui *DefaultUI) PrettyJSON(obj interface{}) error {

	b, err := json.MarshalIndent(obj, "", "  ")

	if err != nil {

		return err

	}

	fmt.Println(string(b))

	return nil

}
