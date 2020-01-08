package cmd

import (
	"github.com/LINBIT/linstor-iscsi/pkg/rest"
	"github.com/spf13/cobra"
)

// serverCommand represents the server command
func serverCommand() *cobra.Command {
	var addr string

	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Starts a web server serving a REST API",
		Long: `Starts a web server serving a REST API
An up to date version of the REST-API documentation can be found here:
https://app.swaggerhub.com/apis-docs/Linstor/linstor-iscsi/

For example:
linstor-iscsi server --addr=":8080"`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			rest.ListenAndServe(addr)
		},
	}

	serverCmd.ResetCommands()
	serverCmd.Flags().StringVar(&addr, "addr", ":8080", "Host and port as defined by http.ListenAndServe()")
	serverCmd.DisableAutoGenTag = true

	return serverCmd
}
