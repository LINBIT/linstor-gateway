package cmd

import (
	"net"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/targetutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func startCommand() *cobra.Command {
	var controller net.IP
	var iqn string
	var lun int

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts an iSCSI target",
		Long: `Sets the target role attribute of a Pacemaker primitive to started.
In case it does not start use your favourite pacemaker tools to analyze
the root cause.

For example:
linstor-iscsi start --iqn=iqn.2019-08.com.linbit:example --lun=1`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			linstorCfg := linstorcontrol.Linstor{
				Loglevel:     log.GetLevel().String(),
				ControllerIP: controller,
			}
			targetCfg := targetutil.TargetConfig{
				IQN:  iqn,
				LUNs: []*targetutil.LUN{&targetutil.LUN{ID: uint8(lun)}},
			}
			target := cliNewTargetMust(cmd, targetCfg)
			iscsiCfg := &iscsi.ISCSI{Linstor: linstorCfg, Target: target}
			if err := iscsiCfg.StartResource(); err != nil {
				log.Fatal(err)
			}
		},
	}

	startCmd.Flags().StringVarP(&iqn, "iqn", "i", "", "Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)")
	startCmd.Flags().IntVarP(&lun, "lun", "l", 1, "Set the LUN Number (required)")
	startCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller node")

	startCmd.MarkFlagRequired("iqn")
	startCmd.MarkFlagRequired("lun")

	return startCmd
}
