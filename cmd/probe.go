package cmd

import (
	"fmt"
	"log"

	"github.com/LINBIT/linstor-remote-storage/application"
	term "github.com/LINBIT/linstor-remote-storage/termcontrol"
	"github.com/spf13/cobra"
)

// probeCmd represents the probe command
var probeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Probes an iSCSI starget",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		rscStateMap, _, err := application.ProbeResource(iqn, uint8(lun))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Current state of CRM resources\niSCSI resource %s, logical unit #%d:\n", iqn, uint8(lun))
		for rscName, runState := range *rscStateMap {
			label := term.COLOR_YELLOW + "Unknown" + term.COLOR_RESET
			if runState.HaveState {
				if runState.Running {
					label = term.COLOR_GREEN + "Running" + term.COLOR_RESET
				} else {
					label = term.COLOR_RED + "Stopped" + term.COLOR_RESET
				}
			}
			fmt.Printf("    %-40s %s\n", rscName, label)
		}
	},
}

func init() {
	rootCmd.AddCommand(probeCmd)
}
