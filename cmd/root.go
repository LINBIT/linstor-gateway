package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

var iqn string
var lun int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "linstor-remote-storage",
	Short: "Manages Highly-Available iSCSI targets",
	Long:
`linstor-iscsi manages higly available iSCSI targets by leveraging on linstor
and Pacemaker. Setting linstor including storage pools and resource groups
as well as Corosync and Pacemaker's properties a prerequisite to use this tool.`,
	Args: cobra.NoArgs,
	// We could have our custom flag types, but this is really simple enough...
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// TODO: regex most likely needs review
		matched, err := regexp.MatchString(`^iqn\.\d{4}-\d{2}\..*:.*`, iqn)
		if err != nil {
			log.Fatal(err)
		} else if !matched {
			log.Fatal("Given IQN does not match specification")
		}

		if lun < 0 || lun > 255 {
			log.Fatal("LUN out of range [0-255]")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&iqn, "iqn", "i", "", "Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.libit:unique) (required)")
	rootCmd.PersistentFlags().IntVarP(&lun, "lun", "l", 0, "Set the LUN Number (required)")

	rootCmd.MarkPersistentFlagRequired("iqn")
	rootCmd.MarkPersistentFlagRequired("lun")
}
