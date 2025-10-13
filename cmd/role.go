package cmd

import (
	"fmt"

	"ai-team/config"
	"ai-team/pkg/cli"
	"ai-team/pkg/roles"

	"github.com/spf13/cobra"
)

var roleCmd = &cobra.Command{
	Use:   "role [role] [inputs...]",
	Short: "Execute a role.",
	Run: func(cmd *cobra.Command, args []string) {
		interactive, _ := cmd.Flags().GetBool("interactive")

		if interactive {
			cfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				HandleError(err)
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			model, _ := cmd.Flags().GetString("model")
			maxIterations, _ := cmd.Flags().GetInt("max-iterations")
			contextFile, _ := cmd.Flags().GetString("context-file")

			session := &roles.Session{
				DryRun:        dryRun,
				Model:         model,
				MaxIterations: maxIterations,
				ContextFile:   contextFile,
				UI:            &cli.DefaultUI{},
				Config:        &cfg,
			}

			roles.StartSession(session)
		} else {
			// TODO: Implement the non-interactive mode.
			fmt.Println("Non-interactive mode is not yet implemented.")
		}
	},
}

func init() {
	roleCmd.Flags().Bool("interactive", false, "Enable interactive mode.")
	roleCmd.Flags().Bool("dry-run", false, "Enable dry-run mode.")
	roleCmd.Flags().String("model", "", "The model to use.")
	roleCmd.Flags().Int("max-iterations", 5, "The maximum number of iterations.")
	roleCmd.Flags().String("context-file", "", "The path to a context file.")
	rootCmd.AddCommand(roleCmd)
}