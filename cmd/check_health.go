package cmd

import (
	"fmt"
	"github.com/LINBIT/linstor-gateway/pkg/healthcheck"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func checkHealthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check-health",
		Short: "Check if all requirements and dependencies are met on the current system",
		Run: func(cmd *cobra.Command, args []string) {
			err := healthcheck.CheckRequirements(controllers)
			if err != nil {
				fmt.Println()
				log.Fatalf("Health check failed: %v", err)
			}
		},
	}
}
