# AI Agent Team CLI

A command-line tool to manage a team of AI agents for programming.

## Description

This tool allows you to interact with different AI models (OpenAI, Gemini, and Ollama) from the command line. You can provide a task to the AI model and get a response.

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
