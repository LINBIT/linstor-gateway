package cmd

import (
	"fmt"
	"github.com/LINBIT/linstor-gateway/pkg/healthcheck"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func checkHealthCommand() *cobra.Command {
	var mode string
	cmd := &cobra.Command{
		Use:   "check-health",
		Short: "Check if all requirements and dependencies are met on the current system",
		Long: `Check if all requirements and dependencies are met on the current system.

The "mode" argument can be used to specify the type of node this
system is intended to be used as.

An "agent" node is responsible for actually running the highly available storage
endpoint (for example, an iSCSI target). Thus, the health check ensures the
current system is correctly configured to host the different kinds of storage
protocols supported by LINSTOR Gateway.

A "server" node acts as a relay between the client and the LINSTOR controller.
The only requirement is that the LINSTOR controller can be reached.

A "client" node interacts with the LINSTOR Gateway API. Its only requirement is
that a LINSTOR Gateway server can be reached.
`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			controllers := viper.GetStringSlice("linstor.controllers")
			err := healthcheck.CheckRequirements(mode, controllers, cli)
			if err != nil {
				fmt.Println()
				log.Fatalf("Health check failed: %v", err)
			}
		},
	}
	cmd.Flags().StringSlice("controllers", nil, "List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)")
	viper.BindPFlag("linstor.controllers", cmd.Flags().Lookup("controllers"))
	cmd.Flags().StringVarP(&mode, "mode", "m", "agent", `Which type of node to check requirements for. Can be "agent", "server", or "client"`)

	return cmd
}
