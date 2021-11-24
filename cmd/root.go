package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// (potentially) injected by makefile
var (
	version   string
	builddate string
	githash   string
)

var (
	cfgFile     string
	controllers []string
	loglevel    string
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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			controllers = viper.GetStringSlice("linstor.controllers")
			level, err := log.ParseLevel(loglevel)
			if err != nil {
				log.Fatal(err)
			}
			log.SetLevel(level)
		},
	}
	rootCmd.AddCommand(iscsiCommands())
	rootCmd.AddCommand(nfsCommands())
	rootCmd.AddCommand(nvmeCommands())
	rootCmd.AddCommand(serverCommand())
	rootCmd.AddCommand(versionCommand())
	rootCmd.AddCommand(completionCommand(rootCmd))
	rootCmd.AddCommand(docsCommand(rootCmd))
	rootCmd.AddCommand(checkHealthCommand())
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/linstor-gateway/linstor-gateway.toml", "Config file to load")
	rootCmd.PersistentFlags().StringSlice("controllers", nil, "List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)")
	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", log.InfoLevel.String(), "Set the log level (as defined by logrus)")
	viper.BindPFlag("linstor.controllers", rootCmd.PersistentFlags().Lookup("controllers"))
	return rootCmd
}

func initConfig() {
	viper.SetDefault("linstor.controllers", "")
	viper.SetConfigType("toml")
	viper.SetConfigFile(cfgFile)
	viper.ReadInConfig()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.OnInitialize(initConfig)
	rootCmd := rootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
