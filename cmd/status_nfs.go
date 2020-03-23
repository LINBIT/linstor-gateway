package cmd

import (
	"fmt"
	"net"

	"github.com/LINBIT/linstor-iscsi/pkg/crmcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/nfs"
	"github.com/LINBIT/linstor-iscsi/pkg/nfsbase"
	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// statusCommand represents the status command
func statusNFSCommand() *cobra.Command {
	var controller net.IP
	var resourceName string

	var statusCmd = &cobra.Command{
		Use:   "status-nfs",
		Short: "Displays the status of an NFS export",
		Long: `Triggers Pacemaker to probe the resoruce primitives of this NFS export.
That means Pacemaker will run the status operation on the nodes where the
resource can run.
This makes sure that Pacemakers view of the world is updated to the state
of the world.

For example:
linstor-iscsi status-nfs --resource=example`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if !cmd.Flags().Changed("controller") {
				foundIP, err := crmcontrol.FindLinstorController()
				if err == nil { // it might be ok to not find it...
					controller = foundIP
				} else {
					log.Debugf("Could not find LINSTOR controller in CIB")
				}
			}
			log.Debugf("Using LINSTOR controller at %s", controller)

			linstorCfg := linstorcontrol.Linstor{
				Loglevel:     log.GetLevel().String(),
				ControllerIP: controller,
				ResourceName: resourceName,
			}
			nfsCfg := nfsbase.NfsConfig{
				ResourceName: resourceName,
			}
			nfsRsc := nfs.NfsResource{
				Nfs:     nfsCfg,
				Linstor: linstorCfg,
			}

			state, err := nfsRsc.ProbeResource()
			if err != nil {
				log.Fatal(err)
			}

			linstorState, err := linstorCfg.AggregateResourceState()
			if err != nil {
				log.Warning(err)
				linstorState = linstorcontrol.Unknown
			}

			fmt.Printf("Status of NFS export %s:\n", aurora.Cyan(resourceName))
			fmt.Printf("- Mountpoint: %s\n", stateToLongStatus(state.MountpointState))
			fmt.Printf("- NFS export: %s\n", stateToLongStatus(state.ExportFSState))
			fmt.Printf("- Service IP: %s\n", stateToLongStatus(state.ServiceIPState))
			fmt.Printf("- LINSTOR resource: %s\n", linstorStateToLongStatus(linstorState))
		},
	}

	statusCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller")
	statusCmd.Flags().StringVarP(&resourceName, "resource", "r", "", "Set the resource name (required)")
	statusCmd.MarkFlagRequired("resource")

	return statusCmd
}
