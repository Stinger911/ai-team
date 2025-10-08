package types

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
	Name        string `mapstructure:"Name"`
	Type        string `mapstructure:"Type"`
	Description string `mapstructure:"Description"`
}

// ConfigurableTool represents a tool defined in the configuration.
type ConfigurableTool struct {
	Name            string         `mapstructure:"Name"`
	Description     string         `mapstructure:"Description"`
	CommandTemplate string         `mapstructure:"CommandTemplate"`
	Arguments       []ToolArgument `mapstructure:"Arguments"`
}

// Role represents an AI role defined in the configuration.
type Role struct {
	Name   string `mapstructure:"Name"`
	Model  string `mapstructure:"Model"`
	Prompt string `mapstructure:"Prompt"`
}

// ChainRole represents a role within a chain.
type ChainRole struct {
	Name          string                 `mapstructure:"Name"`
	Input         map[string]interface{} `mapstructure:"Input"`
	OutputKey     string                 `mapstructure:"OutputKey"`
	Loop          bool                   `mapstructure:"Loop"`          // If true, loop this role
	LoopCount     int                    `mapstructure:"LoopCount"`     // Number of times to loop (if Loop is true)
	LoopCondition string                 `mapstructure:"LoopCondition"` // Optional: loop until a condition is met (Go template, evaluated after each iteration)
}

// RoleChain represents a chain of AI roles defined in the configuration.
type RoleChain struct {
	Name  string      `mapstructure:"Name"`
	Roles []ChainRole `mapstructure:"Roles"`
}

// RoleCallLogEntry represents a log entry for a single role call.
type RoleCallLogEntry struct {
	Timestamp string                 `json:"timestamp"`
	RoleName  string                 `json:"role_name"`
	Input     map[string]interface{} `json:"input"`
	Output    string                 `json:"output"`
	Error     string                 `json:"error,omitempty"`
}
