package upgrade

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/nfs"
)

func upgradeNfs(ctx context.Context, linstor *client.Client, name string, forceYes bool, dryRun bool) (bool, error) {
	const gatewayConfigPath = "/etc/drbd-reactor.d/linstor-gateway-nfs-%s.toml"
	cfg, resourceDefinition, volumeDefinitions, resources, err := parseExistingConfig(ctx, linstor, fmt.Sprintf(gatewayConfigPath, name))
	if err != nil {
		return false, err
	}

	parsedCfg, err := nfs.FromPromoter(cfg, resourceDefinition, volumeDefinitions)
	if err != nil {
		return false, fmt.Errorf("failed to parse config: %w", err)
	}
	newConfig, err := parsedCfg.ToPromoter(resources)
	if err != nil {
		return false, fmt.Errorf("failed to generate config: %w", err)
	}

	return maybeWriteNewConfig(ctx, linstor, cfg, newConfig, forceYes, dryRun)
}

func Nfs(ctx context.Context, linstor *client.Client, name string, forceYes bool, dryRun bool) error {
	var didAny bool
	didDrbd, err := upgradeDrbdOptions(ctx, linstor, name, forceYes, dryRun)
	if err != nil {
		return fmt.Errorf("failed to upgrade drbd options: %w", err)
	}
	if didDrbd {
		didAny = true
	}
	didNfs, err := upgradeNfs(ctx, linstor, name, forceYes, false)
	if err != nil {
		return fmt.Errorf("failed to upgrade promoter config: %w", err)
	}
	if didNfs {
		didAny = true
	}
	if !didAny {
		fmt.Printf("%s is already up to date.\n", name)
	}
	return nil
}
