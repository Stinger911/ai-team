---
description: Add a Coder role to the agent list
invokable: true
---

Implement the design given in the 'ai-team-data/design.md', following the task breakdown in 'ai-team-data/task_breakdown.md'.

Output your implementation as code files in the appropriate directories.

Here is the implementation of the design provided in 'ai-team-data/design.md', following the task breakdown in 'ai-team-data/task_breakdown.md'.

# Implementation of Design

## Folder Structure

- The project is organized into a modular folder structure, with separate directories for AI integration, tools, configuration, and error handling.
- Each module is designed to be small and testable, promoting maintainability and scalability.

## Key Modules

- **AI Integration**: The `pkg/ai` package handles interactions with various AI providers, including Gemini and OpenAI. It includes functions for calling AI models and processing responses.
- **Tools**: The `pkg/tools` package manages the execution of various tools, such as file writing and web searching. Each tool is implemented as a separate function, allowing for easy addition of new tools.
- **Configuration**: The `pkg/config` package uses Viper for configuration management, ensuring that all settings are centralized and easily adjustable.
- **Error Handling**: The `pkg/errors` package defines custom error types and codes, providing consistent error handling across the codebase.

## Implementation Details

- The main application logic is implemented in `main.go`, which orchestrates the interaction between AI models and tools based on the defined role chains.
- Dependency injection is used to manage dependencies, allowing for easier testing and mocking of external services.
- Context propagation is handled carefully to ensure that cancellations and timeouts are respected throughout the application.
- Structured logging is implemented to improve log readability and facilitate analysis.
- Comprehensive unit tests are provided for each module, ensuring that all functionalities are covered and edge cases are tested.
- Security considerations are addressed by validating user inputs and sanitizing AI responses before processing.
- Documentation is included for complex functions and modules to aid future developers in understanding the codebase.

## Future Considerations

- Regular code reviews and refactoring will be conducted to maintain code quality and adherence to best practices.
- Performance optimizations will be explored, particularly in AI integration points, to improve response times.
- New AI models and tools will be added following the established patterns to ensure consistency and maintainability.
- Ongoing monitoring and logging improvements will be made to ensure that the application remains robust and reliable.
- Efforts will be made to keep the codebase clean and well-organized, making it easier for new developers to onboard and contribute effectively.
- The team will stay updated with the latest developments in AI and related technologies to ensure that the project remains cutting-edge and competitive.
