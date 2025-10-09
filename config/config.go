package config

import (
	"ai-team/pkg/errors"
	"ai-team/pkg/types" // Import types package

	"github.com/spf13/viper"
)

// Config holds the configuration for the application.
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
	LogFilePath string                     `mapstructure:"log_file_path"`
	LogStdout   bool                       `mapstructure:"log_stdout"`
	Tools       []types.ConfigurableTool   `mapstructure:"tools"`
	Roles       map[string]types.Role      `mapstructure:"roles"`
	Chains      map[string]types.RoleChain `mapstructure:"chains"`
}

type ModelConfig struct {
	Model       string  `mapstructure:"model"`
	Temperature float32 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Apikey      string  `mapstructure:"apikey"` // Model-specific API key
	Apiurl      string  `mapstructure:"apiurl"` // Model-specific API URL
	// ... other model parameters ...
}

// LoadConfig loads the configuration from a file.
func LoadConfig(configPath string) (Config, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.ai-team")
	}

	viper.AutomaticEnv() // Allow env var overrides
	viper.SetEnvPrefix("AI_TEAM")

	// Set sensible defaults
	viper.SetDefault("LogStdout", true)
	viper.SetDefault("Ollama.APIURL", "http://localhost:11434")
	// ...add more defaults as needed...

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, errors.New(errors.ErrCodeConfig, "failed to read config file: "+viper.ConfigFileUsed(), err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, errors.New(errors.ErrCodeConfig, "failed to unmarshal config: "+viper.ConfigFileUsed(), err)
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

// Validate checks for required config fields
func (c *Config) Validate() error {
	// if c.OpenAI.APIKey == "" && c.Gemini.APIKey == "" {
	// 	return errors.New(errors.ErrCodeConfig, "at least one API key must be set (OpenAI or Gemini)", nil)
	// }
	// ...add more checks as needed...
	return nil
}

func IsModelDefined(name string, cfg Config) bool {
	models := []string{"Ollama", "Gemini", "OpenAI"}
	for _, s := range models {
		if s == name {
			return true
		}
	}
	return false
}
