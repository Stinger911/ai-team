package config

import (
	"ai-team/pkg/errors"
	"ai-team/pkg/types" // Import types package

	"github.com/spf13/viper"
)

// Config holds the configuration for the application.
type Config struct {
	OpenAI struct {
		APIKey string `mapstructure:"APIKey"`
		APIURL string `mapstructure:"APIURL"`
	}
	Gemini struct {
		APIKey string `mapstructure:"APIKey"`
		APIURL string `mapstructure:"APIURL"`
		Model  string `mapstructure:"Model"`
	}
	Ollama struct {
		APIURL string `mapstructure:"APIURL"`
	}
	LogFilePath string                   `mapstructure:"LogFilePath"`
	LogStdout   bool                     `mapstructure:"LogStdout"`
	Tools       []types.ConfigurableTool `mapstructure:"Tools"`
	Roles       []types.Role             `mapstructure:"Roles"`
	Chains      []types.RoleChain        `mapstructure:"Chains"`
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

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, errors.New(errors.ErrCodeConfig, "failed to read config file", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, errors.New(errors.ErrCodeConfig, "failed to unmarshal config", err)
	}
	// Default LogStdout to true if not set
	if !viper.IsSet("LogStdout") {
		config.LogStdout = true
	}
	return config, nil
}
