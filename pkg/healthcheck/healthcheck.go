package healthcheck

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/fatih/color"
	"github.com/pelletier/go-toml"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"strings"
	"time"
)

var bold = color.New(color.Bold).SprintfFunc()

func checkLinstor(controllers []string) check {
	format := func(err error) string {
		var b strings.Builder
		fmt.Fprintf(&b, "    %s %s\n", color.RedString("✗"), "No connection to a LINSTOR controller")
		fmt.Fprintf(&b, "      %s\n", err.Error())
		fmt.Fprintf(&b, "      Make sure that either\n")
		fmt.Fprintf(&b, "      • the %s command line option, or\n", bold("--controllers"))
		fmt.Fprintf(&b, "      • the %s environment variable, or\n", bold("LS_CONTROLLERS"))
		fmt.Fprintf(&b, "      • the %s key in your configuration file (%s)\n", bold("linstor.controllers"), bold(viper.ConfigFileUsed()))
		fmt.Fprintf(&b, "      contain an URL to a LINSTOR controller, or that the LINSTOR controller is running on this machine.\n")
		return b.String()
	}

	return func() string {
		ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		cli, err := linstorcontrol.Default(controllers)
		if err != nil {
			return format(err)
		}
		_, err = cli.Controller.GetVersion(ctx)
		if err != nil {
			return format(err)
		}
		return ""
	}
}

var NotFoundError = errors.New("not found")

func checkStartedAndEnabled(service, packageName string) check {
	format := func(err error) string {
		if err != nil {
			var b strings.Builder
			var what string
			if errors.Is(err, NotFoundError) {
				what = "installed"
			} else {
				what = "running"
			}
			fmt.Fprintf(&b, "    %s Service %s is not %s\n", color.RedString("✗"), bold(service), what)
			fmt.Fprintf(&b, "      Make sure that:\n")
			fmt.Fprintf(&b, "      • the %s package is installed\n", bold(packageName))
			fmt.Fprintf(&b, "      • the %s systemd unit is started and enabled\n", bold(service))
			return b.String()
		}
		return ""
	}

	return func() string {
		ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		conn, err := dbus.NewSystemConnectionContext(ctx)
		if err != nil {
			return format(fmt.Errorf("failed to connect to systemd: %w", err))
		}
		defer conn.Close()

		statuses, err := conn.ListUnitsByNamesContext(ctx, []string{service})
		if err != nil {
			return format(err)
		}

		if len(statuses) == 0 {
			return format(NotFoundError)
		}
		if statuses[0].ActiveState != "active" {
			return format(fmt.Errorf("service %s is not started", service))
		}

		return format(nil)
	}
}

type check func() string

func category(name string, checks ...check) error {
	var msgs []string
	for _, c := range checks {
		msg := c()
		if msg != "" {
			msgs = append(msgs, msg)
		}
	}

	if len(msgs) > 0 {
		fmt.Printf("%s %s\n", color.YellowString("[!]"), name)
		for _, m := range msgs {
			fmt.Print(m)
		}
		return fmt.Errorf("some checks failed")
	}
	fmt.Printf("%s %s\n", color.GreenString("[✓]"), name)
	return nil
}

func resourceAgentsDirectoryExists() string {
	_, err := os.Stat("/usr/lib/ocf/resource.d")
	if err != nil {
		var b strings.Builder
		fmt.Fprintf(&b, "    %s The directory /usr/lib/ocf/resource.d does not exist\n", color.RedString("✗"))
		fmt.Fprintf(&b, "      Please install the %s package\n", bold("resource-agents"))
		return b.String()
	}
	return ""
}

func containsAll(haystack []string, needles []string) bool {
	for _, n := range needles {
		contains := false
		for _, h := range haystack {
			if n == h {
				contains = true
				break
			}
		}
		if !contains {
			return false
		}
	}
	return true
}

func checkFileWhitelist() string {
	format := func(err error) string {
		var b strings.Builder
		fmt.Fprintf(&b, "    %s The LINSTOR satellite is not configured correctly on this node\n", color.RedString("✗"))
		fmt.Fprintf(&b, "      %s\n", err.Error())
		fmt.Fprintf(&b, "      Whitelist the correct file paths\n")
		return b.String()
	}

	f, err := os.Open("/etc/linstor/linstor_satellite.toml")
	if err != nil {
		return format(fmt.Errorf("failed to open file: %w", err))
	}
	defer f.Close()

	var satelliteConfig struct {
		Files struct {
			AllowExtFiles []string `toml:"allowExtFiles"`
		} `toml:"files"`
	}
	err = toml.NewDecoder(f).Decode(&satelliteConfig)
	if err != nil {
		return format(fmt.Errorf("failed to decode satellite config: %w", err))
	}

	expect := []string{
		"/etc/systemd/system", "/etc/systemd/system/linstor-satellite.d", "/etc/drbd-reactor.d",
	}
	if !containsAll(satelliteConfig.Files.AllowExtFiles, expect) {
		return format(fmt.Errorf("unexpected allowExtFiles value"))
	}
	return ""
}

func checkInPath(binary, packageName string) check {
	return func() string {
		_, err := exec.LookPath(binary)
		if err != nil {
			var b strings.Builder
			fmt.Fprintf(&b, "    %s The %s tool is not available\n", color.RedString("✗"), bold(binary))
			fmt.Fprintf(&b, "      %s\n", err.Error())
			fmt.Fprintf(&b, "      Please install the %s package\n", bold(packageName))
			return b.String()
		}
		return ""
	}
}

func lsmod() ([]string, error) {
	f, err := os.Open("/proc/modules")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/modules: %w", err)
	}
	defer f.Close()

	var modules []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), " ")
		if len(split) > 0 {
			modules = append(modules, split[0])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read /proc/modules: %w", err)
	}

	return modules, nil
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func checkModuleLoaded(module, packageName string) check {
	return func() string {
		modules, err := lsmod()
		if err != nil {
			var b strings.Builder
			fmt.Fprintf(&b, "    %s Could not check if kernel module %s is loaded\n", color.RedString("✗"), bold(module))
			fmt.Fprintf(&b, "      %s\n", err.Error())
			return b.String()
		}
		if !contains(modules, module) {
			var b strings.Builder
			fmt.Fprintf(&b, "    %s Kernel module %s is not loaded\n", color.RedString("✗"), bold(module))
			fmt.Fprintf(&b, "      Execute %s or install package %s\n", bold("modprobe %s", module), bold(packageName))
			return b.String()
		}
		return ""
	}
}

func CheckRequirements(controllers []string) error {
	errs := 0
	err := category(
		"LINSTOR",
		checkLinstor(controllers),
		checkFileWhitelist,
	)
	if err != nil {
		errs++
	}
	err = category("drbd-reactor", checkStartedAndEnabled("drbd-reactor.service", "drbd-reactor"))
	if err != nil {
		errs++
	}
	err = category("Resource Agents", resourceAgentsDirectoryExists)
	if err != nil {
		errs++
	}
	err = category("iSCSI", checkInPath("targetcli", "targetcli"))
	if err != nil {
		errs++
	}
	err = category(
		"NVMe-oF",
		checkInPath("nvmetcli", "nvmetcli"),
		checkModuleLoaded("nvmet", "nvmetcli"),
	)
	if err != nil {
		errs++
	}
	err = category("NFS", checkStartedAndEnabled("nfs-server.service", "nfs-utils"))
	if err != nil {
		errs++
	}
	if errs > 0 {
		return fmt.Errorf("found %d issues", errs)
	}
	return nil
}
