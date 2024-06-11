package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/LINBIT/linstor-gateway/pkg/version"
)

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information of LINSTOR Gateway",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("LINSTOR Gateway version %s\n", version.Version)
			fmt.Printf("Built at %s\n", version.BuildDate)
			fmt.Printf("Version control hash: %s\n", version.GitCommit)
		},
	}
}
