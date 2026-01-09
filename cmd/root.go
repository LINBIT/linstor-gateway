package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/LINBIT/linstor-gateway/client"
	"github.com/LINBIT/linstor-gateway/pkg/version"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	loglevel string
	host     string
	cli      *client.Client
)

func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}

func parseBaseURL(urlString string) (*url.URL, error) {
	// Check scheme
	urlSplit := strings.Split(urlString, "://")

	if len(urlSplit) == 1 {
		if urlSplit[0] == "" {
			urlSplit[0] = client.DefaultHost
		}
		urlSplit = []string{client.DefaultScheme, urlSplit[0]}
	}

	if len(urlSplit) != 2 {
		return nil, fmt.Errorf("URL with multiple scheme separators. parts: %v", urlSplit)
	}
	scheme, endpoint := urlSplit[0], urlSplit[1]

	// Check port
	endpointSplit := strings.Split(endpoint, ":")
	if len(endpointSplit) == 1 {
		endpointSplit = []string{endpointSplit[0], strconv.Itoa(client.DefaultPort)}
	}
	if len(endpointSplit) != 2 {
		return nil, fmt.Errorf("URL with multiple port separators. parts: %v", endpointSplit)
	}
	host, port := endpointSplit[0], endpointSplit[1]

	return url.Parse(fmt.Sprintf("%s://%s:%s", scheme, host, port))
}

// rootCommand represents the base command when called without any subcommands
func rootCommand() *cobra.Command {
	if len(os.Args) < 1 {
		log.Fatal("Program started with a zero-length argument list")
	}

	rootCmd := &cobra.Command{
		Use:           "linstor-gateway",
		Version:       version.Version,
		Short:         "Manage linstor-gateway targets and exports",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			level, err := log.ParseLevel(loglevel)
			if err != nil {
				return err
			}
			log.SetLevel(level)

			base, err := parseBaseURL(host)
			if err != nil {
				return err
			}
			cli, err = client.NewClient(
				client.BaseURL(base),
				client.Log(log.StandardLogger()),
				client.UserAgent(version.UserAgent()),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to LINSTOR Gateway server: %w", err)
			}
			return nil
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
	defaultConnect := fmt.Sprintf("%s://%s:%d", client.DefaultScheme, client.DefaultHost, client.DefaultPort)
	rootCmd.PersistentFlags().StringVarP(&host, "connect", "c", defaultConnect, "LINSTOR Gateway server to connect to")
	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", log.InfoLevel.String(), "Set the log level (as defined by logrus)")
	return rootCmd
}

func initConfig() {
	viper.SetDefault("linstor.controllers", "")
	viper.SetDefault("server.cors_allowed_origins", []string{})
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
		log.Error(err)
		os.Exit(1)
	}
}
