package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/rck/unit"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

func nvmeCommands() *cobra.Command {
	var loglevel string

	var rootCmd = &cobra.Command{
		Use:     "nvme",
		Version: version,
		Short:   "Manages Highly-Available NVME targets",
		Long:    `nvme manages highly available NVME targets by leveraging LINSTOR and DRBD.`,
		Args:    cobra.NoArgs,
		// We could have our custom flag types, but this is really simple enough...
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level, err := log.ParseLevel(loglevel)
			if err != nil {
				log.Fatal(err)
			}
			log.SetLevel(level)
		},
	}

	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", log.InfoLevel.String(), "Set the log level (as defined by logrus)")
	rootCmd.DisableAutoGenTag = true

	rootCmd.AddCommand(completionCommand(rootCmd))
	rootCmd.AddCommand(listNVMECommand())
	rootCmd.AddCommand(createNVMECommand())
	rootCmd.AddCommand(deleteNVMECommand())
	rootCmd.AddCommand(startNVMECommand())
	rootCmd.AddCommand(stopNVMECommand())
	rootCmd.AddCommand(addVolumeNVMECommand())
	rootCmd.AddCommand(deleteVolumeNVMECommand())

	return rootCmd
}

func listNVMECommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list configured NVMe-oF targets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgs, err := nvmeof.List(context.Background())
			if err != nil {
				return err
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"NQN", "Service IP", "Service state", "Namespace", "LINSTOR state"})
			table.SetHeaderColor(tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader, tableColorHeader)

			for _, cfg := range cfgs {
				for _, vol := range cfg.Status.Volumes {
					table.Rich(
						[]string{cfg.NQN.String(), cfg.ServiceIP.String(), cfg.Status.Service.String(), strconv.Itoa(vol.Number), vol.State.String()},
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

func createNVMECommand() *cobra.Command {
	resourceGroup := "DfltRscGrp"

	cmd := &cobra.Command{
		Use:   "create NQN SERVICE_IP [VOLUME_SIZE]...",
		Short: "Create a new NVMe-oF target",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			nqn, err := nvmeof.NewNqn(args[0])
			if err != nil {
				return err
			}

			serviceIP, err := common.ServiceIPFromString(args[1])
			if err != nil {
				return err
			}

			volumes := make([]common.VolumeConfig, 0, len(args[2:]))
			for _, rawvalue := range args[2:] {
				val, err := unit.MustNewUnit(unit.DefaultUnits).ValueFromString(rawvalue)
				if err != nil {
					return err
				}

				volumes = append(volumes, common.VolumeConfig{SizeKiB: uint64(val.Value / unit.K)})
			}

			_, err = nvmeof.Create(context.Background(), &nvmeof.ResourceConfig{
				NQN:           nqn,
				ServiceIP:     serviceIP,
				ResourceGroup: resourceGroup,
				Volumes:       volumes,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Created target \"%s\"\n", nqn)

			return nil
		},
	}
	cmd.Flags().StringVarP(&resourceGroup, "resource-group", "r", resourceGroup, "resource group to use.")

	return cmd
}

func deleteNVMECommand() *cobra.Command {
	return &cobra.Command{
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

				err = nvmeof.Delete(context.Background(), nqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				fmt.Printf("Deleted target \"%s\"\n", nqn)
			}

			return allErrs.Err()
		},
	}
}

func startNVMECommand() *cobra.Command {
	return &cobra.Command{
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

				cfg, err := nvmeof.Start(context.Background(), nqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				if cfg == nil {
					allErrs = append(allErrs, noTarget(nqn))
					continue
				}

				fmt.Printf("Started target \"%s\"\n", nqn)
			}

			return allErrs.Err()
		},
	}
}

func stopNVMECommand() *cobra.Command {
	return &cobra.Command{
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

				cfg, err := nvmeof.Stop(context.Background(), nqn)
				if err != nil {
					allErrs = append(allErrs, err)
					continue
				}

				if cfg == nil {
					allErrs = append(allErrs, noTarget(nqn))
					continue
				}

				fmt.Printf("Stopped target \"%s\"\n", nqn)
			}

			return allErrs.Err()
		},
	}
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

			cfg, err := nvmeof.AddVolume(context.Background(), nqn, &common.VolumeConfig{Number: volNr, SizeKiB: uint64(size.Value / unit.K)})
			if err != nil {
				return err
			}

			if cfg == nil {
				return noTarget(nqn)
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

			cfg, err := nvmeof.DeleteVolume(context.Background(), nqn, volNr)
			if err != nil {
				return err
			}

			if cfg == nil {
				return noTarget(nqn)
			}

			fmt.Printf("Deleted volume %d of \"%s\"\n", volNr, nqn)
			return nil
		},
	}
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

func volumeString(vols []common.VolumeConfig) string {
	names := make([]string, 0, len(vols))
	for _, vol := range vols {
		names = append(names, strconv.Itoa(vol.Number))
	}
	return strings.Join(names, ",")
}
