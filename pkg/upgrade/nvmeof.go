package upgrade

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
)

// nvmeOfMigrations defines the operations for upgrading a single version.
// The array index ("n") denotes the starting version; the function at
// index "n" migrates from version "n" to version "n+1".
var nvmeOfMigrations = []func(cfg *reactor.PromoterConfig) error{
	0: removeID,
}

func upgradeNvmeOf(ctx context.Context, linstor *client.Client, name string, forceYes, dryRun bool) (bool, error) {
	const gatewayConfigPath = "/etc/drbd-reactor.d/linstor-gateway-nvmeof-%s.toml"
	cfg, _, _, _, err := parseExistingConfig(ctx, linstor, fmt.Sprintf(gatewayConfigPath, name))
	if err != nil {
		return false, err
	}
	if cfg.Metadata.LinstorGatewaySchemaVersion > nvmeof.CurrentVersion {
		return false, fmt.Errorf("schema version %d is not supported",
			cfg.Metadata.LinstorGatewaySchemaVersion)
	}
	newCfg, _, _, _, err := parseExistingConfig(ctx, linstor, fmt.Sprintf(gatewayConfigPath, name))
	if err != nil {
		return false, err
	}

	for i := cfg.Metadata.LinstorGatewaySchemaVersion; i < nvmeof.CurrentVersion; i++ {
		err := nvmeOfMigrations[i](newCfg)
		if err != nil {
			return false, fmt.Errorf("failed to migrate from version %d to %d: %w", i, i+1, err)
		}
	}
	newCfg.Metadata.LinstorGatewaySchemaVersion = nvmeof.CurrentVersion

	return maybeWriteNewConfig(ctx, linstor, cfg, newCfg, fmt.Sprintf(nvmeof.IDFormat, name), forceYes, dryRun)
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
