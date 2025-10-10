# AI Agent Team CLI

A command-line tool to manage a team of AI agents for programming.

## Description

This tool allows you to interact with different AI models (OpenAI, Gemini, and Ollama) from the command line. You can provide a task to the AI model and get a response.

## Dependencies

This project uses the following main Go libraries:

- `github.com/sirupsen/logrus` — Structured logging
- `github.com/spf13/cobra` — CLI framework
- `github.com/spf13/viper` — Configuration management
- `gopkg.in/yaml.v3` — YAML parsing

All dependencies are managed via Go modules (`go.mod`). Run `go mod tidy` to ensure all dependencies are installed.

## Installation

1. Clone the repository:

```bash
git clone https://github.com/Stinger911/ai-team.git
```

2. Go to the project directory:

```bash
cd ai-team
```

3. Build the binary:

```bash
make build
```

## Usage

```bash
./ai-team [model] --task "[your task]"
```

### Models

- `openai`: Use the OpenAI model.
- `gemini`: Use the Gemini model.
- `ollama`: Use the Ollama model.

### Example

```bash
./ai-team openai --task "write a hello world program in Go"
```

### Role Chains & Automatic File Output

You can execute predefined chains of AI roles, and the system will automatically write output files (such as `design.md`, `code.py`, `test_cases.md`, etc.) if the AI response includes a tool call to `write_file`:

```bash
./ai-team run-chain design-code-test --input "initial_problem=Create a calculator function"
```

### Running a Single Role

You can run a single role directly (without a chain):

```bash
./ai-team role coder --input "design=your design here"
```

**How it works:**

- If the AI model returns a JSON object with a top-level `tool_call` (e.g., `{ "tool_call": { "name": "write_file", "arguments": { "file_path": "design.md", "content": "..." }}}`), the file will be written automatically.
- Output files are created in the current working directory unless otherwise specified.
- If you do not see the expected files, enable debug logging (see below) and check for warnings about file writing in the logs.

### Debugging & Troubleshooting File Output

To enable debug output for troubleshooting tool execution and file writing:

```bash
AI_TEAM_DEBUG=1 ./ai-team run-chain design-code-test --input "initial_problem=Create a calculator function"
```

Check the log output for lines like:

- `[ToolCallWrap] Writing file: design.md`
- `[ToolCall] Writing file: code.py`
- `[Fallback] Writing file: ...`

If you see warnings such as `file_path is empty, skipping file write`, check that your AI prompt and role chain are producing the correct tool call JSON structure.

## Robust Tool-Call Extraction

AI responses are now parsed using a robust extraction pipeline that supports:

- JSON code blocks (`json ... `)
- Inline JSON objects
- Multiple tool-call formats (tool_call, tool_name, direct JSON)
- Strict schema validation for tool-calls
- Graceful error handling and fallback for malformed or ambiguous responses

If a tool-call is present in the response, it will be detected and executed automatically. If the response is malformed, the system will log a warning and attempt to recover or skip the tool-call.

See `pkg/ai/toolcallextract.go` for implementation details and `pkg/ai/toolcallextract_test.go` for test cases.

## Configuration

The tool uses a `config.yaml` file to configure the API keys, URLs, roles, chains, and logging. Example keys:

```yaml
LogFilePath: "ai-team.log"
LogStdout: true  # Set to false to log only to file
Gemini:
	APIKey: "..."
	APIURL: "..."
	Model: "gemini-2.5-flash"
Roles:
	- Name: "architect"
		...
Chains:
	- Name: "design-code-test"
		...
```

## Development

### Running tests

```bash
make test
```

- Unit tests mock all AI calls and do not require network or API keys.
- Integration tests (CLI) are skipped if the `ai-team` binary is not present.

### Building the binary

```bash
make build
```

### Cleaning up

```bash
make clean
```

## Recent changes: tool-call extraction & validation

Small but important updates were made to make tool-call extraction and execution more tolerant and robust when AI model outputs vary in naming conventions or JSON formatting. Highlights:

