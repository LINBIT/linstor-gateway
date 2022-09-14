package upgrade

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

func upgradeNvmeOf(ctx context.Context, linstor *client.Client, name string, forceYes, dryRun bool) (bool, error) {
	const gatewayConfigPath = "/etc/drbd-reactor.d/linstor-gateway-nvmeof-%s.toml"
	cfg, resourceDefinition, volumeDefinitions, resources, err := parseExistingConfig(ctx, linstor, fmt.Sprintf(gatewayConfigPath, name))
	if err != nil {
		return false, err
	}

	parsedCfg, err := nvmeof.FromPromoter(cfg, resourceDefinition, volumeDefinitions)
	if err != nil {
		return false, fmt.Errorf("failed to parse config: %w", err)
	}
	newConfig, err := parsedCfg.ToPromoter(resources)
	if err != nil {
		return false, fmt.Errorf("failed to generate config: %w", err)
	}

	return maybeWriteNewConfig(ctx, linstor, cfg, newConfig, forceYes, dryRun)
}

func NvmeOf(ctx context.Context, linstor *client.Client, nqn nvmeof.Nqn, forceYes, dryRun bool) error {
	var didAny bool
	didDrbd, err := upgradeDrbdOptions(ctx, linstor, nqn.Subsystem(), forceYes, dryRun)
	if err != nil {
		return fmt.Errorf("failed to upgrade drbd options: %w", err)
	}
	if didDrbd {
		didAny = true
	}
	didNvme, err := upgradeNvmeOf(ctx, linstor, nqn.Subsystem(), forceYes, dryRun)
	if err != nil {
		return fmt.Errorf("failed to upgrade promoter config: %w", err)
	}
	if didNvme {
		didAny = true
	}
	if !didAny {
		fmt.Printf("%s is already up to date.\n", nqn.Subsystem())
	}
	return nil
}
