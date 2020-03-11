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
	var loglevel string

	var rootCmd = &cobra.Command{
		Use:     "linstor-iscsi",
		Version: version,
		Short:   "Manages Highly-Available iSCSI targets",
		Long: `linstor-iscsi manages higly available iSCSI targets by leveraging on linstor
and Pacemaker. Setting linstor including storage pools and resource groups
as well as Corosync and Pacemaker's properties a prerequisite to use this tool.`,
		Args: cobra.NoArgs,
		// We could have our custom flag types, but this is really simple enough...
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// TODO: regex most likely needs review
			level, err := log.ParseLevel(loglevel)
			if err != nil {
				log.Fatal(err)
			}
			log.SetLevel(level)
		},
	}

	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", log.InfoLevel.String(), "Set the log level (as defined by logrus)")
	rootCmd.DisableAutoGenTag = true

	rootCmd.AddCommand(versionCommand())
	rootCmd.AddCommand(completionCommand(rootCmd))
	rootCmd.AddCommand(corosyncCommand())
	rootCmd.AddCommand(createIscsiCommand())
	rootCmd.AddCommand(deleteIscsiCommand())
	rootCmd.AddCommand(createNfsCommand())
	rootCmd.AddCommand(deleteNfsCommand())
	rootCmd.AddCommand(docsCommand(rootCmd))
	rootCmd.AddCommand(listCommand())
	rootCmd.AddCommand(serverCommand())
	rootCmd.AddCommand(startCommand())
	rootCmd.AddCommand(statusCommand())
	rootCmd.AddCommand(stopCommand())

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