- The `ToolCallExtractor` now accepts multiple formats (inline JSON, JSON code blocks, and recursive JSON searches) and is constructed with a `ToolRegistry` so extraction uses the same schemas as execution.
- The `ToolRegistry` and tools accept both camelCase and snake_case tool names and argument keys (for example `WriteFile` / `write_file`, `filePath` / `file_path`).
- Argument lookup and validation are flexible: `lookupArgFlexible` matches case-insensitive, snake_case, and camelCase variants and tolerates JSON numeric types that represent integers.
- Tests added to cover these behaviors: validation permutations, extractor registry usage, executor retry/timeout behavior, lookup variants, and negative/malformed-response cases. See `pkg/tools/*_test.go` and `pkg/ai/*_test.go`.

These changes reduce false negatives like `"No valid tool-call found in response"` when the model returns valid but differently-formatted tool-call JSON.

---

## Prompting and Tool-Call Conventions

When authoring role prompts (in `work.nogit.yaml` or `config.yaml`) follow these conventions to make tool calls reliable and machine-parseable:

- Output only one JSON object per response when you mean to request a tool call. Avoid extra text, Markdown, or status messages.
- Use snake_case for tool names and argument keys (the system is tolerant, but snake_case is canonical): `write_file`, `read_file`, `list_dir`, `run_command`, `apply_patch`.
- The exact expected structure for tool calls is:

  {"tool_call": {"name": "<tool_name>", "arguments": {<arg_key>: <arg_value>}}}

- Example canonical tool calls:
  - Write a file (preferred):
    {"tool_call": {"name": "write_file", "arguments": {"file_path": "ai-team-data/design.md", "content": "# Design\n..."}}}
  - Read a file:
    {"tool_call": {"name": "read_file", "arguments": {"file_path": "ai-team-data/design.md"}}}
  - List directory:
    {"tool_call": {"name": "list_dir", "arguments": {"path": "."}}}
  - Run a command:
    {"tool_call": {"name": "run_command", "arguments": {"command": "go test ./..."}}}
  - Apply a patch:
    {"tool_call": {"name": "apply_patch", "arguments": {"file_path": "ai-team-data/design.md", "patch_content": "@@ -1 +1 @@\n-Old\n+New\n"}}}

Tool-call extraction is robust (the extractor accepts inline JSON and JSON inside code blocks, and the registry tolerates common casing variants). Still, keeping to the canonical structure avoids ambiguity.

### lastToolResponse

When a role calls a tool and it is executed by the system, the next role invocation receives the execution result in its input under two fields:

- `lastToolResponse`: the raw, structured result (map, list, string) when available.
- `lastToolResponse_json`: the JSON stringified representation (useful for template rendering inside prompts that expect a string).

Prompts can reference these in templates, for example:

    - Use `{{.lastToolResponse_json}}` to inspect the previous tool output as a string.
    - Use `{{if .lastToolResponse.error}}`...`{{end}}` to check for execution errors (when `lastToolResponse` contains an `error` key).

### Looping and `loop_condition`

Role chain steps can request iterative behavior by setting `loop: true` and a `loop_count`. To stop the loop early based on a runtime condition, set `loop_condition` to a Go template expression that evaluates to `true` or an equality expression after rendering.

Examples:

- Stop when the role returns a `write_file` tool call:
  loop: true
  loop_count: 10
  loop_condition: "{{.tool_call.name}} == 'write_file'"

Behavior:

- After each iteration the system will render `loop_condition` against the current context (which includes `tool_call` when present) and evaluate simple forms:
  - literal `true` / `false`
  - equality `{{.a}} == 'b'` and inequality `{{.a}} != 'b'`
- If the rendered condition evaluates to true, the loop stops early.
- For safety the evaluator accepts only the simple forms above. If you need more complex expressions (numeric comparisons, logical AND/OR), let me know and I can extend the evaluator or add a small expression parser.

### Prompt examples and guidance

- Keep prompts concise and instruct the model to produce only the JSON tool_call object.
- Provide canonical examples in the prompt to make format expectations explicit (we recommend including one or two minimal examples).
- Use `lastToolResponse_json` in prompts when you want the model to reason about the previous tool call output as text.

If you follow these rules, the extractor and tool executor will reliably detect and execute tool calls and pass results between roles in the chain.
