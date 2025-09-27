package cmd

import (
	"fmt"
	"os"

	"ai-team/config"
	"ai-team/pkg/errors"
	"github.com/spf13/cobra"
)

var cfg config.Config

var rootCmd = &cobra.Command{
	Use:   "ai-team",
	Short: "A command-line tool to manage a team of AI agents for programming.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			handleError(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Do nothing
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	if e, ok := err.(*errors.Error); ok {
		fmt.Fprintf(os.Stderr, "Error: %s (code: %d)\n", e.Message, e.Code)
		if e.Err != nil {
			fmt.Fprintf(os.Stderr, "  Caused by: %v\n", e.Err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "An unexpected error occurred: %v\n", err)
	}
	os.Exit(1)
}
