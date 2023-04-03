package healthcheck

import (
	"errors"
	"fmt"
	"github.com/LINBIT/linstor-gateway/client"
	"github.com/fatih/color"
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

func checkAgent() error {
	errs := 0
	err := category(
		"LINSTOR",
		&checkFileWhitelist{},
	)
	if err != nil {
		errs++
	}
	err = category(
		"drbd-reactor",
		&checkInPath{"drbd-reactor", "drbd-reactor"},
		&checkStartedAndEnabled{"drbd-reactor.service", "drbd-reactor"},
		&checkReactorAutoReload{},
	)
	if err != nil {
		errs++
	}
	err = category(
		"Resource Agents",
		&checkFileExists{"/usr/lib/ocf/resource.d/heartbeat", "resource-agents", true},
	)
	if err != nil {
		errs++
	}
	err = category(
		"iSCSI",
		&checkInPath{"targetcli", "targetcli"},
	)
	if err != nil {
		errs++
	}
	err = category(
		"NVMe-oF",
		&checkInPath{"nvmetcli", "nvmetcli"},
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

func CheckRequirements(mode string, controllers []string, cli *client.Client) error {
	doPrint := func() {
		fmt.Printf("Checking %s requirements.\n\n", bold(mode))
	}
	switch mode {
	case "agent":
		doPrint()
		return checkAgent()
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
