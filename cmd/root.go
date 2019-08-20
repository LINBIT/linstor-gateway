package cmd

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var iqn string
var lun int

var loglevel string

var controller net.IP // create and delete

const (
	// This format is currently dictated by the iSCSI target backend,
	// specifically the rtslib-fb library.
	// A notable difference in this implementation (which also differs from
	// RFC3720, where the IQN format is defined) is that we require the
	// "unique" part after the colon to be present.
	//
	// See also the source code of rtslib-fb for the original regex:
	// https://github.com/open-iscsi/rtslib-fb/blob/b5be390be961/rtslib/utils.py#L384
	regexIQN = `iqn\.\d{4}-[0-1][0-9]\..*\..*`

	// This format is mandated by LINSTOR. Since we use the unique part
	// directly for LINSTOR resource names, it needs to be compliant.
	regexResourceName = `[[:alpha:]][[:alnum:]]+`

	regexWWN = `^` + regexIQN + `:` + regexResourceName + `$`
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "linstor-remote-storage",
	Short: "Manages Highly-Available iSCSI targets",
	Long: `linstor-iscsi manages higly available iSCSI targets by leveraging on linstor
and Pacemaker. Setting linstor including storage pools and resource groups
as well as Corosync and Pacemaker's properties a prerequisite to use this tool.`,
	Args: cobra.NoArgs,
	// We could have our custom flag types, but this is really simple enough...
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// TODO: regex most likely needs review
		level, err := log.ParseLevel(loglevel)
		if err != nil {
			log.Fatal(err)
		}
		log.SetLevel(level)

		if strings.ContainsAny(iqn, "_ ") {
			log.Fatal("IQN cannot contain the characters '_' (underscore) or ' ' (space)")
		}

		matched, err := regexp.MatchString(regexWWN, iqn)
		if err != nil {
			log.Fatal(err)
		} else if !matched {
			log.Fatal("Given IQN does not match the regular expression: " + regexWWN)
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
	rootCmd.PersistentFlags().StringVarP(&iqn, "iqn", "i", "", "Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)")
	rootCmd.PersistentFlags().IntVarP(&lun, "lun", "l", 0, "Set the LUN Number (required)")
	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", log.InfoLevel.String(), "Set the log level (as defined by logrus)")
}
