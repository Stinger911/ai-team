package cmd

import (
	"fmt"
	"net/http"

	"ai-team/pkg/ai"
	"github.com/spf13/cobra"
)

var geminiCmd = &cobra.Command{
	Use:   "gemini",
	Short: "Use the Gemini model.",
	Run: func(cmd *cobra.Command, args []string) {
		task, _ := cmd.Flags().GetString("task")
		client := &http.Client{}
		response, err := ai.CallGemini(client, task, cfg.Gemini.APIURL, cfg.Gemini.APIKey)
		if err != nil {
			handleError(err)
		}
		fmt.Println("Response:", response)
	},
}

func init() {
	geminiCmd.Flags().String("task", "", "The task to perform.")
	geminiCmd.MarkFlagRequired("task")
	rootCmd.AddCommand(geminiCmd)
}
