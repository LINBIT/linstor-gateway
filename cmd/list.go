package cmd

import (
	"fmt"
	"log"

	"github.com/LINBIT/linstor-remote-storage/application"
	"github.com/LINBIT/linstor-remote-storage/crmcontrol"
	term "github.com/LINBIT/linstor-remote-storage/termcontrol"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists iSCSI targets",
	Long:
`Lists the iSCSI targets created with this tool and provides an overview
about the existing Pacemaker and linstor parts

For example:
linstor-iscsi list`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		_, config, _, err := application.ListResources()
		if err != nil {
			log.Fatal(err)
		}

		term.Color(term.COLOR_YELLOW)
		fmt.Print("Cluster resources:")

		indent := 1
		term.Color(term.COLOR_GREEN)
		application.IndentPrint(indent, "\x1b[1;32miSCSI resources:\x1b[0m\n")
		indent++
		application.IndentPrint(indent, "\x1b[1;32miSCSI targets:\x1b[0m\n")
		term.DefaultColor()

		indent++
		if len(config.TargetList) > 0 {
			for _, rscName := range config.TargetList {
				application.IndentPrintf(indent, "%s\n", rscName)
			}
		} else {
			application.IndentPrint(indent, "No resources\n")
		}
		indent--

		term.Color(term.COLOR_GREEN)
		application.IndentPrint(indent, "\x1b[1;32miSCSI logical units:\x1b[0m\n")
		term.DefaultColor()

		indent++
		if len(config.LuList) > 0 {
			for _, rscName := range config.LuList {
				application.IndentPrintf(indent, "%s\n", rscName)
			}
		} else {
			application.IndentPrint(indent, "No resources\n")
		}
		indent -= 2

		term.Color(term.COLOR_TEAL)
		application.IndentPrint(indent, "\x1b[1;32mOther cluster resources:\x1b[0m\n")
		term.DefaultColor()

		indent++
		if len(config.OtherRscList) > 0 {
			for _, rscName := range config.OtherRscList {
				application.IndentPrintf(indent, "%s\n", rscName)
			}
		} else {
			application.IndentPrint(indent, "No resources\n")
		}
		indent = 0

		fmt.Print("\n")

		if config.TidSet.GetSize() > 0 {
			term.Color(term.COLOR_GREEN)
			application.IndentPrint(indent, "\x1b[1;32mAllocated TIDs:\x1b[0m\n")
			term.DefaultColor()

			indent++
			tidIter := config.TidSet.Iterator()
			for tid, isValid := tidIter.Next(); isValid; tid, isValid = tidIter.Next() {
				application.IndentPrintf(indent, "%d\n", tid)
			}
			indent--
		} else {
			term.Color(term.COLOR_DARK_GREEN)
			application.IndentPrint(indent, "\x1b[1;32mNo TIDs allocated\x1b[0m\n")
			term.DefaultColor()
		}
		fmt.Print("\n")

		freeTid, haveFreeTid := crmcontrol.GetFreeTargetId(config.TidSet.ToSortedArray())
		if haveFreeTid {
			term.Color(term.COLOR_GREEN)
			application.IndentPrintf(indent, "\x1b[1;32mNext free TID:\x1b[0m\n    %d\n", int(freeTid))
		} else {
			term.Color(term.COLOR_RED)
			application.IndentPrint(indent, "\x1b[1;31mNo free TIDs\x1b[0m\n")
		}
		term.DefaultColor()
		fmt.Print("\n")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
