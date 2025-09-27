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

* `openai`: Use the OpenAI model.
* `gemini`: Use the Gemini model.
* `ollama`: Use the Ollama model.

### Example

```bash
./ai-team openai --task "write a hello world program in Go"
```

## Configuration

The tool uses a `config.yaml` file to configure the API keys and URLs for the different AI models. You can find an example `config.yaml` file in the repository.

## Development

### Running tests

```bash
make test
```

### Building the binary

```bash
make build
```

### Cleaning up

```bash
make clean
```
