package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// (potentially) injected by makefile
var (
	version   string
	builddate string
	githash   string
)

// rootCommand represents the base command when called without any subcommands
func rootCommand() *cobra.Command {
	if len(os.Args) < 1 {
		log.Fatal("Program started with a zero-length argument list")
	}

	rootCmd := &cobra.Command{
		Use:     "linstor-gateway",
		Version: version,
		Short:   "Manage linstor-gateway targets and exports",
		Args:    cobra.NoArgs,
	}
	rootCmd.AddCommand(iscsiCommands())
	rootCmd.AddCommand(nfsCommands())
	rootCmd.AddCommand(nvmeCommands())
	rootCmd.AddCommand(serverCommand())
	rootCmd.AddCommand(versionCommand())
	rootCmd.AddCommand(completionCommand(rootCmd))
	rootCmd.AddCommand(docsCommand(rootCmd))
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := rootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
