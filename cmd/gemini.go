package cmd

import (
	"fmt"
	"net/http"

	"ai-team/config"
	"ai-team/pkg/ai"

	"github.com/spf13/cobra"
)

var geminiModelKey string
var listGeminiModels bool

var geminiCmd = &cobra.Command{
	Use:   "gemini",
	Short: "Use the Gemini model.",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			HandleError(err)
		}

		client := &http.Client{}

		if listGeminiModels {
			apiKey := cfg.Gemini.Apikey
			apiURL := cfg.Gemini.Apiurl
			models, err := ai.ListGeminiModels(client, apiURL, apiKey)
			if err != nil {
				HandleError(err)
			}
			fmt.Println("Available Gemini Models:")
			for _, model := range models {
				fmt.Println("-", model)
			}
			return
		}

		task, _ := cmd.Flags().GetString("task")
		modelKey := geminiModelKey
		if modelKey == "" {
			HandleError(fmt.Errorf("--model flag is required"))
		}
		modelCfg, ok := cfg.Gemini.Models[modelKey]
		if !ok {
			HandleError(fmt.Errorf("model key '%s' not found in config for Gemini", modelKey))
		}
		apiKey := modelCfg.Apikey
		if apiKey == "" {
			apiKey = cfg.Gemini.Apikey
		}
		apiURL := modelCfg.Apiurl
		if apiURL == "" {
			apiURL = cfg.Gemini.Apiurl
		}
		response, err := ai.CallGemini(client, task, modelCfg.Model, apiURL, apiKey, cfg.Tools)
		if err != nil {
			HandleError(err)
		}
		fmt.Println("Response:", response)
	},
}

func init() {
	geminiCmd.Flags().StringVar(&geminiModelKey, "model", "", "The Gemini model key to use (from config).")
	geminiCmd.Flags().BoolVar(&listGeminiModels, "list-models", false, "List available Gemini models.")
	geminiCmd.Flags().String("task", "", "The task to perform.")
	geminiCmd.MarkFlagRequired("task")
	geminiCmd.MarkFlagRequired("model")
	rootCmd.AddCommand(geminiCmd)
}
