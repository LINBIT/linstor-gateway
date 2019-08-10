package cmd

import (
	"fmt"
	"log"

	"github.com/LINBIT/linstor-remote-storage/iscsi"
	term "github.com/LINBIT/linstor-remote-storage/termcontrol"
	"github.com/spf13/cobra"
)

// probeCmd represents the probe command
var probeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Probes an iSCSI starget",
	Long: `Triggers Pacemaker to probe the resoruce primitives of this iSCSI target.
That means Pacemaker will run the status operation on the nodes where the
resource can run.
This makes sure that Pacemakers view of the world is updated to the state
of the world.

For example:
./linstor-iscsi probe --iqn=iqn.2019-08.com.libit:example --lun=0`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		rscStateMap, err := iscsi.ProbeResource(iqn, uint8(lun))
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
