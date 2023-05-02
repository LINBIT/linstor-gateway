/*
Package upgrade migrates existing resources to the latest version.
We will always try to upgrade to the most modern version of the drbd-reactor configuration.
*/
package upgrade

import (
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/prompt"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
	"github.com/google/go-cmp/cmp"
	"github.com/pelletier/go-toml"
	"github.com/sergi/go-diff/diffmatchpatch"
	"strings"
)

func encode(cfg *reactor.PromoterConfig) (string, error) {
	buffer := strings.Builder{}
	encoder := toml.NewEncoder(&buffer).ArraysWithOneElementPerLine(true)

	err := encoder.Encode(&reactor.Config{Promoter: []reactor.PromoterConfig{*cfg}})
	if err != nil {
		return "", fmt.Errorf("error encoding toml: %w", err)
	}
	return buffer.String(), nil
}

func parseExistingConfig(ctx context.Context, linstor *client.Client, path string) (*reactor.PromoterConfig, *client.ResourceDefinition, []client.VolumeDefinition, []client.ResourceWithVolumes, error) {
	file, err := linstor.Controller.GetExternalFile(ctx, path)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch promoter config: %w", err)
	}

	var fullCfg reactor.Config
	err = toml.Unmarshal(file.Content, &fullCfg)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to decode promoter config: %w", err)
	}
	cfg := &fullCfg.Promoter[0]

	resourceDefinition, _, volumeDefinitions, resources, err := cfg.DeployedResources(ctx, linstor)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch deployed resources: %w", err)
	}
	return cfg, resourceDefinition, volumeDefinitions, resources, nil
}

func maybeWriteNewConfig(ctx context.Context, linstor *client.Client, oldConfig *reactor.PromoterConfig, newConfig *reactor.PromoterConfig, forceYes, dryRun bool) (bool, error) {
	if cmp.Equal(oldConfig, newConfig) {
		// nothing to do
		return false, nil
	}
	oldToml, err := encode(oldConfig)
	if err != nil {
		return false, fmt.Errorf("failed to encode old promoter config: %w", err)
	}
	newToml, err := encode(newConfig)
	if err != nil {
		return false, fmt.Errorf("failed to marshal new config to toml: %w", err)
	}
	fmt.Println("The following configuration changes are necessary:")
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffCleanupSemantic(dmp.DiffMain(oldToml, newToml, false))
	fmt.Println(dmp.DiffPrettyText(diffs))
	fmt.Println()
	if dryRun {
		return true, nil
	}
	if !forceYes {
		yes := prompt.Confirm("Apply this configuration now?")
		if !yes {
			// abort
			return false, fmt.Errorf("aborted")
		}
	}
	err = reactor.EnsureConfig(ctx, linstor, newConfig)
	if err != nil {
		return true, fmt.Errorf("failed to install config: %w", err)
	}
	return true, nil
}
