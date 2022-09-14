package upgrade

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

func upgradeIscsi(ctx context.Context, linstor *client.Client, name string, forceYes bool, dryRun bool) (bool, error) {
	const gatewayConfigPath = "/etc/drbd-reactor.d/linstor-gateway-iscsi-%s.toml"
	cfg, resourceDefinition, volumeDefinitions, resources, err := parseExistingConfig(ctx, linstor, fmt.Sprintf(gatewayConfigPath, name))
	if err != nil {
		return false, err
	}

	parsedCfg, err := iscsi.FromPromoter(cfg, resourceDefinition, volumeDefinitions)
	if err != nil {
		return false, fmt.Errorf("failed to parse config: %w", err)
	}
	newConfig, err := parsedCfg.ToPromoter(resources)
	if err != nil {
		return false, fmt.Errorf("failed to generate config: %w", err)
	}

	return maybeWriteNewConfig(ctx, linstor, cfg, newConfig, forceYes, dryRun)
}

func Iscsi(ctx context.Context, linstor *client.Client, iqn iscsi.Iqn, forceYes bool, dryRun bool) error {
	var didAny bool
	didDrbd, err := upgradeDrbdOptions(ctx, linstor, iqn.WWN(), forceYes, dryRun)
	if err != nil {
		return fmt.Errorf("failed to upgrade drbd options: %w", err)
	}
	if didDrbd {
		didAny = true
	}
	didIscsi, err := upgradeIscsi(ctx, linstor, iqn.WWN(), forceYes, dryRun)
	if err != nil {
		return fmt.Errorf("failed to upgrade promoter config: %w", err)
	}
	if didIscsi {
		didAny = true
	}
	if !didAny {
		fmt.Printf("%s is already up to date.\n", iqn.WWN())
	}
	return nil
}
