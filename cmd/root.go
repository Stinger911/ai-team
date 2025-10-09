package cmd

import (
	"ai-team/config"
	"ai-team/pkg/errors"
	"ai-team/pkg/roles"
	"ai-team/pkg/types"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runRoleCmd = &cobra.Command{
	Use:   "role <model-key>",
	Short: "Run a single AI role by model key (from config).",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modelKey := args[0]
		inputStr, _ := cmd.Flags().GetString("input")

		// Find the specified role by model key
		var targetRole types.Role
		foundRole := false
		for _, role := range cfg.Roles {
			if role.Model == modelKey {
				targetRole = role
				foundRole = true
				break
			}
		}
		if !foundRole {
			HandleError(errors.New(errors.ErrCodeRole, fmt.Sprintf("role with model key '%s' not found in config", modelKey), nil))
		}

		// Parse input string into a map
		roleInput := make(map[string]interface{})
		if inputStr != "" {
			parts := strings.Split(inputStr, "=")
			if len(parts) == 2 {
				roleInput[parts[0]] = parts[1]
			} else {
				HandleError(errors.New(errors.ErrCodeRole, "invalid input format. Expected key=value", nil))
			}
		}

		// Prefer flag over config
		logFilePath := logFileFlag
		if logFilePath == "" {
			logFilePath = cfg.LogFilePath
		}

		output, err := roles.ExecuteRole(
			targetRole,
			roleInput,
			cfg,
			logFilePath,
		)
		if err != nil {
			HandleError(err)
		}

		logrus.Info("Role execution complete. Output:")
		logrus.Infof("%s", output)
	},
}

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

		// Determine log file path (flag takes precedence)
		logFilePath := logFileFlag
		if logFilePath == "" {
			logFilePath = cfg.LogFilePath
		}
		if logFilePath != "" {
			// Open log file for append
			logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", logFilePath, err)
				os.Exit(1)
			}
			// Multi-writer: file + stdout if LogStdout is true
			if cfg.LogStdout {
				logrus.SetOutput(io.MultiWriter(os.Stdout, logFile))
			} else {
				logrus.SetOutput(logFile)
			}
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

		// Find the specified chain (map lookup)
		targetChain, foundChain := cfg.Chains[chainName]
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
			initialInput,
			cfg,
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
	runRoleCmd.Flags().String("input", "", "Input for the role (e.g., 'problem=design a new feature')")
	runRoleCmd.Flags().StringVar(&logFileFlag, "logFile", "", "Path to a file to log role calls (e.g., 'role_calls.log') (flag takes precedence over config)")
	rootCmd.AddCommand(runChainCmd)
	rootCmd.AddCommand(runRoleCmd)
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
