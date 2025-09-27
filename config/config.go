package config

import (
	"ai-team/pkg/errors"
	"github.com/spf13/viper"
)

// Config holds the configuration for the application.
type Config struct {
	OpenAI struct {
		APIKey string `mapstructure:"api_key"`
		APIURL string `mapstructure:"api_url"`
	}
	Gemini struct {
		APIKey string `mapstructure:"api_key"`
		APIURL string `mapstructure:"api_url"`
	}
	Ollama struct {
		APIURL string `mapstructure:"api_url"`
	}
}

// LoadConfig loads the configuration from a file.
func LoadConfig() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, errors.New(errors.ErrCodeConfig, "failed to read config file", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, errors.New(errors.ErrCodeConfig, "failed to unmarshal config", err)
	}

	return config, nil
}