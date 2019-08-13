package cmd

import (
	"fmt"
	"math"

	"github.com/LINBIT/linstor-remote-storage/iscsi"
	"github.com/LINBIT/linstor-remote-storage/linstorcontrol"
	term "github.com/LINBIT/linstor-remote-storage/termcontrol"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists iSCSI targets",
	Long: `Lists the iSCSI targets created with this tool and provides an overview
about the existing Pacemaker and linstor parts

For example:
linstor-iscsi list`,
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
		_, config, err := iscsiCfg.ListResources()
		if err != nil {
			log.Fatal(err)
		}

		term.Color(term.COLOR_YELLOW)
		fmt.Print("Cluster resources:")

		indent := 1
		term.Color(term.COLOR_GREEN)
		iscsi.IndentPrint(indent, "\x1b[1;32miSCSI resources:\x1b[0m\n")
		indent++
		iscsi.IndentPrint(indent, "\x1b[1;32miSCSI targets:\x1b[0m\n")
		term.DefaultColor()

		indent++
		if len(config.TargetList) > 0 {
			for _, rscName := range config.TargetList {
				iscsi.IndentPrintf(indent, "%s\n", rscName)
			}
		} else {
			iscsi.IndentPrint(indent, "No resources\n")
		}
		indent--

		term.Color(term.COLOR_GREEN)
		iscsi.IndentPrint(indent, "\x1b[1;32miSCSI logical units:\x1b[0m\n")
		term.DefaultColor()

		indent++
		if len(config.LuList) > 0 {
			for _, rscName := range config.LuList {
				iscsi.IndentPrintf(indent, "%s\n", rscName)
			}
		} else {
			iscsi.IndentPrint(indent, "No resources\n")
		}
		indent -= 2

		term.Color(term.COLOR_TEAL)
		iscsi.IndentPrint(indent, "\x1b[1;32mOther cluster resources:\x1b[0m\n")
		term.DefaultColor()

		indent++
		if len(config.OtherRscList) > 0 {
			for _, rscName := range config.OtherRscList {
				iscsi.IndentPrintf(indent, "%s\n", rscName)
			}
		} else {
			iscsi.IndentPrint(indent, "No resources\n")
		}
		indent = 0

		fmt.Print("\n")

		if config.TidSet.Len() > 0 {
			term.Color(term.COLOR_GREEN)
			iscsi.IndentPrint(indent, "\x1b[1;32mAllocated TIDs:\x1b[0m\n")
			term.DefaultColor()

			indent++

			for _, tid := range config.TidSet.SortedKeys() {
				iscsi.IndentPrintf(indent, "%d\n", tid)
			}

			indent--
		} else {
			term.Color(term.COLOR_DARK_GREEN)
			iscsi.IndentPrint(indent, "\x1b[1;32mNo TIDs allocated\x1b[0m\n")
			term.DefaultColor()
		}
		fmt.Print("\n")

		freeTid, ok := config.TidSet.GetFree(1, math.MaxInt16)
		if ok {
			term.Color(term.COLOR_GREEN)
			iscsi.IndentPrintf(indent, "\x1b[1;32mNext free TID:\x1b[0m\n    %d\n", int(freeTid))
		} else {
			term.Color(term.COLOR_RED)
			iscsi.IndentPrint(indent, "\x1b[1;31mNo free TIDs\x1b[0m\n")
		}
		term.DefaultColor()
		fmt.Print("\n")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
