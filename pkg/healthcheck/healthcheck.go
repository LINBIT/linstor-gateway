package healthcheck

import (
	"errors"
	"fmt"

	"github.com/fatih/color"

	"github.com/LINBIT/linstor-gateway/client"
)

var bold = color.New(color.Bold).SprintfFunc()
var errNotFound = errors.New("not found")

type checker interface {
	check(prevError bool) error
	format(err error) string
}

func category(name string, checks ...checker) error {
	var prevError bool
	var msgs []string
	for _, c := range checks {
		err := c.check(prevError)
		if err != nil {
			prevError = true
			msgs = append(msgs, c.format(err))
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

func containsAll(haystack []string, needles []string) bool {
	for _, n := range needles {
		if !contains(haystack, n) {
			return false
		}
	}
	return true
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

func toMap(slice []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, s := range slice {
		m[s] = struct{}{}
	}
	return m
}

func checkAgent(iscsiBackends []string) error {
	errs := 0
	if err := category(
		"System Utilities",
		&checkInPath{binary: "iptables", packageName: "iptables"},
	); err != nil {
		errs++
	}
	err := category(
		"LINSTOR",
		&checkFileWhitelist{},
	)
	if err != nil {
		errs++
	}
	err = category(
		"drbd-reactor",
		&checkInPath{binary: "drbd-reactor", packageName: "drbd-reactor"},
		&checkStartedAndEnabled{"drbd-reactor.service", "drbd-reactor"},
		&checkReactorAutoReload{},
	)
	if err != nil {
		errs++
	}
	err = category(
		"Resource Agents",
		&checkFileExists{filename: "/usr/lib/ocf/resource.d/heartbeat", packageName: "resource-agents", isDirectory: true},
		&checkFileExists{
			filename:    "/usr/lib/ocf/resource.d/heartbeat/nvmet-subsystem",
			packageName: "resource-agents",
			hint:        "The nvmet-* resource agents are only shipped with resource-agents 4.9.0 or later. See https://github.com/ClusterLabs/resource-agents for instructions on how to manually install a newer version.",
		},
	)
	if err != nil {
		errs++
	}

	var iscsiChecks []checker
	backendsMap := toMap(iscsiBackends)
	for backend := range backendsMap {
		switch backend {
		case "lio-t":
			iscsiChecks = append(iscsiChecks,
				&checkInPath{binary: "targetcli", packageName: "targetcli", hint: "targetcli is only required for the LIO backend. If you are not planning on using LIO, try excluding it via `--iscsi-backends`."},
			)
		case "scst":
			iscsiChecks = append(iscsiChecks,
				&checkInPath{binary: "scstadmin", packageName: "scstadmin", hint: "scstadmin is only required for the SCST backend. If you are not planning on using SCST, try excluding it via `--iscsi-backends`."},
				&checkKernelModuleLoaded{"scst", "scst"},
				&checkKernelModuleLoaded{"iscsi_scst", "scst"},
				&checkKernelModuleLoaded{"scst_vdisk", "scst"},
				&checkProcessRunning{"iscsi-scstd", "scst"},
			)
		}
	}
	if err := category("iSCSI", iscsiChecks...); err != nil {
		errs++
	}
	err = category(
		"NVMe-oF",
		&checkInPath{binary: "nvmetcli", packageName: "nvmetcli", hint: "nvmetcli is not (yet) packaged on all distributions. See https://git.infradead.org/users/hch/nvmetcli.git for instructions on how to manually install it."},
		&checkKernelModuleLoaded{"nvmet", "nvmetcli"},
	)
	if err != nil {
		errs++
	}
	err = category(
		"NFS",
		&checkNotStartedButLoaded{"nfs-server.service", "nfs-server"},
	)
	if err != nil {
		errs++
	}
	if errs > 0 {
		return fmt.Errorf("found %d issues", errs)
	}
	return nil
}

func checkServer(controllers []string) error {
	errs := 0
	err := category(
		"LINSTOR",
		&checkLinstor{controllers},
	)
	if err != nil {
		errs++
	}
	if errs > 0 {
		return fmt.Errorf("found %d issues", errs)
	}
	return nil
}

func checkClient(cli *client.Client) error {
	errs := 0
	err := category(
		"Server Connection",
		&checkGatewayServerConnection{cli},
	)
	if err != nil {
		errs++
	}
	if errs > 0 {
		return fmt.Errorf("found %d issues", errs)
	}
	return nil
}

func CheckRequirements(mode string, iscsiBackends []string, controllers []string, cli *client.Client) error {
	doPrint := func() {
		fmt.Printf("Checking %s requirements.\n\n", bold(mode))
	}
	switch mode {
	case "agent":
		doPrint()
		return checkAgent(iscsiBackends)
	case "server":
		doPrint()
		return checkServer(controllers)
	case "client":
		doPrint()
		return checkClient(cli)
	default:
		return fmt.Errorf("unknown mode %q. Expected \"agent\", \"server\", or \"client\"", mode)
	}
}
