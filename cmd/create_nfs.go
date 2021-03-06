package cmd

import (
	"net"

	"github.com/LINBIT/gopacemaker/cib"
	"github.com/LINBIT/linstor-gateway/pkg/crmcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/nfs"
	"github.com/LINBIT/linstor-gateway/pkg/nfsbase"
	"github.com/rck/unit"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func createNFSCommand() *cobra.Command {
	var controller net.IP
	var resourceGroupName string
	var resourceName string
	var serviceIPCIDR string
	var allowedIPsCIDR string

	var sz *unit.Value
	var sizeKiB uint64
	var serviceIP net.IP
	var serviceIPNet *net.IPNet
	var allowedIPs net.IP
	var allowedIPsNet *net.IPNet

	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Creates an NFS export",
		Long: `Creates a highly available NFS export based on LINSTOR and Pacemaker.
At first it creates a new resource within the linstor system under the
specified name and using the specified resource group.
After that it creates resource primitives in the Pacemaker cluster including
all necessary order and location constraints. The Pacemaker primites are
prefixed with p_, contain the resource name and a resource type postfix.

For example:
linstor-nfs create --resource=example --service-ip=192.168.211.122/32  \
 --allowed-ips=192.168.0.0/16 --resource-group=ssd_thin_2way --size=2G

Creates linstor resource example, volume 0 and
pacemaker primitives p_nfs_example_ip, p_nfs_example, p_nfs_example_export`,

		Args: cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			sizeKiB = uint64(sz.Value / unit.DefaultUnits["K"])

			var err error
			serviceIP, serviceIPNet, err = net.ParseCIDR(serviceIPCIDR)
			if err != nil {
				log.Fatalf("Invalid service IP address: %s", err.Error())
			}
			allowedIPs, allowedIPsNet, err = net.ParseCIDR(allowedIPsCIDR)
			if err != nil {
				log.Fatalf("Invalid allowed-ips parameter: %s", err.Error())
			}
		},

		Run: func(cmd *cobra.Command, args []string) {
			if !cmd.Flags().Changed("controller") {
				foundIP, err := crmcontrol.FindLinstorController()
				if err == nil { // it might be ok to not find it...
					controller = foundIP
				} else if err == cib.ErrCibFailed {
					log.Fatal(err)
				}
			}
			linstorCfg := linstorcontrol.Linstor{
				ControllerIP:      controller,
				ResourceGroupName: resourceGroupName,
			}

			serviceIPNetBits, _ := serviceIPNet.Mask.Size()
			allowedIPsNetBits, _ := allowedIPsNet.Mask.Size()

			nfsCfg := nfsbase.NFSConfig{
				ResourceName:      resourceName,
				ServiceIP:         serviceIP,
				ServiceIPNetBits:  serviceIPNetBits,
				AllowedIPs:        allowedIPs,
				AllowedIPsNetBits: allowedIPsNetBits,
				SizeKiB:           sizeKiB,
			}
			nfsRsc := nfs.NFSResource{
				NFS:     nfsCfg,
				Linstor: linstorCfg,
			}
			err := nfsRsc.CreateResource()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	createCmd.Flags().StringVarP(&resourceGroupName, "resource-group", "g", "default", "Set the LINSTOR resource group name")
	createCmd.Flags().StringVarP(&resourceName, "resource", "r", "", "Set the resource name (required)")
	createCmd.Flags().StringVar(&serviceIPCIDR, "service-ip", "127.0.0.1/8", "Set the service IP and netmask of the target (required)")
	createCmd.Flags().StringVar(&allowedIPsCIDR, "allowed-ips", "127.0.0.1/8", "Set the IP address mask of clients that are allowed access")
	createCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller node")

	units := unit.DefaultUnits
	units["KiB"] = units["K"]
	units["MiB"] = units["M"]
	units["GiB"] = units["G"]
	units["TiB"] = units["T"]
	units["PiB"] = units["P"]
	units["EiB"] = units["E"]
	u := unit.MustNewUnit(units)
	sz = u.MustNewValue(1*units["G"], unit.None)
	createCmd.Flags().Var(sz, "size", "Set a size (e.g, 1TiB)")

	createCmd.MarkFlagRequired("resource")
	createCmd.MarkFlagRequired("service-ip")
	createCmd.MarkFlagRequired("allowed-ips")

	return createCmd
}
