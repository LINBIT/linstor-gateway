package healthcheck

import (
	"context"
	"fmt"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/fatih/color"
	"github.com/pelletier/go-toml"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

const satelliteConfigFile = "/etc/linstor/linstor_satellite.toml"

type checkLinstor struct {
	controllers []string
}

func (c *checkLinstor) check(bool) error {
	ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()
	cli, err := linstorcontrol.Default(c.controllers)
	if err != nil {
		return err
	}
	_, err = cli.Controller.GetVersion(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *checkLinstor) format(err error) string {
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

type checkFileWhitelist struct {
}

func (c *checkFileWhitelist) check(bool) error {
	f, err := os.Open(satelliteConfigFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var satelliteConfig struct {
		Files struct {
			AllowExtFiles []string `toml:"allowExtFiles"`
		} `toml:"files"`
	}
	err = toml.NewDecoder(f).Decode(&satelliteConfig)
	if err != nil {
		return fmt.Errorf("failed to decode satellite config: %w", err)
	}

	expect := []string{
		"/etc/systemd/system", "/etc/systemd/system/linstor-satellite.service.d", "/etc/drbd-reactor.d",
	}
	if !containsAll(satelliteConfig.Files.AllowExtFiles, expect) {
		return fmt.Errorf("unexpected allowExtFiles value")
	}
	return nil
}

func (c *checkFileWhitelist) format(err error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "    %s The LINSTOR satellite is not configured correctly on this node\n", color.RedString("✗"))
	fmt.Fprintf(&b, "      %s\n", err.Error())
	fmt.Fprintf(&b, "      Edit the LINSTOR satellite configuration file (%s) to include the following:\n\n", bold(satelliteConfigFile))
	fmt.Fprintf(&b, "      [files]\n")
	fmt.Fprintf(&b, `        allowExtFiles = ["/etc/systemd/system", "/etc/systemd/system/linstor-satellite.service.d", "/etc/drbd-reactor.d"]`+"\n\n")
	fmt.Fprintf(&b, "      and execute %s.\n", bold("systemctl restart linstor-satellite.service"))
	return b.String()
}
