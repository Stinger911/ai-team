package cmd

import (
	"fmt"
	"net/http"

	"ai-team/pkg/ai"
	"github.com/spf13/cobra"
)

var ollamaCmd = &cobra.Command{
	Use:   "ollama",
	Short: "Use the Ollama model.",
	Run: func(cmd *cobra.Command, args []string) {
		task, _ := cmd.Flags().GetString("task")
		client := &http.Client{}
		response, err := ai.CallOllama(client, task, cfg.Ollama.APIURL)
		if err != nil {
			handleError(err)
		}
		fmt.Println("Response:", response)
	},
}

func init() {
	ollamaCmd.Flags().String("task", "", "The task to perform.")
	ollamaCmd.MarkFlagRequired("task")
	rootCmd.AddCommand(ollamaCmd)
}