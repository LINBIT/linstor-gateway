package cmd

import (
	"fmt"

	"github.com/LINBIT/linstor-gateway/client"
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
linstor-gateway server --addr=":8337"`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			controllers := viper.GetStringSlice("linstor.controllers")
			corsOrigins := viper.GetStringSlice("server.cors_allowed_origins")

			// Always add controller-based origins if controllers are configured
			if len(controllers) > 0 {
				for _, ctrl := range controllers {
					corsOrigins = append(corsOrigins, fmt.Sprintf("http://%s:3370", ctrl))
					corsOrigins = append(corsOrigins, fmt.Sprintf("https://%s:3370", ctrl))
				}
			}

			rest.ListenAndServe(addr, controllers, corsOrigins)
		},
	}

	serverCmd.ResetCommands()
	defaultAddr := fmt.Sprintf(":%d", client.DefaultPort)
	serverCmd.Flags().StringVar(&addr, "addr", defaultAddr, "Host and port as defined by http.ListenAndServe()")
	serverCmd.Flags().StringSlice("controllers", nil, "List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)")
	viper.BindPFlag("linstor.controllers", serverCmd.Flags().Lookup("controllers"))
	serverCmd.DisableAutoGenTag = true

	return serverCmd
}
