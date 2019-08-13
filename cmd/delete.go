package cmd

import (
	"net"

	"github.com/LINBIT/linstor-remote-storage/iscsi"
	"github.com/LINBIT/linstor-remote-storage/linstorcontrol"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes an iSCSI target",
	Long: `Deletes an iSCSI target by stopping and deliting the pacemaker resource
primitives and removing the linstor resources.

For example:
linstor-iscsi delete --iqn=iqn.2019-08.com.libit:example --lun=0`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		linstorCfg := linstorcontrol.Linstor{
			Loglevel:     log.GetLevel().String(),
			ControllerIP: controller,
		}
		targetCfg := iscsi.Target{
			IQN: iqn,
			LUN: uint8(lun),
		}
		iscsiCfg := &iscsi.ISCSI{Linstor: linstorCfg, Target: targetCfg}
		if err := iscsiCfg.DeleteResource(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller node")
}
