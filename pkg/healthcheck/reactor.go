package healthcheck

import (
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type checkReactorAutoReload struct {
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
	return nil
}

func guessReactorReloadDir() string {
	paths := []string{
		"/usr/share/doc/drbd-reactor/examples",
		"/usr/share/doc/drbd-reactor",
		"/usr/share/doc/drbd-reactor-*/examples",
		"/usr/share/doc/drbd-reactor-*",
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

func (c *checkReactorAutoReload) format(_ error) string {
	dir := guessReactorReloadDir()
	var b strings.Builder
	fmt.Fprintf(&b, "    %s drbd-reactor is not configured to automatically reload\n", color.RedString("✗"))
	if dir != "" {
		path := filepath.Join(dir, "drbd-reactor-reload.{path,service}")
		fmt.Fprintf(&b, "      Please execute:\n")
		fmt.Fprintf(&b, "        %s\n", bold("cp %s /etc/systemd/system/", path))
		fmt.Fprintf(&b, "        %s\n", bold("systemctl enable --now drbd-reactor-reload.path"))
	}
	fmt.Fprintf(&b, "      Learn more at https://github.com/LINBIT/drbd-reactor/#automatic-reload\n")
	return b.String()
}
