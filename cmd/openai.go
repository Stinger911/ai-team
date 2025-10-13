package cmd

import (
	"fmt"
	"net/http"

	"ai-team/config"
	"ai-team/pkg/ai"

	"github.com/spf13/cobra"
)

var openaiModelKey string

var openaiCmd = &cobra.Command{
	Use:   "openai",
	Short: "Use the OpenAI model.",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			HandleError(err)
		}

		task, _ := cmd.Flags().GetString("task")
		modelKey := openaiModelKey
		if modelKey == "" {
			HandleError(fmt.Errorf("--model flag is required"))
		}
		modelCfg, ok := cfg.OpenAI.Models[modelKey]
		if !ok {
			HandleError(fmt.Errorf("model key '%s' not found in config for OpenAI", modelKey))
		}
		apiKey := modelCfg.Apikey
		if apiKey == "" {
			apiKey = cfg.OpenAI.Apikey
		}
		apiURL := modelCfg.Apiurl
		if apiURL == "" {
			apiURL = cfg.OpenAI.DefaultApiurl
		}
		client := &http.Client{}
		response, err := ai.CallOpenAI(client, task, apiURL, apiKey)
		if err != nil {
			HandleError(err)
		}
		fmt.Println("Response:", response)
	},
}

func init() {
	openaiCmd.Flags().String("task", "", "The task to perform.")
	openaiCmd.Flags().StringVar(&openaiModelKey, "model", "", "The OpenAI model key to use (from config).")
	openaiCmd.MarkFlagRequired("task")
	openaiCmd.MarkFlagRequired("model")
	rootCmd.AddCommand(openaiCmd)
}
