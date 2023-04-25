package upgrade

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/LINBIT/linstor-gateway/pkg/nfs"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
)

// nfsMigrations defines the operations for upgrading a single version.
// The array index ("n") denotes the starting version; the function at
// index "n" migrates from version "n" to version "n+1".
var nfsMigrations = []func(cfg *reactor.PromoterConfig) error{
	0: initialMigrations,
}

// initialMigrations combines two migrations that make up version 1.
func initialMigrations(cfg *reactor.PromoterConfig) error {
	if err := removeID(cfg); err != nil {
		return err
	}
	return switchIpAndNfsServer(cfg)
}

// switchIpAndNfsServer changes the order of the resource agents such that
// the service IP is started before the NFS server. This was previously not
// the case, leading to an NFS server that is not limited to the service IP.
func switchIpAndNfsServer(cfg *reactor.PromoterConfig) error {
	id := firstResourceId(cfg)
	firstResource := cfg.Resources[id]
	var serviceIpIndex, nfsServerIndex int
	for i, entry := range firstResource.Start {
		switch agent := entry.(type) {
		case *reactor.ResourceAgent:
			if agent.Type == "ocf:heartbeat:IPaddr2" {
				serviceIpIndex = i
			}
			if agent.Type == "ocf:heartbeat:nfsserver" {
				nfsServerIndex = i
			}
		}
	}

	// slice up the "start" list and create a new one with this order:
	// 1. everything before the nfs server
	// 2. the service IP
	// 3. the NFS server
	// 4. all the exportfs entries
	// 5. everything after the original service IP index
	start := firstResource.Start[:nfsServerIndex]
	serviceIpEntry := firstResource.Start[serviceIpIndex]
	nfsServerEntry := firstResource.Start[nfsServerIndex]
	exportFsEntries := firstResource.Start[nfsServerIndex+1 : serviceIpIndex]
	rest := firstResource.Start[serviceIpIndex+1:]

	var s []reactor.StartEntry
	s = append(s, start...)
	s = append(s, serviceIpEntry)
	s = append(s, nfsServerEntry)
	s = append(s, exportFsEntries...)
	s = append(s, rest...)
	firstResource.Start = s
	cfg.Resources[id] = firstResource
	return nil
}

func upgradeNfs(ctx context.Context, linstor *client.Client, name string, forceYes bool, dryRun bool) (bool, error) {
	const gatewayConfigPath = "/etc/drbd-reactor.d/linstor-gateway-nfs-%s.toml"
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

	for i := cfg.Metadata.LinstorGatewaySchemaVersion; i < nfs.CurrentVersion; i++ {
		err := nfsMigrations[i](newCfg)
		if err != nil {
			return false, fmt.Errorf("failed to migrate from version %d to %d: %w", i, i+1, err)
		}
	}
	newCfg.Metadata.LinstorGatewaySchemaVersion = nfs.CurrentVersion

	return maybeWriteNewConfig(ctx, linstor, cfg, newCfg, forceYes, dryRun)
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
	didNfs, err := upgradeNfs(ctx, linstor, name, forceYes, dryRun)
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
