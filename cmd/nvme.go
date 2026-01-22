package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/LINBIT/linstor-gateway/client"
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
	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

func nvmeCommands() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:     "nvme",
		Version: version.Version,
		Short:   "Manages Highly-Available NVME targets",
		Long:    `nvme manages highly available NVME targets by leveraging LINSTOR and DRBD.`,
		Args:    cobra.NoArgs,
	}

	rootCmd.DisableAutoGenTag = true

	rootCmd.AddCommand(listNVMECommand())
	rootCmd.AddCommand(createNVMECommand())
	rootCmd.AddCommand(deleteNVMECommand())
	rootCmd.AddCommand(startNVMECommand())
	rootCmd.AddCommand(stopNVMECommand())
	rootCmd.AddCommand(addVolumeNVMECommand())
	rootCmd.AddCommand(deleteVolumeNVMECommand())
	rootCmd.AddCommand(upgradeNVMECommand())

	return rootCmd
}

func listNVMECommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list configured NVMe-oF targets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgs, err := cli.NvmeOf.GetAll(context.Background())
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"NQN", "Service IP", "Service state", "Namespace", "LINSTOR state"})
			table.SetHeaderColor(tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader)

			var degradedResources, badResources []string
			for _, cfg := range cfgs {
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
						[]string{cfg.NQN.String(), cfg.ServiceIP.String(), serviceStatus, strconv.Itoa(vol.Number), vol.State.String()},
						[]tablewriter.Colors{{}, {}, ServiceStateColor(cfg.Status.Service), {}, ResourceStateColor(vol.State)},
					)
					if vol.State != common.ResourceStateOK {
						id := cfg.NQN.Subsystem()
						if !contains(degradedResources, id) {
							degradedResources = append(degradedResources, id)
						}
					}
				}

				if len(cfg.Status.Volumes) == 0 {
					table.Rich(
						[]string{cfg.NQN.String(), cfg.ServiceIP.String(), cfg.Status.Service.String(), "", common.ResourceStateBad.String()},
						[]tablewriter.Colors{{}, {}, ServiceStateColor(cfg.Status.Service), {}, ResourceStateColor(common.ResourceStateBad)},
					)
					badResources = append(badResources, cfg.NQN.Subsystem())
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

func createNVMECommand() *cobra.Command {
	resourceGroup := "DfltRscGrp"
	grossSize := false
	var resourceTimeout time.Duration

	cmd := &cobra.Command{
		Use:     "create NQN SERVICE_IP VOLUME_SIZE [VOLUME_SIZE]...",
		Short:   "Create a new NVMe-oF target",
		Long:    `Create a new NVMe-oF target. The NQN consists of <vendor>:nvme:<subsystem>.`,
		Example: `linstor-gateway nvme create linbit:nvme:example`,
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			nqn, err := nvmeof.NewNqn(args[0])
			if err != nil {
				return err
			}

			serviceIP, err := common.ServiceIPFromString(args[1])
			if err != nil {
				return err
			}

			var volumes []common.VolumeConfig
			for i, rawvalue := range args[2:] {
				val, err := unit.MustNewUnit(unit.DefaultUnits).ValueFromString(rawvalue)
				if err != nil {
					return err
				}

				volumes = append(volumes, common.VolumeConfig{
					Number:  i + 1,
					SizeKiB: uint64(val.Value / unit.K)})
			}

			_, err = cli.NvmeOf.Create(context.Background(), &nvmeof.ResourceConfig{
				NQN:             nqn,
				ServiceIP:       serviceIP,
				ResourceGroup:   resourceGroup,
				Volumes:         volumes,
				GrossSize:       grossSize,
				ResourceTimeout: resourceTimeout,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Created target \"%s\"\n", nqn)

			return nil
		},
	}
	cmd.Flags().StringVarP(&resourceGroup, "resource-group", "r", resourceGroup, "resource group to use.")
	cmd.Flags().BoolVar(&grossSize, "gross", false, "Make all size options specify gross size, i.e. the actual space used on disk")
	cmd.Flags().DurationVar(&resourceTimeout, "resource-timeout", nvmeof.DefaultResourceTimeout, "Timeout for waiting for the resource to become available")

	return cmd
}

func deleteNVMECommand() *cobra.Command {
	var force bool
	var resourceTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "delete NQN...",
		Short: "Delete existing NVMe-oF targets",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var allErrs multiError
			for _, rawnqn := range args {
				nqn, err := nvmeof.NewNqn(rawnqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				var yes bool
				if force {
					yes = true
				} else {
					fmt.Printf("%s: Deleting NVMe-oF target %q %s.\n",
						color.YellowString("WARNING"), nqn.String(),
						bold("and all data stored on it"))
					yes = prompt.Confirm("Continue?")
				}

				if yes {
					err = cli.NvmeOf.Delete(context.Background(), nqn, resourceTimeout)
					if err == client.NotFoundError {
						allErrs = append(allErrs, noTarget(nqn))
						continue
					}
					if err != nil {
						allErrs = append(allErrs, err)
						continue
					}
					fmt.Printf("Deleted target %q\n", nqn)
				} else {
					fmt.Println("Aborted")
				}
			}

			return allErrs.Err()
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Delete without prompting for confirmation")
	cmd.Flags().DurationVar(&resourceTimeout, "resource-timeout", nvmeof.DefaultResourceTimeout, "Timeout for waiting for the resource to become unavailable")
	return cmd
}

func startNVMECommand() *cobra.Command {
	var resourceTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "start NQN...",
		Short: "Start a stopped NVMe-oF target",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var allErrs multiError
			for _, rawnqn := range args {
				nqn, err := nvmeof.NewNqn(rawnqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				_, err = cli.NvmeOf.Start(context.Background(), nqn, resourceTimeout)
				if err == client.NotFoundError {
					allErrs = append(allErrs, noTarget(nqn))
					continue
				}
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				fmt.Printf("Started target \"%s\"\n", nqn)
			}

			return allErrs.Err()
		},
	}

	cmd.Flags().DurationVar(&resourceTimeout, "resource-timeout", nvmeof.DefaultResourceTimeout, "Timeout for waiting for the resource to become available")

	return cmd
}

func stopNVMECommand() *cobra.Command {
	var resourceTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "stop NQN...",
		Short: "Stop a started NVMe-oF target",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var allErrs multiError
			for _, rawnqn := range args {
				nqn, err := nvmeof.NewNqn(rawnqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				_, err = cli.NvmeOf.Stop(context.Background(), nqn, resourceTimeout)
				if err == client.NotFoundError {
					allErrs = append(allErrs, noTarget(nqn))
					continue
				}
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				fmt.Printf("Stopped target \"%s\"\n", nqn)
			}

			return allErrs.Err()
		},
	}

	cmd.Flags().DurationVar(&resourceTimeout, "resource-timeout", nvmeof.DefaultResourceTimeout, "Timeout for waiting for the resource to become unavailable")

	return cmd
}

func addVolumeNVMECommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-volume NQN VOLUME_NR VOLUME_SIZE",
		Short: "Add a new volume to an existing NVMe-oF target",
		Long:  "Add a new volume to an existing NVMe-oF target. The target needs to be stopped.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			nqn, err := nvmeof.NewNqn(args[0])
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

			_, err = cli.NvmeOf.AddVolume(context.Background(), nqn, &common.VolumeConfig{Number: volNr, SizeKiB: uint64(size.Value / unit.K)})
			if err == client.NotFoundError {
				return noTarget(nqn)
			}
			if err != nil {
				return err
			}

			fmt.Printf("Added volume to \"%s\"\n", nqn)
			return nil
		},
	}
}

func deleteVolumeNVMECommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-volume NQN VOLUME_NR",
		Short: "Delete a volume of an existing NVMe-oF target",
		Long:  "Delete a volume of an existing NVMe-oF target. The target needs to be stopped.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			nqn, err := nvmeof.NewNqn(args[0])
			if err != nil {
				return err
			}

			volNr, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}

			err = cli.NvmeOf.DeleteVolume(context.Background(), nqn, volNr)
			if err != nil {
				if err == client.NotFoundError {
					return noTarget(nqn)
				}
				return err
			}

			fmt.Printf("Deleted volume %d of \"%s\"\n", volNr, nqn)
			return nil
		},
	}
}

func upgradeNVMECommand() *cobra.Command {
	var forceYes bool
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "upgrade NQN",
		Short: "Check existing resources and upgrade their configuration if necessary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			controllers := viper.GetStringSlice("linstor.controllers")
			cli, err := linstorcontrol.Default(controllers)
			if err != nil {
				return err
			}
			nqn, err := nvmeof.NewNqn(args[0])
			if err != nil {
				return err
			}
			err = upgrade.NvmeOf(cmd.Context(), cli.Client, nqn, forceYes, dryRun)
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

type noTarget nvmeof.Nqn

func (n noTarget) Error() string {
	return fmt.Sprintf("no target named %s", nvmeof.Nqn(n))
}

type multiError []error

func (m multiError) Error() string {
	if len(m) == 0 {
		return "<none>"
	}

	if len(m) == 1 {
		return m[0].Error()
	}

	formatted := make([]string, len(m))
	for i := range m {
		formatted[i] = m[i].Error()
	}
	return fmt.Sprintf("%d errors: [%s]", len(m), strings.Join(formatted, ", "))
}

func (m multiError) Err() error {
	if len(m) == 0 {
		return nil
	}

	return m
}
