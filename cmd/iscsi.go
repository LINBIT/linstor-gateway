package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/prompt"
	"github.com/LINBIT/linstor-gateway/pkg/upgrade"
	"github.com/LINBIT/linstor-gateway/pkg/version"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/olekukonko/tablewriter"
	"github.com/rck/unit"
	"github.com/spf13/cobra"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

var bold = color.New(color.Bold).SprintfFunc()

func iscsiCommands() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:     "iscsi",
		Version: version.Version,
		Short:   "Manages Highly-Available iSCSI targets",
		Long: `linstor-gateway iscsi manages highly available iSCSI targets by leveraging
LINSTOR and drbd-reactor. Setting up LINSTOR, including storage pools and resource groups,
as well as drbd-reactor is a prerequisite to use this tool.`,
		Args: cobra.NoArgs,
	}

	rootCmd.DisableAutoGenTag = true

	rootCmd.AddCommand(createISCSICommand())
	rootCmd.AddCommand(deleteISCSICommand())
	rootCmd.AddCommand(listISCSICommand())
	rootCmd.AddCommand(startISCSICommand())
	rootCmd.AddCommand(stopISCSICommand())
	rootCmd.AddCommand(addVolumeISCSICommand())
	rootCmd.AddCommand(deleteVolumeISCSICommand())
	rootCmd.AddCommand(upgradeISCSICommand())

	return rootCmd
}

