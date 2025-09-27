package cmd

import (
	"fmt"
	"net/http"

	"ai-team/pkg/ai"
	"github.com/spf13/cobra"
)

var openaiCmd = &cobra.Command{
	Use:   "openai",
	Short: "Use the OpenAI model.",
	Run: func(cmd *cobra.Command, args []string) {
		task, _ := cmd.Flags().GetString("task")
		client := &http.Client{}
		response, err := ai.CallOpenAI(client, task, cfg.OpenAI.APIURL, cfg.OpenAI.APIKey)
		if err != nil {
			handleError(err)
		}
		fmt.Println("Response:", response)
	},
}

func init() {
	openaiCmd.Flags().String("task", "", "The task to perform.")
	openaiCmd.MarkFlagRequired("task")
	rootCmd.AddCommand(openaiCmd)
}
