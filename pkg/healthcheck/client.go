package healthcheck

import (
	"context"
	"fmt"
	"github.com/LINBIT/linstor-gateway/client"
	"github.com/fatih/color"
	"strings"
	"time"
)

type checkGatewayServerConnection struct {
	cli *client.Client
}

func (c *checkGatewayServerConnection) check(bool) error {
	ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()
	status, err := c.cli.Status.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	if status == nil {
		return fmt.Errorf("received nil status from server")
	}
	if status.Status != "ok" {
		return fmt.Errorf("received invalid status from server: %q", status.Status)
	}
	return nil
}

func (c *checkGatewayServerConnection) format(err error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "    %s The LINSTOR Gateway server cannot be reached from this node\n", color.RedString("✗"))
	fmt.Fprintf(&b, "      %s\n\n", err.Error())
	fmt.Fprintf(&b, "      Make sure the %s command line option points to a running LINSTOR Gateway server.\n", bold("--connect"))
	return b.String()
}