func createISCSICommand() *cobra.Command {
	var username, password, group string
	var serviceIps []common.IpCidr
	var allowedInitiators []string
	var grossSize bool
	var implementation string

	cmd := &cobra.Command{
		Use:   "create IQN SERVICE_IPS [VOLUME_SIZE]...",
		Short: "Creates an iSCSI target",
		Long: `Creates a highly available iSCSI target based on LINSTOR and drbd-reactor.
At first it creates a new resource within the LINSTOR system, using the
specified resource group. The name of the linstor resources is derived
from the IQN's World Wide Name, which must be unique.
After that it creates a configuration for drbd-reactor to manage the
high availability primitives.`,
		Example: `linstor-gateway iscsi create iqn.2019-08.com.linbit:example 192.168.122.181/24 2G`,
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			iqn, err := iscsi.NewIqn(args[0])
			if err != nil {
				return fmt.Errorf("invalid IQN '%s': %w", args[0], err)
			}

			for _, ipString := range strings.Split(args[1], ",") {
				ip, err := common.ServiceIPFromString(ipString)
				if err != nil {
					return fmt.Errorf("invalid service IP '%s': %w", ipString, err)
				}
				serviceIps = append(serviceIps, ip)
			}

			var volumes []common.VolumeConfig
			for i, rawvalue := range args[2:] {
				val, err := unit.MustNewUnit(unit.DefaultUnits).ValueFromString(rawvalue)
				if err != nil {
					return err
				}

				volumes = append(volumes, common.VolumeConfig{
					Number:  i + 1,
					SizeKiB: uint64(val.Value / unit.K),
				})
			}

			var allowedInitiatorIqns []iscsi.Iqn
			for _, i := range allowedInitiators {
				iqn, err := iscsi.NewIqn(i)
				if err != nil {
					log.WithField("error", err).WithField("iqn", i).Warnf("Invalid IQN for allowed initiator, ignoring")
					continue
				}
				allowedInitiatorIqns = append(allowedInitiatorIqns, iqn)
			}

			_, err = cli.Iscsi.Create(ctx, &iscsi.ResourceConfig{
				IQN:               iqn,
				Username:          username,
				Password:          password,
				ServiceIPs:        serviceIps,
				Volumes:           volumes,
				AllowedInitiators: allowedInitiatorIqns,
				ResourceGroup:     group,
				GrossSize:         grossSize,
				Implementation:    implementation,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Created iSCSI target '%s'\n", iqn)

			return nil
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "Set the username to use for CHAP authentication")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Set the password to use for CHAP authentication")
	cmd.Flags().StringVarP(&group, "deprecated-resource-group", "g", "DfltRscGrp", "Set the LINSTOR resource group")
	_ = cmd.Flags().MarkHidden("deprecated-resource-group")
	_ = cmd.Flags().MarkShorthandDeprecated("deprecated-resource-group", "use -r instead")
	cmd.Flags().StringVarP(&group, "resource-group", "r", "DfltRscGrp", "Set the LINSTOR resource group")
	cmd.Flags().StringSliceVar(&allowedInitiators, "allowed-initiators", []string{}, "Restrict which initiator IQNs are allowed to connect to the target")
	cmd.Flags().BoolVar(&grossSize, "gross", false, "Make all size options specify gross size, i.e. the actual space used on disk")
	cmd.Flags().StringVar(&implementation, "implementation", "", `Set the iSCSI target implementation to use ("iet", "tgt", "lio", "lio-t", or "scst")`)

	return cmd
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
			cfgs, err := cli.Iscsi.GetAll(context.TODO())
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"IQN", "Service IP", "Service state", "LUN", "LINSTOR state"})
			table.SetHeaderColor(tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader)

			var degradedResources, badResources []string
			for _, cfg := range cfgs {
				serviceIpStrings := make([]string, len(cfg.ServiceIPs))
				for i := range cfg.ServiceIPs {
					serviceIpStrings[i] = cfg.ServiceIPs[i].String()
				}
				for i, vol := range cfg.Status.Volumes {
					if i == 0 {
						log.Debugf("not displaying cluster private volume: %+v", vol)
						continue
					}

					serviceStatus := cfg.Status.Service.String()
					if cfg.Status.Service == common.ServiceStateStarted && cfg.Status.Primary != "" {
						serviceStatus += " (" + cfg.Status.Primary + ")"
					}
					table.Rich(
						[]string{cfg.IQN.String(), strings.Join(serviceIpStrings, ", "), serviceStatus, strconv.Itoa(vol.Number), vol.State.String()},
						[]tablewriter.Colors{{}, {}, ServiceStateColor(cfg.Status.Service), {}, ResourceStateColor(vol.State)},
					)
					if vol.State != common.ResourceStateOK {
						id := cfg.IQN.WWN()
						if !contains(degradedResources, id) {
							degradedResources = append(degradedResources, id)
						}
					}
				}
				if len(cfg.Status.Volumes) == 0 {
					table.Rich(
						[]string{cfg.IQN.String(), strings.Join(serviceIpStrings, ", "), cfg.Status.Service.String(), "", common.ResourceStateBad.String()},
						[]tablewriter.Colors{{}, {}, ServiceStateColor(cfg.Status.Service), {}, ResourceStateColor(common.ResourceStateBad)},
					)
					badResources = append(badResources, cfg.IQN.WWN())
				}
			}

			table.SetAutoMergeCellsByColumnIndex([]int{0, 1})
			table.SetAutoFormatHeaders(false)
			table.Render()

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

func startISCSICommand() *cobra.Command {
	return &cobra.Command{
		Use:     "start IQN...",
		Short:   "Starts an iSCSI target",
		Long:    `Makes an iSCSI target available by starting it.`,
		Example: "linstor-gateway iscsi start iqn.2019-08.com.linbit:example",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var allErrs multiError
			for _, rawiqn := range args {
				iqn, err := iscsi.NewIqn(rawiqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				_, err = cli.Iscsi.Start(context.Background(), iqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				fmt.Printf("Started target \"%s\"\n", iqn)
			}

			return allErrs.Err()
		},
	}
}

func stopISCSICommand() *cobra.Command {
	return &cobra.Command{
		Use:     "stop IQN",
		Short:   "Stops an iSCSI target",
		Long:    `Disables an iSCSI target, making it unavailable to initiators while not deleting it.`,
		Example: `linstor-gateway iscsi stop iqn.2019-08.com.linbit:example`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var allErrs multiError
			for _, rawiqn := range args {
				iqn, err := iscsi.NewIqn(rawiqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				_, err = cli.Iscsi.Stop(context.Background(), iqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				fmt.Printf("Stopped target \"%s\"\n", iqn)
			}

			return allErrs.Err()
		},
	}
}

func deleteISCSICommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete IQN...",
		Short: "Deletes an iSCSI target",
		Long: `Deletes an iSCSI target by stopping and deleting the corresponding
drbd-reactor configuration and removing the LINSTOR resources. All logical units
of the target will be deleted.`,
		Example: "linstor-gateway iscsi delete iqn.2019-08.com.linbit:example",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var allErrs multiError
			for _, rawiqn := range args {
				iqn, err := iscsi.NewIqn(rawiqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				var yes bool
				if force {
					yes = true
				} else {
					fmt.Printf("%s: Deleting iSCSI target %q %s.\n",
						color.YellowString("WARNING"), iqn.String(),
						bold("and all data stored on it"))
					yes = prompt.Confirm("Continue?")
				}

				if yes {
					err = cli.Iscsi.Delete(context.Background(), iqn)
					if err != nil {
						allErrs = append(allErrs, err)
						continue
					}

					fmt.Printf("Deleted target %q\n", iqn)
				} else {
					fmt.Println("Aborted")
				}
			}

			return allErrs.Err()
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Delete without prompting for confirmation")

	return cmd
}

func addVolumeISCSICommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-volume IQN LU_NR LU_SIZE",
		Short: "Add a new logical unit to an existing iSCSI target",
		Long:  "Add a new logical unit to an existing iSCSI target. The target needs to be stopped.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			iqn, err := iscsi.NewIqn(args[0])
			if err != nil {
				return err
			}

			volNr, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}

			size, err := unit.MustNewUnit(unit.DefaultUnits).ValueFromString(args[2])
			if err != nil {
				return err
			}

			_, err = cli.Iscsi.AddLogicalUnit(context.Background(), iqn, &common.VolumeConfig{Number: volNr, SizeKiB: uint64(size.Value / unit.K)})
			if err != nil {
				return err
			}

			fmt.Printf("Added volume to \"%s\"\n", iqn)
			return nil
		},
	}
}

func deleteVolumeISCSICommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-volume IQN LU_NR",
		Short: "Delete a logical unit of an existing iSCSI target",
		Long:  "Delete a logical unit of an existing iSCSI target. The target needs to be stopped.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			iqn, err := iscsi.NewIqn(args[0])
			if err != nil {
				return err
			}

			volNr, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}

			err = cli.Iscsi.DeleteLogicalUnit(context.Background(), iqn, volNr)
			if err != nil {
				return err
			}

			fmt.Printf("Deleted volume %d of \"%s\"\n", volNr, iqn)
			return nil
		},
	}
}

func upgradeISCSICommand() *cobra.Command {
	var forceYes bool
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "upgrade IQN",
		Short: "Check existing resources and upgrade their configuration if necessary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			controllers := viper.GetStringSlice("linstor.controllers")
			cli, err := linstorcontrol.Default(controllers)
			if err != nil {
				return err
			}
			iqn, err := iscsi.NewIqn(args[0])
			if err != nil {
				return err
			}
			err = upgrade.Iscsi(cmd.Context(), cli.Client, iqn, forceYes, dryRun)
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
