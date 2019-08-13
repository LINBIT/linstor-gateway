package cmd

import (
	"github.com/LINBIT/linstor-remote-storage/rest"
	"github.com/spf13/cobra"
)

var addr string

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts a web server serving a REST API",
	Long: `Starts a web server serving a REST API

For example:
linstor-iscsi server --addr=":8080"`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		rest.ListenAndServe(addr)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVar(&addr, "addr", ":8080", "Host and port as defined by http.ListenAndServe()")
}
