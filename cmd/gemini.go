package cmd

import (
	"fmt"
	"net/http"

	"ai-team/pkg/ai"
	"github.com/spf13/cobra"
)

var listGeminiModels bool

var geminiCmd = &cobra.Command{
	Use:   "gemini",
	Short: "Use the Gemini model.",
	Run: func(cmd *cobra.Command, args []string) {
		client := &http.Client{}

		if listGeminiModels {
			models, err := ai.ListGeminiModels(client, cfg.Gemini.APIURL, cfg.Gemini.APIKey)
			if err != nil {
				HandleError(err) // Use HandleError
			}
			fmt.Println("Available Gemini Models:")
			for _, model := range models {
				fmt.Println("-", model)
			}
			return
		}

		task, _ := cmd.Flags().GetString("task")
		if task == "" {
			cmd.Help()
			return
		}

		model := cfg.Gemini.Model
		if model == "" {
			// Default model if not specified in config
			model = "gemini-pro" // Or another suitable default
		}

		response, err := ai.CallGemini(client, task, model, cfg.Gemini.APIURL, cfg.Gemini.APIKey, cfg.Tools)
		if err != nil {
			HandleError(err) // Use HandleError
		}
		fmt.Println("Response:", response)
	},
}

func init() {
	geminiCmd.Flags().StringVar(&cfg.Gemini.Model, "model", "", "The Gemini model to use (e.g., gemini-pro).")
	geminiCmd.Flags().BoolVar(&listGeminiModels, "list-models", false, "List available Gemini models.")
	geminiCmd.Flags().String("task", "", "The task to perform.")
	rootCmd.AddCommand(geminiCmd)
}
