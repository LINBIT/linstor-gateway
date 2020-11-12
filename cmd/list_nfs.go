package cmd

import (
	"net"
	"os"

	"github.com/LINBIT/linstor-gateway/pkg/crmcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/nfs"
	"github.com/LINBIT/linstor-gateway/pkg/nfsbase"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// listCommand represents the list command
func listNFSCommand() *cobra.Command {
	var controller net.IP
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists NFS resources",
		Long: `Lists the NFS resources created with this tool and provides an overview
about the existing Pacemaker and linstor parts

For example:
linstor-nfs list`,
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

			nfsList, err := nfs.ListResources()
			if err != nil {
				log.Fatal(err)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Resource name", "LINSTOR resource", "Filesystem mountpoint", "NFS export", "Service IP"})
			whiteBold := tablewriter.Colors{tablewriter.FgBlueColor, tablewriter.Bold}
			table.SetHeaderColor(whiteBold, whiteBold, whiteBold, whiteBold, whiteBold)

			for _, nfsItem := range nfsList {
				nfsCfg := nfsbase.NFSConfig{
					ResourceName: nfsItem.ResourceName,
				}
				linstorCfg := linstorcontrol.Linstor{
					ResourceName: nfsItem.ResourceName,
					Loglevel:     log.GetLevel().String(),
					ControllerIP: controller,
				}
				nfsRsc := nfs.NFSResource{
					NFS:     nfsCfg,
					Linstor: linstorCfg,
				}
				rscState, err := nfsRsc.ProbeResource()
				if err != nil {
					log.WithFields(log.Fields{
						"resource": nfsItem.ResourceName,
					}).Warning("Cannot probe NFS resource: ", err)
				}

				linstorState, err := linstorCfg.AggregateResourceState()
				if err != nil {
					log.Warning(err)
					linstorState = linstorcontrol.Unknown
				}

				row := []string{
					nfsItem.ResourceName,
					linstorStateToStatus(linstorState),
					stateToStatus(rscState.MountpointState),
					stateToStatus(rscState.ExportFSState),
					stateToStatus(rscState.ServiceIPState),
				}
				table.Append(row)
			}

			table.SetAutoFormatHeaders(false)
			table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})

			table.Render() // Send output
		},
	}

	listCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller node")
	return listCmd
}
