package cmd

import (
	"github.com/LINBIT/linstor-remote-storage/iscsi"
	"github.com/LINBIT/linstor-remote-storage/linstorcontrol"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops an iSCSI target",
	Long: `Sets the target role attribute of a Pacemaker primitive to stopped.
This causes pacemaker to stop the components of an iSCSI target.

For example:
linstor-iscsi start --iqn=iqn.2019-08.com.libit:example --lun=0`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		linstorCfg := linstorcontrol.Linstor{
			Loglevel:     log.GetLevel().String(),
			ControllerIP: controller,
		}
		targetCfg := iscsi.Target{
			IQN:  iqn,
			LUNs: []*iscsi.LUN{&iscsi.LUN{ID: uint8(lun)}},
		}
		iscsiCfg := &iscsi.ISCSI{Linstor: linstorCfg, Target: targetCfg}
		if err := iscsiCfg.StopResource(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	stopCmd.MarkPersistentFlagRequired("iqn")
	stopCmd.MarkPersistentFlagRequired("lun")
}
