package cmd

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
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

	return rootCmd

}

func createNFSCommand() *cobra.Command {
	resourceGroup := "DfltRscGrp"
	allowedIPsCIDR := common.ServiceIPFromParts(net.IPv4zero, 0)
	exportPath := "/"

	cmd := &cobra.Command{
		Use:   "create NAME SERVICE_IP SIZE",
		Short: "Creates an NFS export",
		Long: `Creates a highly available NFS export based on LINSTOR and drbd-reactor.
At first it creates a new resource within the LINSTOR system under the
specified name and using the specified resource group.
After that it creates a drbd-reactor configuration to bring up a highly available NFS 
export.`,
		Example: `linstor-gateway nfs create example 192.168.211.122/24 2G
linstor-gateway nfs create restricted 10.10.22.44/16 2G --allowed-ips 10.10.0.0/16
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			resource := args[0]
			serviceIP, err := common.ServiceIPFromString(args[1])
			if err != nil {
				return err
			}

			size, err := unit.MustNewUnit(unit.DefaultUnits).ValueFromString(args[2])
			if err != nil {
				return err
			}

			rsc := &nfs.ResourceConfig{
				Name:          resource,
				ResourceGroup: resourceGroup,
				ServiceIP:     serviceIP,
				AllowedIPs:    []common.IpCidr{allowedIPsCIDR},
				Volumes: []nfs.VolumeConfig{{
					ExportPath: exportPath,
					VolumeConfig: common.VolumeConfig{
						Number:              1,
						SizeKiB:             uint64(size.Value / unit.K),
						FileSystem:          "ext4",
						FileSystemRootOwner: common.UidGid{Uid: 65534, Gid: 65534}, // corresponds to "nobody:nobody"
					},
				}},
			}
			_, err = cli.Nfs.Create(ctx, rsc)
			if err != nil {
				return err
			}

			fmt.Printf("Created export '%s' at %s:%s\n", resource, serviceIP.IP().String(), nfs.ExportPath(rsc, &rsc.Volumes[0]))
			return nil
		},
	}

	cmd.Flags().StringVarP(&resourceGroup, "resource-group", "r", resourceGroup, "LINSTOR resource group to use")
	cmd.Flags().StringVarP(&exportPath, "export-path", "p", exportPath, fmt.Sprintf("Set the export path, relative to %s", nfs.ExportBasePath))
	cmd.Flags().VarP(&allowedIPsCIDR, "allowed-ips", "", "Set the IP address mask of clients that are allowed access")

	return cmd
}

func deleteNFSCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete NAME",
		Short: "Deletes an NFS export",
		Long: `Deletes an NFS export by stopping and deleting the drbd-reactor config
and removing the LINSTOR resources.`,
		Example: "linstor-gateway nfs delete example",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			resourceName := args[0]
			err := cli.Nfs.Delete(ctx, resourceName)
			if err != nil {
				return err
			}

			fmt.Printf("Deleted export %s\n", resourceName)
			return nil
		},
	}
}

func listNFSCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists NFS resources",
		Long: `Lists the NFS resources created with this tool and provides an
overview about the existing LINSTOR resources and service status.`,
		Example: "linstor-gateway nfs list",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			list, err := cli.Nfs.GetAll(ctx)
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Resource", "Service IP", "Service state", "NFS export", "LINSTOR state"})
			table.SetHeaderColor(tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader)

			degradedResources := 0
			for _, resource := range list {
				for i, vol := range resource.Volumes {
					withStatus := resource.VolumeConfig(vol.Number)
					if withStatus == nil {
						withStatus = &common.Volume{Status: common.VolumeState{State: common.Unknown}}
					}

					if i == 0 {
						log.Debugf("not displaying cluster private volume: %+v", vol)
						continue
					}

					log.Debugf("listing volume: %+v", vol)

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
					if withStatus.Status.State != common.ResourceStateOK {
						degradedResources++
					}
				}
			}

			table.SetAutoMergeCellsByColumnIndex([]int{0, 1})
			table.SetAutoFormatHeaders(false)

			table.Render() // Send output

			if degradedResources > 0 {
				log.Warnf("Some resources are degraded. Run %s for possible solutions.", bold("linstor advise resource"))
			}

			return nil
		},
	}
}
