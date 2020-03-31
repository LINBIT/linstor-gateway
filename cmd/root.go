package cmd

import (
	"fmt"
	"os"
	"path"

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
	if os.Args == nil {
		log.Fatal("Program started with null-pointer argument list")
	}
	if len(os.Args) < 1 {
		log.Fatal("Program started with a zero-length argument list")
	}

	var cmd *cobra.Command
	pgmName := path.Base(os.Args[0])
	if (pgmName == "linstor-iscsi") {
		cmd = iscsiCommands()
	} else if (pgmName == "linstor-nfs") {
		cmd = nfsCommands()
	} else {
		execHint(pgmName)
		log.Fatal("Could not determine whether to run in iSCSI or NFS mode - aborting")
	}
        return cmd
}
		
func iscsiCommands() *cobra.Command {
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
	rootCmd.AddCommand(createISCSICommand())
	rootCmd.AddCommand(deleteISCSICommand())
	rootCmd.AddCommand(docsCommand(rootCmd))
	rootCmd.AddCommand(listISCSICommand())
	rootCmd.AddCommand(serverCommand())
	rootCmd.AddCommand(startCommand())
	rootCmd.AddCommand(statusISCSICommand())
	rootCmd.AddCommand(stopCommand())

	return rootCmd

}

func nfsCommands() *cobra.Command {
	var loglevel string

	var rootCmd = &cobra.Command{
		Use:     "linstor-nfs",
		Version: "0.1.0",
		Short:   "Manages Highly-Available NFS exports",
		Long: `linstor-nfs manages higly available NFS exports by leveraging on linstor
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

	rootCmd.AddCommand(completionCommand(rootCmd))
	rootCmd.AddCommand(corosyncCommand())
	rootCmd.AddCommand(createNFSCommand())
	rootCmd.AddCommand(deleteNFSCommand())
	rootCmd.AddCommand(docsCommand(rootCmd))
	rootCmd.AddCommand(listNFSCommand())
	rootCmd.AddCommand(serverCommand())
	rootCmd.AddCommand(startCommand())
	rootCmd.AddCommand(statusNFSCommand())
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

func execHint(currentName string) {
	fmt.Printf(
		"This program must be executed as either\n" +
		" - linstor-iscsi\n" +
		"or\n" +
		" - linstor-nfs\n" +
		"Rename the program executable file accordingly, or create symbolic links\n" +
		"using those names and point them at the " + currentName + " executable.\n",
	)
}
