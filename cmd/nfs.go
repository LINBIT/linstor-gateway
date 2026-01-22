package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/nfs"
	"github.com/LINBIT/linstor-gateway/pkg/prompt"
	"github.com/LINBIT/linstor-gateway/pkg/upgrade"
	"github.com/LINBIT/linstor-gateway/pkg/version"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/rck/unit"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func nfsCommands() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:     "nfs",
		Version: version.Version,
		Short:   "Manages Highly-Available NFS exports",
		Long: `linstor-gateway nfs manages highly available NFS exports by leveraging LINSTOR
and drbd-reactor. A running LINSTOR cluster including storage pools and resource groups
is a prerequisite to use this tool.

NOTE that only one NFS resource can exist in a cluster.
See "help nfs create" for more information`,
		Args: cobra.NoArgs,
	}

	rootCmd.DisableAutoGenTag = true

	rootCmd.AddCommand(createNFSCommand())
	rootCmd.AddCommand(deleteNFSCommand())
	rootCmd.AddCommand(listNFSCommand())
	rootCmd.AddCommand(upgradeNFSCommand())

	return rootCmd

}

func autoGenerateExportPaths(exportPath string, numVolumes int) []string {
	exportPaths := make([]string, numVolumes)
	for i := 0; i < numVolumes; i++ {
		exportPaths[i] = filepath.Join(exportPath, fmt.Sprintf("vol%d", i+1))
	}
	return exportPaths
}

