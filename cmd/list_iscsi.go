package cmd

import (
	"net"
	"os"
	"strconv"

	"github.com/LINBIT/linstor-gateway/pkg/crmcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/targetutil"
	"github.com/mattn/go-isatty"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

func listISCSICommand() *cobra.Command {
	var controller net.IP
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
					log.Debugf("Found LINSTOR controller at %s", foundIP)
				} else {
					log.Debugf("Could not find LINSTOR controller in CIB")
				}
			}
			log.Debugf("Using LINSTOR controller at %s", controller)
			targets, err := iscsi.ListResources()
			if err != nil {
				log.Fatal(err)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"IQN", "LUN", "Pacemaker LUN", "Pacemaker", "Pacemaker IP", "LINSTOR"})

			if isatty.IsTerminal(os.Stdout.Fd()) {
				whiteBold := tablewriter.Colors{tablewriter.FgBlueColor, tablewriter.Bold}
				table.SetHeaderColor(whiteBold, whiteBold, whiteBold, whiteBold, whiteBold, whiteBold)
			}

			for _, target := range targets {
				linstorCfg := linstorcontrol.Linstor{
					Loglevel:     log.GetLevel().String(),
					ControllerIP: controller,
				}
				targetCfg := targetutil.TargetConfig{
					IQN:  target.IQN,
					LUNs: target.LUNs,
				}
				tgt := targetutil.NewTargetMust(targetCfg)
				iscsiCfg := &iscsi.ISCSI{Linstor: linstorCfg, Target: tgt}

				resourceState, err := iscsiCfg.ProbeResource()
				if err != nil {
					log.WithFields(log.Fields{
						"iqn": target.IQN,
					}).Warning("Cannot probe target: ", err)
				}

				for _, lu := range target.LUNs {
					targetName, err := targetutil.ExtractTargetName(target.IQN)
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

	listCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller node")
	return listCmd
}
