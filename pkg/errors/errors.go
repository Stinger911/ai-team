package errors

import "fmt"

// Error is a custom error type for the application.
type Error struct {
	Code    int
	Message string
	Err     error
}

// Error returns the error message.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d, message=%s, error=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

// New creates a new custom error.
func New(code int, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

const (
	// ErrCodeUnknown is the default error code.
	ErrCodeUnknown = iota
	// ErrCodeConfig is the error code for configuration errors.
	ErrCodeConfig
	// ErrCodeAPI is the error code for API errors.
	ErrCodeAPI
	// ErrCodeTool is the error code for tool execution errors.
	ErrCodeTool
	// ErrCodeRole is the error code for role execution errors.
	ErrCodeRole
)