func createNFSCommand() *cobra.Command {
	resourceGroup := "DfltRscGrp"
	allowedIPsCIDR := common.ServiceIPFromParts(net.IPv4zero, 0)
	exportPaths := []string{"/"}
	grossSize := false
	filesystem := "ext4"
	var resourceTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "create NAME SERVICE_IP [VOLUME_SIZE]...",
		Short: "Creates an NFS export",
		Long: `Creates a highly available NFS export based on LINSTOR and drbd-reactor.
At first it creates a new resource within the LINSTOR system under the
specified name and using the specified resource group.
After that it creates a drbd-reactor configuration to bring up a highly available NFS
export.

!!! NOTE that only one NFS resource can exist in a cluster.
To create multiple exports, create a single resource with multiple volumes.`,
		Example: `linstor-gateway nfs create example 192.168.211.122/24 2G
linstor-gateway nfs create restricted 10.10.22.44/16 2G --allowed-ips 10.10.0.0/16
linstor-gateway nfs create multi 172.16.16.55/24 1G 2G --export-path /music --export-path /movies
`,
		Args: cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			resource := args[0]
			serviceIP, err := common.ServiceIPFromString(args[1])
			if err != nil {
				return err
			}

			rawSizes := args[2:]
			if len(rawSizes) != len(exportPaths) {
				if len(exportPaths) == 1 {
					// special case: when we have multiple volumes, but only one export path, try to auto-generate the export paths
					exportPaths = autoGenerateExportPaths(exportPaths[0], len(rawSizes))
				} else {
					return fmt.Errorf("specified %d volumes but %d export paths. Need exactly one export path per volume", len(rawSizes), len(exportPaths))
				}
			}

			var volumes []nfs.VolumeConfig
			for i, rawValue := range rawSizes {
				val, err := unit.MustNewUnit(unit.DefaultUnits).ValueFromString(rawValue)
				if err != nil {
					return err
				}

				volumes = append(volumes, nfs.VolumeConfig{
					ExportPath: exportPaths[i],
					VolumeConfig: common.VolumeConfig{
						Number:              i + 1,
						SizeKiB:             uint64(val.Value / unit.K),
						FileSystem:          filesystem,
						FileSystemRootOwner: common.UserGroup{User: "nobody", Group: "nobody"},
					},
				})
			}

			rsc := &nfs.ResourceConfig{
				Name:            resource,
				ResourceGroup:   resourceGroup,
				ServiceIP:       serviceIP,
				AllowedIPs:      []common.IpCidr{allowedIPsCIDR},
				Volumes:         volumes,
				GrossSize:       grossSize,
				ResourceTimeout: resourceTimeout,
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
	cmd.Flags().StringSliceVarP(&exportPaths, "export-path", "p", exportPaths, fmt.Sprintf("Set the export path, relative to %s. Can be specified multiple times when creating more than one volume", nfs.ExportBasePath))
	cmd.Flags().VarP(&allowedIPsCIDR, "allowed-ips", "", "Set the IP address mask of clients that are allowed access")
	cmd.Flags().BoolVar(&grossSize, "gross", false, "Make all size options specify gross size, i.e. the actual space used on disk")
	cmd.Flags().StringVarP(&filesystem, "filesystem", "f", filesystem, "File system type to use (ext4 or xfs)")
	cmd.Flags().DurationVar(&resourceTimeout, "resource-timeout", nfs.DefaultResourceTimeout, "Timeout for waiting for the resource to become available")

	return cmd
}

func deleteNFSCommand() *cobra.Command {
	var force bool
	var resourceTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Deletes an NFS export",
		Long: `Deletes an NFS export by stopping and deleting the drbd-reactor config
and removing the LINSTOR resources.`,
		Example: "linstor-gateway nfs delete example",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			resourceName := args[0]

			var yes bool
			if force {
				yes = true
			} else {
				fmt.Printf("%s: Deleting NFS export %q %s.\n",
					color.YellowString("WARNING"), resourceName,
					bold("and all data stored on it"))
				yes = prompt.Confirm("Continue?")
			}

			if yes {
				err := cli.Nfs.Delete(ctx, resourceName, resourceTimeout)
				if err != nil {
					return err
				}

				fmt.Printf("Deleted export %q\n", resourceName)
			} else {
				fmt.Println("Aborted")
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Delete without prompting for confirmation")
	cmd.Flags().DurationVar(&resourceTimeout, "resource-timeout", nfs.DefaultResourceTimeout, "Timeout for waiting for the resource to become unavailable")

	return cmd
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

			var degradedResources, badResources []string
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

					serviceStatus := resource.Status.Service.String()
					if resource.Status.Service == common.ServiceStateStarted && resource.Status.Primary != "" {
						serviceStatus += " (" + resource.Status.Primary + ")"
					}
					table.Rich([]string{
						resource.Name,
						resource.ServiceIP.String(),
						serviceStatus,
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
						if !contains(degradedResources, resource.Name) {
							degradedResources = append(degradedResources, resource.Name)
						}
					}
				}
				if len(resource.Volumes) == 0 {
					table.Rich(
						[]string{resource.Name, resource.ServiceIP.String(), resource.Status.Service.String(), "", common.ResourceStateBad.String()},
						[]tablewriter.Colors{{}, {}, ServiceStateColor(resource.Status.Service), {}, ResourceStateColor(common.ResourceStateBad)},
					)
					badResources = append(badResources, resource.Name)
				}
			}

			table.SetAutoMergeCellsByColumnIndex([]int{0, 1})
			table.SetAutoFormatHeaders(false)

			table.Render() // Send output

			if len(degradedResources) > 0 {
				log.Warnf("Some resources are degraded. Run %s for possible solutions.", bold("linstor advise resource"))
				for _, r := range degradedResources {
					log.Warnf("- %s", r)
				}
			}

			if len(badResources) > 0 {
				log.Warnf("Some resources are broken. Check %s and verify that these resources are intact:", bold("linstor volume list"))
				for _, r := range badResources {
					log.Warnf("- %s", r)
				}
			}

			return nil
		},
	}
}

func upgradeNFSCommand() *cobra.Command {
	var forceYes bool
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "upgrade NAME",
		Short: "Check existing resources and upgrade their configuration if necessary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			controllers := viper.GetStringSlice("linstor.controllers")
			cli, err := linstorcontrol.Default(controllers)
			if err != nil {
				return err
			}
			err = upgrade.Nfs(cmd.Context(), cli.Client, args[0], forceYes, dryRun)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringSlice("controllers", nil, "List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)")
	cmd.Flags().BoolVarP(&forceYes, "yes", "y", false, "Run non-interactively; answer all questions with yes")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Display potential updates without taking any actions")
	_ = viper.BindPFlag("linstor.controllers", cmd.Flags().Lookup("controllers"))

	return cmd
}
