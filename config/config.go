package config

import (
	"ai-team/pkg/errors"
	"ai-team/pkg/types" // Import types package
	"fmt"

	"github.com/sirupsen/logrus"
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
	if c.OpenAI.Apikey == "" && c.Gemini.Apikey == "" && c.Ollama.Apiurl == "" {
		return errors.New(errors.ErrCodeConfig, "at least one API configuration must be set (OpenAI, Gemini, or Ollama)", nil)
	}

	// Validate OpenAI models
	for name, m := range c.OpenAI.Models {
		if m.Model == "" {
			return errors.New(errors.ErrCodeConfig, fmt.Sprintf("OpenAI model '%s' missing 'model' field", name), nil)
		}
		if m.MaxTokens <= 0 {
			return errors.New(errors.ErrCodeConfig, fmt.Sprintf("OpenAI model '%s' has invalid max_tokens", name), nil)
		}
	}
	// Validate Gemini models
	for name, m := range c.Gemini.Models {
		if m.Model == "" {
			return errors.New(errors.ErrCodeConfig, fmt.Sprintf("Gemini model '%s' missing 'model' field", name), nil)
		}
		if m.MaxTokens <= 0 {
			return errors.New(errors.ErrCodeConfig, fmt.Sprintf("Gemini model '%s' has invalid max_tokens", name), nil)
		}
	}
	// Validate Ollama models
	for name, m := range c.Ollama.Models {
		if m.Model == "" {
			return errors.New(errors.ErrCodeConfig, fmt.Sprintf("Ollama model '%s' missing 'model' field", name), nil)
		}
	}

	for _, tool := range c.Tools {
		logrus.Debugf("Validating tool: %+v", tool)
		if tool.Name == "" {
			return errors.New(errors.ErrCodeConfig, "tool must have a Name", nil)
		}
		if tool.CommandTemplate == "" {
			return errors.New(errors.ErrCodeConfig, fmt.Sprintf("tool '%s' must have a CommandTemplate", tool.Name), nil)
		}
		for _, arg := range tool.Arguments {
			if arg.Name == "" || arg.Type == "" {
				return errors.New(errors.ErrCodeConfig, fmt.Sprintf("tool '%s' has argument with missing name or type", tool.Name), nil)
			}
		}
	}

	for name, role := range c.Roles {
		if role.Model == "" {
			return errors.New(errors.ErrCodeConfig, fmt.Sprintf("role '%s' must have a Model", name), nil)
		}
	}

	// Validate chains: referenced roles must exist
	for cname, chain := range c.Chains {
		for _, step := range chain.Steps {
			if step.Role != "" {
				if _, ok := c.Roles[step.Role]; !ok {
					return errors.New(errors.ErrCodeConfig, fmt.Sprintf("chain '%s' references undefined role '%s'", cname, step.Role), nil)
				}
			}
		}
	}

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
