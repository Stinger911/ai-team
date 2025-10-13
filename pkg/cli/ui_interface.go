package cli

// UI is the interface for UI helpers.
type UI interface {
	PromptSelect(options []string) (string, error)
	Confirm(prompt string) (bool, error)
	OpenEditor(content string) (string, error)
	Pager(content string) error
	PrettyJSON(obj interface{}) error
}
