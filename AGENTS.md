# AI Agent Team CLI Overview

## Description

AI Agent Team CLI orchestrates AI agents for programming tasks.
It supports OpenAI, Gemini, and Ollama, enabling workflows via
role chains with tool calling and file writing.

### Goals

- Unified CLI for multiple AI models.
- AI agents collaborate via role chains.
- Tool support: file ops, commands, patches.
- Configurable roles with custom prompts.
- Auto file writing from AI responses.

### Technologies

- Go 1.23+, Cobra, Viper, Logrus, HTTP/JSON

## Architecture

Modular design with clear separation:

```
┌─────────────┐
│   CLI Layer │  (cmd/ - Cobra)
└──────┬──────┘
       │
┌──────▼──────────────────────┐
│   Core Business Logic       │
├─────────────────────────────┤
│ • Roles (pkg/roles/): Orchestrates agents
│ • AI Clients (pkg/ai/): LLM communication
│ • Tools (pkg/tools/): File/command operations
└──────┬──────────────────────┘
       │
┌──────▼──────────────────────┐
│   Supporting Packages       │
├─────────────────────────────┤
│ • Config (config/): YAML configuration
│ • Types (pkg/types/): Data structures
│ • Errors (pkg/errors/): Error handling
│ • Logger (pkg/logger/): Logging utilities
└─────────────────────────────┘
```

### Components

1.  **CMD Layer:** `root.go`, `openai.go`, etc.
2.  **Roles:** `ExecuteRole()`, `ExecuteChain()`, tool calls, JSON extraction.
3.  **AI:** `CallOpenAI()`, etc., `executeToolCall()`, `ListGeminiModels()`.
4.  **Tools:** `WriteFile()`, `RunCommand()`, `ApplyPatch()`.

### Data Flow

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

**Patterns:** Context via templates, tool calls (JSON), tool results
in context, looping roles.

## Directory Structure

```
ai-team/
├── cmd/ # CLI commands
├── config/ # Configuration
├── pkg/ # Core packages
│   ├── ai/ # AI provider integrations
│   ├── roles/ # Role orchestration
│   ├── tools/ # Tool execution
│   ├── types/ # Shared types
│   ├── errors/ # Error handling
│   └── logger/ # Logging
├── main.go # Entry point
├── config.yaml # Configuration
├── Makefile # Build automation
├── go.mod # Dependencies
└── README.md # Documentation
```

### Key Files

- `config.yaml`: API keys, model configs, roles, chains.
- `main.go`: Initializes logger, executes CLI.
- `Makefile`: Build, test, clean.

### Entry Points

1.  CLI: `main.go` → `cmd.ExecuteCmd()`
2.  Role: `./ai-team role <role> --input "k=v"`
3.  Chain: `./ai-team run-chain <chain> --input "k=v"`
4.  Model: `./ai-team openai --task "task"`

## Development

### Building

```bash
make build
go build -o ai-team main.go # Manual
```

### Running

```bash
./ai-team run-chain design-code-test --input "..."
./ai-team role coder --input "design=..."
./ai-team openai --task "hello world in Go"
AI_TEAM_DEBUG=1 ./ai-team run-chain ...
```

### Testing

Go's testing framework, mocks, `httptest.NewServer()`, function injection.
`roles_integration_test.go` tests CLI execution.

```bash
make test
go test -v ./...
go test -v ./pkg/ai/
```

### Setup

1.  Go 1.23+, Git, Bash
2.  Configure `config.yaml` (API keys).
3.  `make build`, `make test`.

### Linting

```bash
go fmt ./...
go vet ./...
staticcheck ./... # If installed
golangci-lint run # If installed
```

### Cleaning

```bash
make clean
rm -f ai-team # Manual
```

## Notes for Agents

### Conventions

1.  Errors: `pkg/errors` with codes.
2.  Logging: Logrus, respect `AI_TEAM_DEBUG`.
3.  Config: Viper, no hardcoding.
4.  Testing: Mock dependencies, no real API calls.
5.  Tool Calls: JSON `tool_call` auto-executed.
6.  Context: Go templates (`{{.variable}}`).

### Patterns

- Template rendering for prompts.
- Tool-call extraction via `toolcallextract.go`.
- Tool detection: `tool_call`, `tool_name`, JSON.
- Function injection for testing.
- Structured logging with logrus.

### Extension Points

- New AI providers in `pkg/ai/`.
- Custom tools in `config.yaml`.
- New roles/chains in `config.yaml`.
- New CLI commands in `cmd/`.
