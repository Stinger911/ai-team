package cmd

import (
	"fmt"
	"strings"

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
			localCfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				HandleError(err)
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			model, _ := cmd.Flags().GetString("model")
			maxIterations, _ := cmd.Flags().GetInt("max-iterations")
			contextFile, _ := cmd.Flags().GetString("context-file")
			transcriptPath, _ := cmd.Flags().GetString("transcript")
			yes, _ := cmd.Flags().GetBool("yes")
			editor, _ := cmd.Flags().GetString("editor")

			session := &roles.Session{
				DryRun:        dryRun,
				Model:         model,
				MaxIterations: maxIterations,
				ContextFile:   contextFile,
				UI:            &cli.DefaultUI{Editor: editor},
				Config:        &localCfg,
				TranscriptPath: transcriptPath,
				Yes:           yes,
			}

			roles.StartSession(session)
		} else {
			fmt.Printf("cfgFile in roleCmd: %s\n", cfgFile)
			localCfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				HandleError(err)
			}

			if len(args) < 1 {
				HandleError(fmt.Errorf("role name is required for non-interactive mode"))
				return
			}
			roleName := args[0]

			role, ok := localCfg.Roles[roleName]
			if !ok {
				HandleError(fmt.Errorf("role not found: %s", roleName))
				return
			}

			inputs := make(map[string]interface{})
			for _, input := range args[1:] {
				parts := strings.SplitN(input, "=", 2)
				if len(parts) != 2 {
					HandleError(fmt.Errorf("invalid input format: %s", input))
					return
				}
				inputs[parts[0]] = parts[1]
			}

			output, err := roles.ExecuteRole(role, inputs, &localCfg, "")
			if err != nil {
				HandleError(err)
			}
			fmt.Println(output)
		}
	},
}

func init() {
	roleCmd.Flags().Bool("interactive", false, "Enable interactive mode.")
	roleCmd.Flags().Bool("dry-run", false, "Enable dry-run mode.")
	roleCmd.Flags().String("model", "", "The model to use.")
	roleCmd.Flags().Int("max-iterations", 5, "The maximum number of iterations.")
	roleCmd.Flags().String("context-file", "", "The path to a context file.")
	roleCmd.Flags().String("transcript", "", "Path to a file to save the session transcript.")
	roleCmd.Flags().Bool("yes", false, "Automatically approve all tool calls without prompting.")
	roleCmd.Flags().String("editor", "", "Specify the editor to use for editing tool calls.")
	rootCmd.AddCommand(roleCmd)

	// Add completion for role names
	roleCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var roleNames []string
		for name := range cfg.Roles {
			if strings.HasPrefix(name, toComplete) {
				roleNames = append(roleNames, name)
			}
		}
		return roleNames, cobra.ShellCompDirectiveNoFileComp
	}
}