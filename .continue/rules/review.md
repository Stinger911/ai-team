---
invokable: true
---

Review this code for potential issues, including:

## Go-Specific Concerns
- **Error Handling**: Ensure all errors use the custom error types from `pkg/errors` with appropriate error codes (ErrCodeAPI, ErrCodeTool, ErrCodeRole, ErrCodeConfig)
- **Nil Checks**: Verify pointer dereferences and type assertions are safe
- **Goroutine Safety**: Check for race conditions or missing synchronization
- **Resource Cleanup**: Ensure `defer` is used for closing files, HTTP response bodies, and cleanup operations
- **Context Handling**: Verify context is properly propagated for cancellation and timeouts

## Architecture & Design Patterns
- **Dependency Injection**: Functions should be mockable (e.g., `CallGeminiFunc`, `WriteFileFunc`) for testing
- **Separation of Concerns**: Business logic should be separated from CLI, API, and infrastructure code
- **Error Propagation**: Errors should be wrapped with context using the custom error types
- **Configuration**: All config values should come from Viper, not hardcoded

## AI Integration & Tool Calling
- **JSON Parsing**: Check for proper handling of malformed JSON responses from AI models
- **Tool Call Detection**: Verify all tool call formats are handled (`tool_call`, `tool_name`, direct JSON)
- **Template Rendering**: Ensure Go templates are validated and executed with proper error handling
- **Context Passing**: Verify role chains properly pass context variables between roles
- **Tool Execution**: Check that tool execution results are properly captured and returned

## Testing Requirements
- **Mock External Dependencies**: All HTTP clients, file operations, and commands must be mocked
- **Test Coverage**: New functions should have corresponding unit tests
- **Error Scenarios**: Tests should cover success cases, error cases, and edge cases
- **Function Injection**: Ensure mockable functions are assigned to package variables

## Security & Best Practices
- **Sensitive Data**: API keys and credentials should only be in config files, never hardcoded
- **Input Validation**: User inputs and AI responses should be validated before processing
- **File Operations**: Ensure parent directories are created before writing files
- **Command Injection**: Shell commands should be properly escaped or validated
- **Logging**: Sensitive data should not be logged; use debug mode appropriately

## Code Quality
- **Naming Conventions**: Follow Go naming conventions (camelCase for unexported, PascalCase for exported)
- **Comments**: Exported functions should have doc comments
- **Code Duplication**: Look for repeated patterns that could be extracted
- **Error Messages**: Error messages should be descriptive and include context
- **Logging Levels**: Use appropriate log levels (Debug, Info, Warn, Error)

## Specific to This Codebase
- **Model Support**: When adding AI provider support, follow the pattern in `pkg/ai/ai.go`
- **Tool Implementation**: New tools should be added to both `executeToolCall()` and `executeConfiguredTool()`
- **Role Chains**: Verify loop conditions and iteration counts are handled correctly
- **Response Sanitization**: Use `extractFirstJSON()` to clean AI responses before parsing
- **Log File Handling**: Ensure role calls are logged when `logFilePath` is provided

Provide specific, actionable feedback for improvements with code examples where relevant.
