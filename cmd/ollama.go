package cmd

import (
	"fmt"
	"net/http"

	"ai-team/config"
	"ai-team/pkg/ai"

	"github.com/spf13/cobra"
)

var ollamaModelKey string

var ollamaCmd = &cobra.Command{
	Use:   "ollama",
	Short: "Use the Ollama model.",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			HandleError(err)
		}

		task, _ := cmd.Flags().GetString("task")
		modelKey := ollamaModelKey
		if modelKey == "" {
			HandleError(fmt.Errorf("--model flag is required"))
		}
		modelCfg, ok := cfg.Ollama.Models[modelKey]
		if !ok {
			HandleError(fmt.Errorf("model key '%s' not found in config for Ollama", modelKey))
		}
		apiURL := modelCfg.Apiurl
		if apiURL == "" {
			apiURL = cfg.Ollama.Apiurl
		}
		client := &http.Client{}
		response, err := ai.CallOllama(client, task, apiURL, modelCfg.Model, cfg.Tools)
		if err != nil {
			HandleError(err)
		}
		fmt.Println("Response:", response)
	},
}

func init() {
	ollamaCmd.Flags().String("task", "", "The task to perform.")
	ollamaCmd.Flags().StringVar(&ollamaModelKey, "model", "", "The Ollama model key to use (from config).")
	ollamaCmd.MarkFlagRequired("task")
	ollamaCmd.MarkFlagRequired("model")
	rootCmd.AddCommand(ollamaCmd)
}
