package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/olekukonko/tablewriter"
	"github.com/rck/unit"
	"github.com/spf13/cobra"
)

func iscsiCommands() *cobra.Command {

	var rootCmd = &cobra.Command{
		Use:     "iscsi",
		Version: version,
		Short:   "Manages Highly-Available iSCSI targets",
		Long: `linstor-gateway iscsi manages highly available iSCSI targets by leveraging
LINSTOR and drbd-reacor. Setting up LINSTOR, including storage pools and resource groups,
as well as drbd-reactor is a prerequisite to use this tool.`,
		Args: cobra.NoArgs,
	}

	rootCmd.DisableAutoGenTag = true

	rootCmd.AddCommand(createISCSICommand())
	rootCmd.AddCommand(deleteISCSICommand())
	rootCmd.AddCommand(listISCSICommand())
	rootCmd.AddCommand(serverCommand())
	rootCmd.AddCommand(startISCSICommand())
	rootCmd.AddCommand(stopISCSICommand())

	return rootCmd
}

func createISCSICommand() *cobra.Command {
	var username, password, portals, group string
	var iqn iscsi.Iqn

	var lun int

	var sz *unit.Value
	var serviceIp common.IpCidr

	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Creates an iSCSI target",
		Long: `Creates a highly available iSCSI target based on LINSTOR and drbd-reactor.
At first it creates a new resource within the LINSTOR system, using the
specified resource group. The name of the linstor resources is derived
from the IQN's World Wide Name, which must be unique'.
After that it creates a configuration for drbd-reactor to manage the
high availabilitiy primitives.`,
		Example: "linstor-gateway iscsi create --iqn=iqn.2019-08.com.linbit:example --ip=192.168.122.181/24 --username=foo --lun=1 --password=bar --resource-group=ssd_thin_2way --size=2G",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			i, err := iscsi.New(controllers)
			if err != nil {
				return fmt.Errorf("failed to initialize iscsi: %w", err)
			}

			_, err = i.Create(ctx, &iscsi.ResourceConfig{
				IQN:       iqn,
				Username:  username,
				Password:  password,
				ServiceIP: serviceIp,
				Volumes: []common.VolumeConfig{
					{Number: lun, SizeKiB: uint64(sz.Value / unit.K)},
				},
			})
			if err != nil {
				return err
			}

			fmt.Printf("Created iSCSI target '%s'\n", iqn)

			return nil
		},
	}

	createCmd.Flags().Var(&serviceIp, "ip", "Set the service IP and netmask of the target")
	createCmd.Flags().VarP(&iqn, "iqn", "i", "Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique)")
	createCmd.Flags().IntVarP(&lun, "lun", "l", 1, "Set the LUN")
	createCmd.Flags().StringVar(&portals, "portals", "", "Set up portals, if unset, the service ip and default port")
	createCmd.Flags().StringVarP(&username, "username", "u", "", "Set the username (required)")
	createCmd.Flags().StringVarP(&password, "password", "p", "", "Set the password (required)")

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

	createCmd.Flags().StringVarP(&group, "resource-group", "g", "DfltRscGrp", "Set the LINSTOR resource-group")

	createCmd.MarkFlagRequired("ip")
	createCmd.MarkFlagRequired("iqn")
	createCmd.MarkFlagRequired("size")

	return createCmd
}

func listISCSICommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists iSCSI targets",
		Long: `Lists the iSCSI targets created with this tool and provides an overview
about the existing drbd-reactor and linstor parts.`,
		Example: "linstor-gateway iscsi list",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			i, err := iscsi.New(controllers)
			if err != nil {
				return fmt.Errorf("failed to initialize iscsi: %w", err)
			}
			cfgs, err := i.List(context.Background())
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"IQN", "Service IP", "Service state", "LUN", "LINSTOR state"})
			table.SetHeaderColor(tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader)

			for _, cfg := range cfgs {
				for _, vol := range cfg.Status.Volumes {
					table.Rich(
						[]string{cfg.IQN.String(), cfg.ServiceIP.String(), cfg.Status.Service.String(), strconv.Itoa(vol.Number), vol.State.String()},
						[]tablewriter.Colors{{}, {}, ServiceStateColor(cfg.Status.Service), {}, ResourceStateColor(vol.State)},
					)
				}
			}

			table.SetAutoMergeCellsByColumnIndex([]int{0, 1})
			table.SetAutoFormatHeaders(false)
			table.Render()

			return nil
		},
	}
}

func startISCSICommand() *cobra.Command {
	var iqn iscsi.Iqn

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts an iSCSI target",
		Long: `Sets the target role attribute of a Pacemaker primitive to started.
In case it does not start use your favourite pacemaker tools to analyze
the root cause.`,
		Example: "linstor-gateway iscsi start --iqn=iqn.2019-08.com.linbit:example",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			i, err := iscsi.New(controllers)
			if err != nil {
				return fmt.Errorf("failed to initialize iscsi: %w", err)
			}
			cfg, err := i.Start(context.Background(), iqn)
			if err != nil {
				return err
			}

			if cfg == nil {
				return errors.New(fmt.Sprintf("Unknown target %s", iqn))
			}

			fmt.Printf("Started target %s\n", iqn)

			return nil
		},
	}

	startCmd.Flags().VarP(&iqn, "iqn", "i", "Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique)")

	startCmd.MarkFlagRequired("iqn")

	return startCmd
}

func stopISCSICommand() *cobra.Command {
	var iqn iscsi.Iqn

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stops an iSCSI target",
		Long: `Sets the target role attribute of a Pacemaker primitive to stopped.
This causes pacemaker to stop the components of an iSCSI target.

For example:
linstor-gateway iscsi stop --iqn=iqn.2019-08.com.linbit:example`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			i, err := iscsi.New(controllers)
			if err != nil {
				return fmt.Errorf("failed to initialize iscsi: %w", err)
			}
			cfg, err := i.Stop(context.Background(), iqn)
			if err != nil {
				return err
			}

			if cfg == nil {
				return errors.New(fmt.Sprintf("Unknown target %s", iqn))
			}

			fmt.Printf("Stopped target %s\n", iqn)

			return nil
		},
	}

	stopCmd.Flags().VarP(&iqn, "iqn", "i", "Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)")

	stopCmd.MarkFlagRequired("iqn")

	return stopCmd
}

func deleteISCSICommand() *cobra.Command {
	var iqn iscsi.Iqn
	var lun int

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Deletes an iSCSI target",
		Long: `Deletes an iSCSI target by stopping and deleting the pacemaker resource
primitives and removing the linstor resources.`,
		Example: "linstor-gateway iscsi delete --iqn=iqn.2019-08.com.linbit:example --lun=1",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			i, err := iscsi.New(controllers)
			if err != nil {
				return fmt.Errorf("failed to initialize iscsi: %w", err)
			}
			if cmd.Flags().Changed("lun") {
				cfg, err := i.DeleteVolume(ctx, iqn, lun)
				if err != nil {
					return err
				}

				if cfg == nil {
					return errors.New(fmt.Sprintf("Unknown target %s\n", iqn))
				} else {
					fmt.Printf("Deleted LU %d for target %s\n", lun, iqn)
				}
			} else {
				err := i.Delete(ctx, iqn)
				if err != nil {
					return err
				}

				fmt.Printf("Deleted target %s\n", iqn)
			}

			return nil
		},
	}

	deleteCmd.Flags().VarP(&iqn, "iqn", "i", "The iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) of the target to delete.")
	deleteCmd.Flags().IntVarP(&lun, "lun", "l", -1, "Set the LUN")

	deleteCmd.MarkFlagRequired("iqn")

	return deleteCmd
}
