package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func completionCommand(dst *cobra.Command) *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion script",
		Long: `To load completion run

. <(linstor-iscsi completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(linstor-iscsi completion)`,
		Run: func(cmd *cobra.Command, args []string) {
			dst.GenBashCompletion(os.Stdout)
		},
	}

	completionCmd.ResetCommands()
	completionCmd.DisableAutoGenTag = true

	return completionCmd
}
