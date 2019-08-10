package cmd

import (
	"net"

	"github.com/LINBIT/linstor-remote-storage/iscsi"
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
		if _, err := iscsi.DeleteResource(iqn, uint8(lun), log.GetLevel().String(), controller); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller node")
}
