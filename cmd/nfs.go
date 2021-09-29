package cmd

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/nfs"
	"github.com/olekukonko/tablewriter"
	"github.com/rck/unit"
	"github.com/spf13/cobra"
)

func nfsCommands() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:     "nfs",
		Version: version,
		Short:   "Manages Highly-Available NFS exports",
		Long: `linstor-gateway nfs manages highly available NFS exports by leveraging LINSTOR
and drbd-reactor. Setting linstor including storage pools and resource groups
as well as Corosync and Pacemaker's properties a prerequisite to use this tool.`,
		Args: cobra.NoArgs,
	}

	rootCmd.DisableAutoGenTag = true

	rootCmd.AddCommand(createNFSCommand())
	rootCmd.AddCommand(deleteNFSCommand())
	rootCmd.AddCommand(listNFSCommand())
	rootCmd.AddCommand(serverCommand())

	return rootCmd

}

func createNFSCommand() *cobra.Command {
	var resourceGroupName string
	var resourceName string
	serviceIPCIDR := common.ServiceIPFromParts(net.IPv6loopback, 64)
	allowedIPsCIDR := common.ServiceIPFromParts(net.IPv6loopback, 64)
	exportPath := "/"

	var sz *unit.Value

	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Creates an NFS export",
		Long: `Creates a highly available NFS export based on LINSTOR and drbd-reactor.
At first it creates a new resource within the LINSTOR system under the
specified name and using the specified resource group.
After that it creates a drbd-reactor configuration to bring up a highly available NFS 
export.`,
		Example: "linstor-gateway nfs create --resource=example --service-ip=192.168.211.122/24 --allowed-ips=192.168.0.0/16 --resource-group=ssd_thin_2way --size=2G",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rsc := &nfs.ResourceConfig{
				Name:          resourceName,
				ResourceGroup: resourceGroupName,
				ServiceIP:     serviceIPCIDR,
				AllowedIPs:    []common.IpCidr{allowedIPsCIDR},
				Volumes: []nfs.VolumeConfig{
					{
						ExportPath: exportPath,
						VolumeConfig: common.VolumeConfig{
							SizeKiB: uint64(sz.Value / unit.K),
						},
					},
				},
			}
			n, err := nfs.New(controllers)
			if err != nil {
				return fmt.Errorf("failed to initialize nfs: %w", err)
			}

			_, err = n.Create(context.Background(), rsc)
			if err != nil {
				return err
			}

			fmt.Printf("Created export '%s' at %s:%s\n", resourceName, serviceIPCIDR.IP().String(), nfs.ExportPath(rsc, &rsc.Volumes[0]))
			return nil
		},
	}

	createCmd.Flags().StringVarP(&resourceGroupName, "resource-group", "g", "", "Set the LINSTOR resource group name")
	createCmd.Flags().StringVarP(&resourceName, "resource", "r", "", "Set the resource name (required)")
	createCmd.Flags().StringVarP(&exportPath, "export-path", "p", exportPath, "Set the export path")
	createCmd.Flags().VarP(&serviceIPCIDR, "service-ip", "", "Set the service IP and netmask of the target (required)")
	createCmd.Flags().VarP(&allowedIPsCIDR, "allowed-ips", "", "Set the IP address mask of clients that are allowed access")

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

func deleteNFSCommand() *cobra.Command {
	var resourceName string

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Deletes an NFS export",
		Long: `Deletes an NFS export by stopping and deleting the drbd-reactor config
and removing the LINSTOR resources.`,
		Example: "linstor-gateway nfs delete --resource=example",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := nfs.New(controllers)
			if err != nil {
				return fmt.Errorf("failed to initialize nfs: %w", err)
			}
			err = n.Delete(context.Background(), resourceName)
			if err != nil {
				return err
			}

			fmt.Printf("Deleted export %s\n", resourceName)
			return nil
		},
	}

	deleteCmd.Flags().StringVarP(&resourceName, "resource", "r", "", "Set the resource name (required)")

	deleteCmd.MarkFlagRequired("resource")

	return deleteCmd
}

func listNFSCommand() *cobra.Command {
	var controller net.IP
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists NFS resources",
		Long: `Lists the NFS resources created with this tool and provides an overview
about the existing LINSTOR resources and service status.`,
		Example: "linstor-gateway nfs list",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := nfs.New(controllers)
			if err != nil {
				return fmt.Errorf("failed to initialize nfs: %w", err)
			}
			list, err := n.List(context.Background())
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Resource", "Service IP", "Service state", "NFS export", "LINSTOR state"})
			table.SetHeaderColor(tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader)

			for _, resource := range list {
				for _, vol := range resource.Volumes {
					withStatus := resource.VolumeConfig(vol.Number)
					if withStatus == nil {
						withStatus = &common.Volume{Status: common.VolumeState{State: common.Unknown}}
					}

					table.Rich([]string{
						resource.Name,
						resource.ServiceIP.String(),
						resource.Status.Service.String(),
						nfs.ExportPath(resource, &vol),
						withStatus.Status.State.String(),
					}, []tablewriter.Colors{
						{},
						{},
						ServiceStateColor(resource.Status.Service),
						{},
						ResourceStateColor(withStatus.Status.State),
					})
				}
			}

			table.SetAutoMergeCellsByColumnIndex([]int{0, 1})
			table.SetAutoFormatHeaders(false)

			table.Render() // Send output

			return nil
		},
	}

	listCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller node")
	return listCmd
}
