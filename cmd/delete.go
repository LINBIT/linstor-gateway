package cmd

import (
	"log"

	"github.com/LINBIT/linstor-remote-storage/application"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes an iSCSI target",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := application.DeleteResource(iqn, uint8(lun)); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
