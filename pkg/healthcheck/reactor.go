package healthcheck

import (
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

type checkReactorAutoReload struct {
}

func fileContentsEqual(filenameA, filenameB string) (bool, error) {
	a, err := os.ReadFile(filenameA)
	if err != nil {
		return false, fmt.Errorf("could not read %s: %w", filenameA, err)
	}

	b, err := os.ReadFile(filenameB)
	if err != nil {
		return false, fmt.Errorf("could not read %s: %w", filenameB, err)
	}

	return string(a) == string(b), nil
}

func (c *checkReactorAutoReload) check(prevError bool) error {
	if prevError {
		// reactor not installed, no need to check
		return nil
	}
	status, err := unitStatus("drbd-reactor-reload.path")
	if err != nil {
		return err
	}
	if status.ActiveState != "active" {
		return fmt.Errorf("service drbd-reactor-reload.path is not started")
	}

	dir := guessReactorReloadDir()
	sourcePath := filepath.Join(dir, "drbd-reactor-reload.path")
	sourceService := filepath.Join(dir, "drbd-reactor-reload.service")

	destPath := "/etc/systemd/system/drbd-reactor-reload.path"
	destService := "/etc/systemd/system/drbd-reactor-reload.service"

	for _, path := range [][]string{{sourcePath, destPath}, {sourceService, destService}} {
		equal, err := fileContentsEqual(path[0], path[1])
		if err != nil {
			return fmt.Errorf("could not compare %s and %s: %w", path[0], path[1], err)
		}
		if !equal {
			return fmt.Errorf("%s differs from %s", path[0], path[1])
		}
	}

	return nil
}

func guessReactorReloadDir() string {
	paths := []string{
		"/usr/share/doc/drbd-reactor/examples",
		"/usr/share/doc/drbd-reactor",
		"/usr/share/doc/drbd-reactor-*/examples",
		"/usr/share/doc/drbd-reactor-*",
		"/usr/share/doc/packages/drbd-reactor/examples",
		"/usr/share/doc/packages/drbd-reactor",
		"/usr/share/doc/packages/drbd-reactor-*/examples",
		"/usr/share/doc/packages/drbd-reactor-*",
	}
	for _, p := range paths {
		matches, err := filepath.Glob(p)
		if err != nil {
			log.Debugf("Glob failed: %v", err)
			continue
		}
		log.Debugf("Glob %s -> %v", p, matches)
		if len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}

func (c *checkReactorAutoReload) format(err error) string {
	dir := guessReactorReloadDir()
	var b strings.Builder
	fmt.Fprintf(&b, "    %s drbd-reactor is not configured to automatically reload\n", color.RedString("✗"))
	fmt.Fprintf(&b, "      %s\n", faint("→ %s", err.Error()))
	if dir != "" {
		path := filepath.Join(dir, "drbd-reactor-reload.{path,service}")
		fmt.Fprintf(&b, "      Please execute:\n")
		fmt.Fprintf(&b, "        %s\n", bold("cp %s /etc/systemd/system/", path))
		fmt.Fprintf(&b, "        %s\n", bold("systemctl enable --now drbd-reactor-reload.path"))
	}
	fmt.Fprintf(&b, "      Learn more at https://github.com/LINBIT/drbd-reactor/#automatic-reload\n")
	return b.String()
}
