# Repository Overview

## Project Description

**AI Agent Team CLI** is a command-line tool that orchestrates multiple AI agents to work together on programming tasks. It supports multiple LLM providers (OpenAI, Google Gemini, and Ollama) and enables complex workflows through role chains, where AI agents can perform sequential tasks, execute tools, and write files automatically.

### Main Purpose and Goals

- Provide a unified CLI interface for interacting with multiple AI models
- Enable AI agents to work together through predefined role chains
- Support tool calling for file operations, command execution, and patches
- Allow configurable AI roles with customizable prompts and behaviors
- Automatically write output files when AI responses include tool calls

### Key Technologies Used

- **Go 1.23+** - Core language
- **Cobra** - CLI framework
- **Viper** - Configuration management
- **Logrus** - Structured logging
- **HTTP/JSON** - API communication with AI providers

## Architecture Overview

### High-Level Architecture

The application follows a clean, modular architecture with clear separation of concerns:

```
┌─────────────┐
│   CLI Layer │  (cmd/)
│   (Cobra)   │
└──────┬──────┘
       │
┌──────▼──────────────────────┐
│   Core Business Logic       │
├─────────────────────────────┤
│ • Roles (pkg/roles/)        │  ← Orchestrates AI agents
│ • AI Clients (pkg/ai/)      │  ← Communicates with LLMs
│ • Tools (pkg/tools/)        │  ← Executes file/command operations
└──────┬──────────────────────┘
       │
┌──────▼──────────────────────┐
│   Supporting Packages       │
├─────────────────────────────┤
│ • Config (config/)          │  ← YAML configuration
│ • Types (pkg/types/)        │  ← Shared data structures
│ • Errors (pkg/errors/)      │  ← Error handling
│ • Logger (pkg/logger/)      │  ← Logging utilities
└─────────────────────────────┘
```

### Main Components and Relationships

1. **CMD Layer** (`cmd/`)
   - `root.go` - Root command, config loading, chain/role execution
   - `openai.go`, `gemini.go`, `ollama.go` - Model-specific commands
2. **Roles Package** (`pkg/roles/`)

   - `ExecuteRole()` - Executes a single AI role with templated prompts
   - `ExecuteChain()` - Orchestrates sequences of roles with context passing
   - Tool call detection and execution
   - JSON extraction from AI responses

3. **AI Package** (`pkg/ai/`)

   - `CallOpenAI()`, `CallGemini()`, `CallOllama()` - Provider-specific API clients
   - `executeToolCall()` - Processes tool requests from AI responses
   - `ListGeminiModels()` - Model discovery

4. **Tools Package** (`pkg/tools/`)
   - `WriteFile()` - File writing with directory creation
   - `RunCommand()` - Shell command execution
   - `ApplyPatch()` - Patch application

### Data Flow and System Interactions

```
User Command → CLI Parser → Config Load → Role Chain Executor
                                               ↓
                              ┌────────────────┴────────────────┐
                              │                                 │
                         Role 1 (AI Call)                  Role 2 (AI Call)
                              ↓                                 ↓
                    Response with tool_call          Response with tool_call
                              ↓                                 ↓
                      Execute Tool (write_file)      Execute Tool (run_command)
                              ↓                                 ↓
                         Store in Context ──────────────────────┘
                                               ↓
                                        Final Output
```

**Key Patterns:**

- Roles pass context between each other using template variables
- AI responses can include tool calls (JSON format)
- Tool execution results are stored in context for subsequent roles
- Supports looping roles with configurable iteration counts

## Directory Structure

```
ai-team/
├── cmd/                      # CLI commands and entry points
│   ├── root.go              # Root command, chain/role execution
│   ├── openai.go            # OpenAI-specific commands
│   ├── gemini.go            # Gemini-specific commands
│   └── ollama.go            # Ollama-specific commands
│
├── config/                   # Configuration management
│   └── config.go            # Config loading with Viper
│
├── pkg/                      # Core packages
│   ├── ai/                  # AI provider integrations
│   │   ├── ai.go           # API clients for OpenAI, Gemini, Ollama
│   │   └── ai_test.go      # AI client tests
│   │
│   ├── roles/              # Role orchestration
│   │   ├── roles.go        # Role and chain execution logic
│   │   ├── roles_test.go   # Unit tests
│   │   └── roles_integration_test.go
│   │
│   ├── tools/              # Tool execution
│   │   └── tools.go        # File ops, command execution, patches
│   │
│   ├── types/              # Shared types and structs
│   │   └── types.go        # API responses, roles, chains, tools
│   │
│   ├── errors/             # Error handling
│   │   └── errors.go       # Custom error types with codes
│   │
│   └── logger/             # Logging utilities
│       └── logger.go       # Debug logging, role call logging
│
├── main.go                  # Application entry point
├── config.yaml             # Main configuration file (API keys, roles, chains)
├── Makefile                # Build automation
├── go.mod                  # Go module dependencies
└── README.md               # Project documentation
```

