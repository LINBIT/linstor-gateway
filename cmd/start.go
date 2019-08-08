package cmd

import (
	"log"

	"github.com/LINBIT/linstor-remote-storage/application"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts an iSCSI target",
	Long:
`Sets the target role attribute of a Pacemaker primitive to started.
In case it does not start use your favourite pacemaker tools to analyze
the root cause.

For example:
linstor-iscsi start --iqn=iqn.2019-08.com.libit:example --lun=0`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := application.StartResource(iqn, uint8(lun)); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
