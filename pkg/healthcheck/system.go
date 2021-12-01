package healthcheck

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"strings"
	"time"
)

type checkStartedAndEnabled struct {
	service     string
	packageName string
}

func unitStartedAndEnabled(service string) error {
	ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()
	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to systemd: %w", err)
	}
	defer conn.Close()

	statuses, err := conn.ListUnitsByNamesContext(ctx, []string{service})
	if err != nil {
		return err
	}

	if len(statuses) == 0 {
		return errNotFound
	}
	if statuses[0].ActiveState != "active" {
		return fmt.Errorf("service %s is not started", service)
	}
	return nil
}

func (c *checkStartedAndEnabled) check() error {
	return unitStartedAndEnabled(c.service)
}

func (c *checkStartedAndEnabled) format(err error) string {
	var b strings.Builder
	var what string
	if errors.Is(err, errNotFound) {
		what = "installed"
	} else {
		what = "running"
	}
	fmt.Fprintf(&b, "    %s Service %s is not %s\n", color.RedString("✗"), bold(c.service), what)
	fmt.Fprintf(&b, "      Make sure that:\n")
	fmt.Fprintf(&b, "      • the %s package is installed\n", bold(c.packageName))
	fmt.Fprintf(&b, "      • the %s systemd unit is started and enabled\n", bold(c.service))
	return b.String()
}

type checkNotStarted struct {
	service string
}

func (c *checkNotStarted) check() error {
	err := unitStartedAndEnabled(c.service)
	if err != nil {
		return nil
	}
	return fmt.Errorf("is started")
}

func (c *checkNotStarted) format(_ error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "    %s Service %s is started, but it is not supposed to be.\n", color.RedString("✗"), bold(c.service))
	fmt.Fprintf(&b, "      This systemd service conflicts with LINSTOR Gateway.\n")
	fmt.Fprintf(&b, "      Execute %s to disable and stop the service.\n", bold("systemctl disable --now %s", c.service))
	return b.String()
}

type checkFileExists struct {
	filename    string
	packageName string
	isDirectory bool
}

func (c *checkFileExists) check() error {
	_, err := os.Stat(c.filename)
	return err
}

func (c *checkFileExists) format(_ error) string {
	what := "file"
	if c.isDirectory {
		what = "directory"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "    %s The %s %s does not exist\n", color.RedString("✗"), what, bold(c.filename))
	fmt.Fprintf(&b, "      Please install the %s package\n", bold(c.packageName))
	return b.String()
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

type errKernelModuleNotLoaded struct {
}

func (e *errKernelModuleNotLoaded) Error() string {
	return "kernel module is not loaded"
}

type checkKernelModuleLoaded struct {
	module      string
	packageName string
}

func (c *checkKernelModuleLoaded) check() error {
	modules, err := lsmod()
	if err != nil {
		return err
	}
	if !contains(modules, c.module) {
		return &errKernelModuleNotLoaded{}
	}
	return nil
}

func (c *checkKernelModuleLoaded) format(err error) string {
	var b strings.Builder
	if _, ok := err.(*errKernelModuleNotLoaded); ok {
		fmt.Fprintf(&b, "    %s Kernel module %s is not loaded\n", color.RedString("✗"), bold(c.module))
		fmt.Fprintf(&b, "      Execute %s or install package %s\n", bold("modprobe %s", c.module), bold(c.packageName))
		return b.String()
	}
	fmt.Fprintf(&b, "    %s Could not check if kernel module %s is loaded\n", color.RedString("✗"), bold(c.module))
	fmt.Fprintf(&b, "      %s\n", err.Error())
	return b.String()
}

type checkInPath struct {
	binary      string
	packageName string
}

func (c *checkInPath) check() error {
	_, err := exec.LookPath(c.binary)
	return err
}

func (c *checkInPath) format(err error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "    %s The %s tool is not available\n", color.RedString("✗"), bold(c.binary))
	fmt.Fprintf(&b, "      %s\n", err.Error())
	fmt.Fprintf(&b, "      Please install the %s package\n", bold(c.packageName))
	return b.String()
}
