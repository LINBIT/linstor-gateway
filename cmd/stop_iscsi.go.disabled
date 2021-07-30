package cmd

import (
	"net"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/targetutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func stopISCSICommand() *cobra.Command {
	var controller net.IP
	var iqn string
	var lun int

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stops an iSCSI target",
		Long: `Sets the target role attribute of a Pacemaker primitive to stopped.
This causes pacemaker to stop the components of an iSCSI target.

For example:
linstor-iscsi start --iqn=iqn.2019-08.com.linbit:example --lun=1`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			linstorCfg := linstorcontrol.Linstor{
				ControllerIP: controller,
			}
			targetCfg := targetutil.TargetConfig{
				IQN:  iqn,
				LUNs: []*targetutil.LUN{&targetutil.LUN{ID: uint8(lun)}},
			}
			target := cliNewTargetMust(cmd, targetCfg)
			iscsiCfg := &iscsi.ISCSI{Linstor: linstorCfg, Target: target}
			if err := iscsiCfg.StopResource(); err != nil {
				log.Fatal(err)
			}
		},
	}

	stopCmd.Flags().StringVarP(&iqn, "iqn", "i", "", "Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)")
	stopCmd.Flags().IntVarP(&lun, "lun", "l", 1, "Set the LUN Number (required)")

	stopCmd.MarkFlagRequired("iqn")
	stopCmd.MarkFlagRequired("lun")

	return stopCmd
}
