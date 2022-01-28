package cmd

import (
	"github.com/LINBIT/linstor-gateway/pkg/rest"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func serverCommand() *cobra.Command {
	var addr string

	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Starts a web server serving a REST API",
		Long: `Starts a web server serving a REST API
An up to date version of the REST-API documentation can be found here:
https://app.swaggerhub.com/apis-docs/Linstor/linstor-gateway

For example:
linstor-gateway server --addr=":8080"`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			controllers := viper.GetStringSlice("linstor.controllers")
			rest.ListenAndServe(addr, controllers)
		},
	}

	serverCmd.ResetCommands()
	serverCmd.Flags().StringVar(&addr, "addr", ":8080", "Host and port as defined by http.ListenAndServe()")
	serverCmd.Flags().StringSlice("controllers", nil, "List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)")
	viper.BindPFlag("linstor.controllers", serverCmd.Flags().Lookup("controllers"))
	serverCmd.DisableAutoGenTag = true

	return serverCmd
}
