package types

import "time"

// OpenAIResponse represents the JSON response from the OpenAI API.
type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// GeminiRequest represents the request body for Gemini API.
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents a content block for Gemini API.
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of the content for Gemini API.
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the JSON response from the Gemini API.
type GeminiResponse struct {
	Candidates []struct {
		Content      GeminiContent `json:"content"`
		FinishReason string        `json:"finishReason"`
		// Tool call payloads may be present in Gemini tool call responses
		ToolCall *ToolCall `json:"toolCall,omitempty"`
	} `json:"candidates"`
}

// OllamaResponse represents the JSON response from the Ollama API.
type OllamaResponse struct {
	Response string `json:"response"`
}

type OllamaRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

// GeminiModelListResponse represents the JSON response from the Gemini models API.
type GeminiModelListResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// ToolCall represents a tool call requested by the AI.
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallRequest represents the overall structure of an AI's tool call request.
type ToolCallRequest struct {
	ToolCall ToolCall `json:"tool_call"`
}

// ToolArgument represents an argument for a configurable tool.
type ToolArgument struct {
	Name        string `mapstructure:"name"`
	Type        string `mapstructure:"type"`
	Description string `mapstructure:"description"`
}

// ConfigurableTool represents a tool defined in the configuration.
type ConfigurableTool struct {
	Name            string         `mapstructure:"name"`
	Description     string         `mapstructure:"description"`
	CommandTemplate string         `mapstructure:"command_template"`
	Arguments       []ToolArgument `mapstructure:"arguments"`
}

// Role represents an AI role defined in the configuration.
type Role struct {
	Provider string `mapstructure:"model_provider"` // e.g., "openai", "gemini", "ollama"
	Model    string `mapstructure:"model_name"`     // e.g., "gpt-4", "gemini-pro"
	Prompt   string `mapstructure:"prompt"`
}

// ChainRole represents a role within a chain.
type ChainRole struct {
	Name          string                 `mapstructure:"name"`
	Role          string                 `mapstructure:"role"`
	Input         map[string]interface{} `mapstructure:"input"`
	OutputKey     string                 `mapstructure:"output_key"`
	Loop          bool                   `mapstructure:"loop"`           // If true, loop this role
	LoopCount     int                    `mapstructure:"loop_count"`     // Number of times to loop (if Loop is true)
	LoopCondition string                 `mapstructure:"loop_condition"` // Optional: loop until a condition is met (Go template, evaluated after each iteration)
}

// RoleChain represents a chain of AI roles defined in the configuration.
type RoleChain struct {
	Steps []ChainRole `mapstructure:"steps"`
}

// RoleCallLogEntry represents a log entry for a single role call.
type RoleCallLogEntry struct {
	Timestamp string                 `json:"timestamp"`
	RoleName  string                 `json:"role_name"`
	Input     map[string]interface{} `json:"input"`
	Output    string                 `json:"output"`
	Error     string                 `json:"error,omitempty"`
}

// Transcript represents a session transcript.
type Transcript struct {
	Role      string    `json:"role"`
	StartedAt time.Time `json:"started_at"`
	Steps     []Step    `json:"steps"`
}

// Step represents a single step in a transcript.
type Step struct {
	LlmOutput string      `json:"llm_output"`
	ToolCall  *ToolCall   `json:"tool_call"`
	Approved  bool        `json:"approved"`
	Result    interface{} `json:"result"`
}

// Config represents the loaded YAML config (for reference, not used in main code)
type Config struct {
	OpenAI struct {
		DefaultApiurl string                 `mapstructure:"default_apiurl"`
		Apikey        string                 `mapstructure:"apikey"`
		Models        map[string]ModelConfig `mapstructure:"models"`
	} `mapstructure:"openai"`
	Gemini struct {
		Apikey string                 `mapstructure:"apikey"`
		Apiurl string                 `mapstructure:"apiurl"`
		Models map[string]ModelConfig `mapstructure:"models"`
	} `mapstructure:"gemini"`
	Ollama struct {
		Apiurl string                 `mapstructure:"apiurl"`
		Models map[string]ModelConfig `mapstructure:"models"`
	} `mapstructure:"ollama"`
	LogFilePath string               `mapstructure:"log_file_path"`
	LogStdout   bool                 `mapstructure:"log_stdout"`
	Tools       []ConfigurableTool   `mapstructure:"tools"`
	Roles       map[string]Role      `mapstructure:"roles"`
	Chains      map[string]RoleChain `mapstructure:"chains"`
}

// ModelConfig for reference (should match config.go)
type ModelConfig struct {
	Model       string  `mapstructure:"model"`
	Temperature float32 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Apikey      string  `mapstructure:"apikey"`
	Apiurl      string  `mapstructure:"apiurl"`
}
