package healthcheck

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"path/filepath"
	"strings"
)

type checkReactorAutoReload struct {
}

func (c *checkReactorAutoReload) check() error {
	return unitStartedAndEnabled("drbd-reactor-reload.path")
}

func guessReactorReloadDir() string {
	paths := []string{
		"/usr/share/doc/drbd-reactor/examples/",
		"/usr/share/doc/drbd-reactor/",
	}
	for _, p := range paths {
		_, err := os.Stat(p)
		if err == nil {
			return p
		}
	}
	return ""
}

func (c *checkReactorAutoReload) format(_ error) string {
	dir := guessReactorReloadDir()
	if dir == "" {
		// drbd-reactor is probably not installed, skip this hint
		return ""
	}
	path := filepath.Join(dir, "drbd-reactor-reload.{path,service}")
	var b strings.Builder
	fmt.Fprintf(&b, "    %s drbd-reactor is not configured to automatically reload\n", color.RedString("✗"))
	fmt.Fprintf(&b, "      Please execute:\n")
	fmt.Fprintf(&b, "        %s\n", bold("cp %s /etc/systemd/system/", path))
	fmt.Fprintf(&b, "        %s\n", bold("systemctl enable --now drbd-reactor-reload.path"))
	fmt.Fprintf(&b, "      Learn more at https://github.com/LINBIT/drbd-reactor/#automatic-reload\n")
	return b.String()
}
