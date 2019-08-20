package cmd

import (
	"fmt"

	"github.com/LINBIT/linstor-remote-storage/crmcontrol"
	"github.com/LINBIT/linstor-remote-storage/iscsi"
	"github.com/LINBIT/linstor-remote-storage/linstorcontrol"
	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// probeCmd represents the probe command
var probeCmd = &cobra.Command{
	Use:   "probe",
	Short: "Probes an iSCSI target",
	Long: `Triggers Pacemaker to probe the resoruce primitives of this iSCSI target.
That means Pacemaker will run the status operation on the nodes where the
resource can run.
This makes sure that Pacemakers view of the world is updated to the state
of the world.

For example:
linstor-iscsi probe --iqn=iqn.2019-08.com.linbit:example --lun=0`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		linstorCfg := linstorcontrol.Linstor{
			Loglevel:     log.GetLevel().String(),
			ControllerIP: controller,
		}
		targetCfg := iscsi.TargetConfig{
			IQN:  iqn,
			LUNs: []*iscsi.LUN{&iscsi.LUN{ID: uint8(lun)}},
		}
		target := iscsi.NewTargetMust(targetCfg)
		iscsiCfg := &iscsi.ISCSI{Linstor: linstorCfg, Target: target}
		rscStateMap, err := iscsiCfg.ProbeResource()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Current state of CRM resources\niSCSI resource %s, logical unit #%d:\n", iqn, uint8(lun))
		for rscName, runState := range rscStateMap {
			var label aurora.Value
			switch runState {
			case crmcontrol.Running:
				label = aurora.Green("Running")
			case crmcontrol.Stopped:
				label = aurora.Red("Stopped")
			default:
				label = aurora.Yellow("Unknown")
			}
			fmt.Printf("    %-40s %s\n", rscName, label)
		}
	},
}

func init() {
	rootCmd.AddCommand(probeCmd)

	probeCmd.MarkPersistentFlagRequired("iqn")
	probeCmd.MarkPersistentFlagRequired("lun")
}