### Key Files and Configuration

- **config.yaml** - Contains API keys, model configurations, role definitions, and role chains (sensitive file)
- **main.go** - Initializes logger and executes root command
- **Makefile** - Provides targets for build, test, and clean operations

### Entry Points

1. **CLI Entry**: `main.go` → `cmd.ExecuteCmd()`
2. **Single Role**: `./ai-team role <role-name> --input "key=value"`
3. **Role Chain**: `./ai-team run-chain <chain-name> --input "key=value"`
4. **Direct Model**: `./ai-team openai --task "your task"`

## Development Workflow

### Building the Project

```bash
# Build the binary
make build

# Or manually
go build -o ai-team main.go
```

The compiled binary is named `ai-team` and placed in the project root.

### Running the Project

```bash
# Execute a role chain
./ai-team run-chain design-code-test --input "initial_problem=Create a calculator"

# Execute a single role
./ai-team role coder --input "design=calculator spec"

# Direct model interaction
./ai-team openai --task "write a hello world program in Go"

# Enable debug logging
AI_TEAM_DEBUG=1 ./ai-team run-chain design-code-test --input "problem=test"
```

### Testing Approach

The project uses **Go's built-in testing framework** with comprehensive test coverage:

**Unit Tests:**

- Mock all external dependencies (HTTP clients, file operations)
- Use `httptest.NewServer()` for testing API clients
- Function injection for testability (`CallGeminiFunc`, `WriteFileFunc`, etc.)
- Test files located alongside source code (`*_test.go`)

**Integration Tests:**

- `roles_integration_test.go` - Tests CLI execution (skipped if binary missing)
- Tests role chains end-to-end

**Running Tests:**

```bash
# Run all tests
make test

# Or manually with verbose output
go test -v ./...

# Run specific package tests
go test -v ./pkg/ai/
go test -v ./pkg/roles/
```

**Test Coverage Areas:**

- API client responses (success, errors, malformed JSON, network failures)
- Tool call parsing and execution
- Role template rendering
- Chain execution and context passing
- Error handling and custom error codes

### Development Environment Setup

**Prerequisites:**

- Go 1.23.0 or higher
- Git
- Bash (for Makefile and tool commands)
- Access to at least one AI provider (OpenAI, Gemini, or Ollama)

**Setup Steps:**

1. Clone the repository
2. Copy and configure `config.yaml` with your API keys
3. Run `make build` to compile the binary
4. Run `make test` to verify everything works

**Configuration:**

- Edit `config.yaml` to add API keys for your chosen providers
- Define custom roles with prompts and model preferences
- Create role chains for complex workflows
- Configure tools with custom commands

### Lint and Format Commands

Currently, no linting tools are explicitly configured in the Makefile, but Go developers typically use:

```bash
# Format code (standard Go)
go fmt ./...

# Vet code for common issues
go vet ./...

# Run staticcheck (if installed)
staticcheck ./...

# Run golangci-lint (if installed)
golangci-lint run
```

### Cleaning Up

```bash
# Remove compiled binary
make clean

# Or manually
rm -f ai-team
```

## Additional Notes for Agents

### Important Conventions

1. **Error Handling**: Use custom error types from `pkg/errors` with error codes
2. **Logging**: Use logrus for structured logging; respect `AI_TEAM_DEBUG` env var
3. **Configuration**: All config goes through Viper; don't hardcode values
4. **Testing**: Always mock external dependencies; no real API calls in tests
5. **Tool Calls**: AI responses with `tool_call` JSON are automatically executed
6. **Context Passing**: Chain roles use Go templates (`{{.variable}}`) for context

### Common Patterns

- **Template Rendering**: Prompts use Go's `text/template` for variable substitution
- **Tool-Call Extraction**: Responses are parsed using a robust handler-based pipeline (`pkg/ai/toolcallextract.go`) that supports multiple formats (JSON code blocks, inline JSON, tool_call, tool_name, etc.) and strict schema validation. Legacy `extractFirstJSON()` is now replaced in orchestration logic.
- **Tool Detection**: Multiple formats supported (`tool_call`, `tool_name`, direct JSON with `file_path`)
- **Function Injection**: Functions assigned to variables for test mocking
- **Structured Logging**: Use logrus with fields for better log analysis

### Extension Points

- Add new AI providers in `pkg/ai/` following the existing pattern
- Define custom tools in `config.yaml` with command templates
- Create new roles and chains in `config.yaml`
- Add new CLI commands in `cmd/` following Cobra conventions
