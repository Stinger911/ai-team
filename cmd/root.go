package cmd

import (
	"ai-team/config"
	"ai-team/pkg/errors"
	"ai-team/pkg/roles"
	"ai-team/pkg/types"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfgFile string
var logFileFlag string
var cfg config.Config

var rootCmd = &cobra.Command{
	Use:   "ai-team",
	Short: "A command-line tool to manage a team of AI agents for programming.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			HandleError(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
}

var runChainCmd = &cobra.Command{
	Use:   "run-chain <chain-name>",
	Short: "Run a defined AI role chain.",
	Args:  cobra.ExactArgs(1), // Expect exactly one argument: the chain name
	Run: func(cmd *cobra.Command, args []string) {
		chainName := args[0]
		inputStr, _ := cmd.Flags().GetString("input")

		// Find the specified chain
		var targetChain types.RoleChain
		foundChain := false
		for _, chain := range cfg.Chains {
			if chain.Name == chainName {
				targetChain = chain
				foundChain = true
				break
			}
		}

		if !foundChain {
			HandleError(errors.New(errors.ErrCodeRole, fmt.Sprintf("role chain '%s' not found in config", chainName), nil))
		}

		// Parse input string into a map
		initialInput := make(map[string]interface{})
		if inputStr != "" {
			parts := strings.Split(inputStr, "=")
			if len(parts) == 2 {
				initialInput[parts[0]] = parts[1]
			} else {
				HandleError(errors.New(errors.ErrCodeRole, "invalid input format. Expected key=value", nil))
			}
		}

		// Prefer flag over config
		logFilePath := logFileFlag
		if logFilePath == "" {
			logFilePath = cfg.LogFilePath
		}

		result, err := roles.ExecuteChain(
			targetChain,
			cfg.Roles,
			initialInput,
			cfg.Gemini.APIURL,
			cfg.Gemini.APIKey,
			cfg.Tools,
			logFilePath, // Pass logFilePath
		)
		if err != nil {
			HandleError(err)
		}

		logrus.Info("Chain execution complete. Final context:")
		for k, v := range result {
			logrus.Infof("  %s: %v", k, v)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ai-team.yaml)")
	runChainCmd.Flags().String("input", "", "Initial input for the chain (e.g., 'problem=design a new feature')")
	runChainCmd.Flags().StringVar(&logFileFlag, "logFile", "", "Path to a file to log role calls (e.g., 'role_calls.log') (flag takes precedence over config)")
	rootCmd.AddCommand(runChainCmd)
}

func ExecuteCmd() { // Renamed to ExecuteCmd
	if err := rootCmd.Execute(); err != nil {
		HandleError(err)
	}
}

// HandleError handles errors by printing them to stderr and exiting.
func HandleError(err error) {
	if e, ok := err.(*errors.Error); ok {
		logrus.Errorf("Error: %s (code: %d)", e.Message, e.Code)
		if e.Err != nil {
			logrus.Errorf("  Caused by: %v", e.Err)
		}
	} else {
		logrus.Errorf("An unexpected error occurred: %v", err)
	}
	// Still exit after logging
	os.Exit(1)
}
