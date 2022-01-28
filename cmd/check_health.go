package cmd

import (
	"fmt"
	"github.com/LINBIT/linstor-gateway/pkg/healthcheck"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func checkHealthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-health",
		Short: "Check if all requirements and dependencies are met on the current system",
		Run: func(cmd *cobra.Command, args []string) {
			controllers := viper.GetStringSlice("linstor.controllers")
			err := healthcheck.CheckRequirements(controllers)
			if err != nil {
				fmt.Println()
				log.Fatalf("Health check failed: %v", err)
			}
		},
	}
	cmd.Flags().StringSlice("controllers", nil, "List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)")
	viper.BindPFlag("linstor.controllers", cmd.Flags().Lookup("controllers"))

	return cmd
}
