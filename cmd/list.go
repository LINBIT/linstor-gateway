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
	statusBad      = aurora.Red("✗").String()
	statusUnknown  = aurora.Yellow("?").String()
	statusDegraded = aurora.Yellow("!").String()
)

func stateToStatus(state crmcontrol.LrmRunState) string {
	switch state {
	case crmcontrol.Running:
		return statusOk
	case crmcontrol.Stopped:
		return statusBad
	default:
		return statusUnknown
	}
}

func linstorStateToStatus(state linstorcontrol.ResourceState) string {
	switch state {
	case linstorcontrol.Ok:
		return statusOk
	case linstorcontrol.Degraded:
		return statusDegraded
	case linstorcontrol.Bad:
		return statusBad
	default:
		return statusUnknown
	}
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
		if !cmd.Flags().Changed("controller") {
			foundIP, err := crmcontrol.FindLinstorController()
			if err == nil { // it might be ok to not find it...
				controller = foundIP
			}
		}
		_, targets, err := iscsi.ListResources()
		if err != nil {
			log.Fatal(err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"IQN", "LUN", "Pacemaker LUN", "Pacemaker", "Pacemaker IP", "LINSTOR"})
		whiteBold := tablewriter.Colors{tablewriter.FgBlueColor, tablewriter.Bold}
		table.SetHeaderColor(whiteBold, whiteBold, whiteBold, whiteBold, whiteBold, whiteBold)

		for _, target := range targets {
			linstorCfg := linstorcontrol.Linstor{
				Loglevel:     log.GetLevel().String(),
				ControllerIP: controller,
			}
			targetCfg := iscsi.TargetConfig{
				IQN:  target.IQN,
				LUNs: target.LUNs,
			}
			tgt := iscsi.NewTargetMust(targetCfg)
			iscsiCfg := &iscsi.ISCSI{Linstor: linstorCfg, Target: tgt}

			resourceState, err := iscsiCfg.ProbeResource()
			if err != nil {
				log.WithFields(log.Fields{
					"iqn": target.IQN,
				}).Warning("Cannot probe target: ", err)
			}

			for _, lu := range target.LUNs {
				targetName, err := iscsi.ExtractTargetName(target.IQN)
				if err != nil {
					log.Fatal(err)
				}

				linstorCfg.ResourceName = linstorcontrol.ResourceNameFromLUN(targetName, lu.ID)
				targetState := resourceState.TargetState
				lunState := resourceState.LUStates[lu.ID]
				ipState := resourceState.IPState
				linstorState, err := linstorCfg.AggregateResourceState()
				if err != nil {
					log.Warning(err)
					linstorState = linstorcontrol.Unknown
				}

				row := []string{
					target.IQN,
					strconv.Itoa(int(lu.ID)),
					stateToStatus(lunState),
					stateToStatus(targetState),
					stateToStatus(ipState),
					linstorStateToStatus(linstorState),
				}
				table.Append(row)
			}
		}

		// TODO this would look cool, but it would merge the ticks too...
		//table.SetAutoMergeCells(true)
		table.SetAutoFormatHeaders(false)
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})

		table.Render() // Send output
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
