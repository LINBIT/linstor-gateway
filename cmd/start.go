package cmd

import (
	"github.com/LINBIT/linstor-remote-storage/iscsi"
	"github.com/LINBIT/linstor-remote-storage/linstorcontrol"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts an iSCSI target",
	Long: `Sets the target role attribute of a Pacemaker primitive to started.
In case it does not start use your favourite pacemaker tools to analyze
the root cause.

For example:
linstor-iscsi start --iqn=iqn.2019-08.com.linbit:example --lun=0`,
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
		if err := iscsiCfg.StartResource(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.MarkPersistentFlagRequired("iqn")
	startCmd.MarkPersistentFlagRequired("lun")
}
