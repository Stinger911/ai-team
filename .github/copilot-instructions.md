# Copilot Instructions for AI Team CLI

## Project Overview

- **Purpose:** Command-line tool to manage a team of AI agents for programming tasks, supporting OpenAI, Gemini, and Ollama models.
- **Architecture:**
  - Main entry: `main.go` (CLI parsing, command dispatch)
  - Commands: `cmd/` (one file per model/command, e.g., `openai.go`, `gemini.go`, `ollama.go`, `root.go`)
  - Core logic: `pkg/ai/` (AI orchestration, tool calls, role chains)
  - Config: `config/` (YAML config parsing, config struct)
  - Logging: `pkg/logger/`
  - Error handling: `pkg/errors/`
  - Roles & chains: `pkg/roles/`
  - Tool call handling: `pkg/tools/`
  - Shared types: `pkg/types/`

## Key Workflows

- **Build:** `make build` (outputs `ai-team` binary)
- **Test:** `make test` (unit tests, mocks AI calls)
- **Run:** `./ai-team [model] --task "[your task]"`
- **Role chains:** `./ai-team run-chain [chain] --input "key=value..."`
- **Debug:** Set `AI_TEAM_DEBUG=1` for verbose logs

## Project-Specific Patterns

- **Tool Calls:** AI responses can trigger tool calls (e.g., `write_file`) by returning a JSON object with a `tool_call` key. See `pkg/ai/` and `pkg/tools/` for implementation.
- **Role Chains:** Defined in config, executed via CLI. Chain logic in `pkg/roles/`.
- **Config:** All API keys, model settings, roles, and chains are in `config.yaml`. For debugging, you can use `work.nogit.yaml` with the same structure and sensitive information like API keys.
- **Logging:** Controlled by config and `AI_TEAM_DEBUG` env var. Logs to file and/or stdout.
- **Testing:** Unit tests mock AI/model calls. Integration tests require the built binary.

## Conventions & Integration

- **Go modules:** Standard Go project (`go.mod`, `go.sum`).
- **File output:** Output files (e.g., `design.md`, `code.py`) are written automatically if the AI response includes a `write_file` tool call.
- **Directory structure:** Keep new commands in `cmd/`, new AI logic/tools in `pkg/ai/` or subfolders.
- **Error handling:** Use `pkg/errors/` for custom error types and wrapping.

## Examples

- Add a new model: create a new file in `cmd/`, implement model logic in `pkg/ai/`, update `config.yaml`.
- Add a new tool: implement in `pkg/tools/`, update tool call handling in `pkg/ai/`.

## References

- See `README.md` for usage, config, and troubleshooting.
- See `config.yaml` for all runtime settings.
- See `pkg/roles/roles.go` and `pkg/ai/ai.go` for orchestration logic.
