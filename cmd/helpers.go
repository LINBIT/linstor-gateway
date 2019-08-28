package cmd

import (
	"os"

	"github.com/LINBIT/linstor-iscsi/pkg/targetutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// cliNewTargetMust is a simple, but helpful helper around the required iqn/lun parameters
func cliNewTargetMust(cmd *cobra.Command, targetCfg targetutil.TargetConfig) targetutil.Target {
	var errs []string
	if !cmd.Flags().Changed("iqn") {
		errs = append(errs, "Parameter '--iqn' is mandatory")
	}
	if !cmd.Flags().Changed("lun") {
		errs = append(errs, "Parameter '--lun' is mandatory")
	}

	if len(errs) > 0 {
		cmd.Help()
		for _, e := range errs {
			log.Errorln(e)
		}
		os.Exit(1)
	}

	return targetutil.NewTargetMust(targetCfg)
}
