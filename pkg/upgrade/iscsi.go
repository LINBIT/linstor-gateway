package upgrade

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
)

// iscsiMigrations defines the operations for upgrading a single version.
// The array index ("n") denotes the starting version; the function at
// index "n" migrates from version "n" to version "n+1".
var iscsiMigrations = []func(cfg *reactor.PromoterConfig) error{
	0: removeID,
}

func upgradeIscsi(ctx context.Context, linstor *client.Client, name string, forceYes bool, dryRun bool) (bool, error) {
	const gatewayConfigPath = "/etc/drbd-reactor.d/linstor-gateway-iscsi-%s.toml"
	cfg, _, _, _, err := parseExistingConfig(ctx, linstor, fmt.Sprintf(gatewayConfigPath, name))
	if err != nil {
		return false, err
	}
	if cfg.Metadata.LinstorGatewaySchemaVersion > iscsi.CurrentVersion {
		return false, fmt.Errorf("schema version %d is not supported",
			cfg.Metadata.LinstorGatewaySchemaVersion)
	}
	newCfg, _, _, _, err := parseExistingConfig(ctx, linstor, fmt.Sprintf(gatewayConfigPath, name))
	if err != nil {
		return false, err
	}

	for i := cfg.Metadata.LinstorGatewaySchemaVersion; i < iscsi.CurrentVersion; i++ {
		err := iscsiMigrations[i](newCfg)
		if err != nil {
			return false, fmt.Errorf("failed to migrate from version %d to %d: %w", i, i+1, err)
		}
	}
	newCfg.Metadata.LinstorGatewaySchemaVersion = iscsi.CurrentVersion

	return maybeWriteNewConfig(ctx, linstor, cfg, newCfg, forceYes, dryRun)
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
