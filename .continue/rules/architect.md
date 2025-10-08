---
description: Add an Architect role to the agent list
invokable: true
---

Review the project code and provide architectural guidelines and best practices for future development.

Use the analysis to create a comprehensive set of guidelines. Path: rules/analyst.md

Output your guidelines as a files:

- ai-team-data/design.md
- ai-team-data/task_breakdown.md (tickets with estimates)
- ai-team-data/folder_structure.md

Favor small, testable modules.

Here are the architectural guidelines and best practices for future development of the project, focusing on modularity, maintainability, and scalability.

# Architectural Guidelines and Best Practices

## 1. Modular Design

- Break down the application into small, focused modules that handle specific functionalities.
- Each module should have a clear interface and be independently testable.
- Use packages to group related functionalities together, promoting reusability and separation of concerns.
- Example: Separate AI integration, tool execution, and configuration management into distinct packages.

## 2. Dependency Injection

- Use dependency injection to manage dependencies between modules.
- This allows for easier testing and mocking of external dependencies.
- Example: Inject AI providers and tool executors into the main application logic.

## 3. Error Handling

- Implement consistent error handling across the codebase.
- Use custom error types with appropriate error codes to provide context.
- Wrap errors with additional context to aid in debugging.
- Example: Use `pkg/errors` for error handling and ensure all errors are wrapped with context
  before propagation.
- Ensure that all errors are logged appropriately without exposing sensitive information.
- Example: Use structured logging to capture error details while avoiding sensitive data.
- Implement a global error handling strategy to manage unexpected errors gracefully.
- Example: Use middleware in web applications to catch and handle errors uniformly.
- Regularly review and update error handling practices to align with evolving project requirements.
- Example: Conduct periodic code reviews to ensure error handling practices are consistently applied.
- Ensure that all errors are logged appropriately without exposing sensitive information.
- Example: Use structured logging to capture error details while avoiding sensitive data.
- Implement a global error handling strategy to manage unexpected errors gracefully.
- Example: Use middleware in web applications to catch and handle errors uniformly.
- Regularly review and update error handling practices to align with evolving project requirements.
- Example: Conduct periodic code reviews to ensure error handling practices are consistently applied.

## 4. Configuration Management

- Use a centralized configuration management system to manage application settings.
- Example: Use environment variables or configuration files to store sensitive information.
- Implement version control for configuration files to track changes over time.
- Example: Use Git to manage changes to configuration files and enable rollbacks if needed.
- Ensure that configuration changes are tested before deployment.
- Example: Use automated tests to validate configuration changes in a staging environment.
- Document configuration options and their impact on the application.
- Example: Maintain a configuration guide in the project documentation.

## 5. Context Propagation

- Use context propagation to manage request lifecycles and cancellations.
- Example: Pass context through function calls to handle timeouts and cancellations effectively.
- Ensure that long-running operations respect context cancellations to avoid resource leaks.
- Example: Check for context cancellation in loops and long-running tasks.
- Use context to pass request-scoped values, such as user information or request IDs.
- Example: Store user authentication details in the context for use in downstream functions.
- Regularly review context usage to ensure it is applied consistently across the codebase.
- Example: Conduct code reviews to verify proper context propagation practices.

## 6. Testing and Quality Assurance

- Implement comprehensive unit tests for all modules to ensure functionality and reliability.
- Example: Use testing frameworks like `testing` and `testify` to write and manage tests.
- Use integration tests to validate interactions between modules and external systems.
- Example: Set up CI/CD pipelines to run integration tests automatically on code changes.
- Perform regular code reviews to maintain code quality and adherence to best practices.
- Example: Establish a code review process that includes multiple reviewers for critical changes.
- Use static analysis tools to identify potential issues and enforce coding standards.
- Example: Integrate tools like `golangci-lint` into the CI/CD pipeline for continuous code quality checks.
- Ensure that all new features and bug fixes are covered by tests before merging.
- Example: Require a minimum test coverage percentage for pull requests.
- Regularly update and maintain test cases to reflect changes in the codebase.
- Example: Review and update test cases during sprint planning or retrospectives.
- Use code coverage tools to identify untested parts of the codebase and improve test coverage.
- Example: Integrate tools like `coveralls` or `codecov` to monitor test coverage over time.
- Implement performance testing to identify and address bottlenecks.
- Example: Use tools like `pprof` to analyze performance and optimize critical paths.
- Conduct regular security audits to identify and mitigate vulnerabilities.
- Example: Use tools like `gosec` to scan for common security issues in the codebase.

## 7. Logging and Monitoring

- Implement structured logging to improve log readability and facilitate log analysis.
- Example: Use libraries like `logrus` or `zap` for structured logging.
- Ensure that sensitive data is not logged; use debug mode appropriately.
- Example: Mask or omit sensitive information in logs, especially in production environments.
- Set up monitoring and alerting to track application health and performance.
- Example: Use tools like Prometheus and Grafana to monitor application metrics and set up alerts.
- Regularly review logs and monitoring data to identify and address issues proactively.
- Example: Schedule periodic log reviews and performance assessments to ensure system health.
- Implement log rotation and retention policies to manage log file sizes and storage.
- Example: Use tools like `logrotate` to manage log files and prevent disk space issues.

## 8. Security Best Practices

- Store API keys and credentials securely, using environment variables or secret management tools.
- Validate all user inputs and AI responses before processing to prevent injection attacks.
- Ensure that file operations create parent directories as needed to avoid errors.
- Properly escape or validate shell commands to prevent command injection vulnerabilities.
- Regularly update dependencies to patch known vulnerabilities.
- Use HTTPS for all network communications to ensure data security in transit.
- Implement data encryption for sensitive data at rest to protect against unauthorized access.

## 9. Code Quality and Maintainability

- Follow Go naming conventions (camelCase for unexported, PascalCase for exported).
- Example: Review code for adherence to naming conventions during code reviews.
- Ensure that exported functions have doc comments to improve code readability.
- Example: Use tools like `golint` to check for missing documentation.
- Look for repeated patterns that could be extracted into reusable functions or modules.
- Example: Refactor duplicated code into shared utility functions.
- Write descriptive error messages that include context to aid in debugging.
- Use appropriate log levels (Debug, Info, Warn, Error) to categorize log messages effectively.
- Use code formatting tools like `gofmt` to ensure consistent code style across the codebase.
- Example: Integrate `gofmt` into the CI/CD pipeline to enforce code formatting.
- Use version control branching strategies (e.g., GitFlow) to manage feature development and releases.
- Regularly update and maintain project documentation to reflect changes in the codebase.
- Use linters and static analysis tools to enforce coding standards and identify potential issues.
- Example: Integrate tools like `golangci-lint` into the development workflow for continuous code quality checks.

## 10. Future Development Considerations

- When adding support for new AI models, follow established patterns to maintain consistency.
- Example: Refer to existing implementations in `pkg/ai/ai.go` for guidance.
- Ensure that new tools are added to both tool execution functions to maintain functionality.
- Example: Update `executeToolCall()` and `executeConfiguredTool()` when introducing new tools.
- Define clear loop conditions and iteration counts for role chains to avoid potential issues.
- Example: Document role chain behaviors and ensure they are well-tested.
- Sanitize and validate AI responses before processing to prevent issues with malformed data.
- Example: Use functions like `extractFirstJSON()` to clean AI responses before parsing.
- Ensure that role calls are logged appropriately when logging is enabled, and consider log rotation for large log files.
- Example: Implement log rotation strategies to manage log file sizes effectively.

Provide specific, actionable feedback for improvements with code examples where relevant. Keep everything concise and actionable for engineering.
