package cmd

import (
	"net"

	"github.com/LINBIT/linstor-iscsi/pkg/crmcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/iscsi"
	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/targetutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deleteCommand represents the delete command
func deleteCommand() *cobra.Command {
	var controller net.IP
	var iqn string
	var lun int

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Deletes an iSCSI target",
		Long: `Deletes an iSCSI target by stopping and deleting the pacemaker resource
primitives and removing the linstor resources.

For example:
linstor-iscsi delete --iqn=iqn.2019-08.com.linbit:example --lun=1`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if !cmd.Flags().Changed("controller") {
				foundIP, err := crmcontrol.FindLinstorController()
				if err == nil { // it might be ok to not find it...
					controller = foundIP
				}
			}
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
			if err := iscsiCfg.DeleteResource(); err != nil {
				log.Fatal(err)
			}
		},
	}

	deleteCmd.Flags().StringVarP(&iqn, "iqn", "i", "", "Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)")
	deleteCmd.Flags().IntVarP(&lun, "lun", "l", 1, "Set the LUN Number (required)")
	deleteCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller node")

	deleteCmd.MarkFlagRequired("iqn")
	deleteCmd.MarkFlagRequired("lun")

	return deleteCmd
}
