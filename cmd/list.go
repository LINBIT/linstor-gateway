package cmd

import (
	"os"
	"strconv"

	"github.com/LINBIT/linstor-remote-storage/crmcontrol"
	"github.com/LINBIT/linstor-remote-storage/iscsi"
	"github.com/LINBIT/linstor-remote-storage/linstorcontrol"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	statusOk       = aurora.Green("✓").String()
	statusStarting = aurora.Yellow("⌛").String()
	statusBad      = aurora.Red("✗").String()
)

func stateToStatus(state crmcontrol.LrmRunState) string {
	if state.HaveState {
		if state.Running {
			return statusOk
		} else {
			return statusBad
		}
	}

	return statusStarting
}

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
		_, targets, err := iscsi.ListResources()
		if err != nil {
			log.Fatal(err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Target Name", "LUN", "Pacemaker LUN", "Pacemaker", "Pacemaker IP"})
		whiteBold := tablewriter.Colors{tablewriter.FgBlueColor, tablewriter.Bold}
		table.SetHeaderColor(whiteBold, whiteBold, whiteBold, whiteBold, whiteBold)

		for _, target := range targets {
			targetCfg := iscsi.Target{
				IQN:  target.IQN,
				LUNs: target.LUNs,
			}
			iscsiCfg := &iscsi.ISCSI{Linstor: linstorCfg, Target: targetCfg}

			rscStateMap, err := iscsiCfg.ProbeResource()
			if err != nil {
				log.WithFields(log.Fields{
					"iqn": target.IQN,
				}).Warning("Cannot probe target: ", err)
			}

			for _, lu := range target.LUNs {
				state := (*rscStateMap)[target.Name]
				// TODO stop using this hack and pass the actual
				// name through once all the data structures are fixed.
				lunState := (*rscStateMap)[target.Name+"_lu"+strconv.Itoa(int(lu.ID))]

				// TODO don't know how to deal with the IPs yet...
				row := []string{target.Name, strconv.Itoa(int(lu.ID)), stateToStatus(state), stateToStatus(lunState), statusOk}
				table.Append(row)
			}
		}

		// TODO this would look cool, but it would merge the ticks too...
		//table.SetAutoMergeCells(true)
		table.SetAutoFormatHeaders(false)
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})

		table.Render() // Send output
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
