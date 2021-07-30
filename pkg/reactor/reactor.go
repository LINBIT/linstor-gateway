package reactor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/LINBIT/golinstor/client"
)

const (
	promoterDir       = "/etc/drbd-reactor.d"
	gatewayConfigPath = promoterDir + "/linstor-gateway-%s.toml"
)

// Config is the root configuration for drbd-reactor.
//
// Currently, only supports Promoter plugins.
type Config struct {
	Promoter []PromoterConfig `toml:"promoter,omitempty"`
}

// PromoterConfig is the configuration for drbd-reactors "promoter" plugin.
type PromoterConfig struct {
	ID        string                            `toml:"id"`
	Resources map[string]PromoterResourceConfig `toml:"resources,omitempty"`
}

// DeployedResources fetches the current state of the resources referenced in the promoter config.
func (p *PromoterConfig) DeployedResources(ctx context.Context, cli *client.Client) (*client.ResourceDefinition, []client.VolumeDefinition, []client.ResourceWithVolumes, error) {
	var rscNames []string
	for k := range p.Resources {
		rscNames = append(rscNames, k)
	}

	if len(rscNames) != 1 {
		return nil, nil, nil, errors.New(fmt.Sprintf("expected exactly 1 resource, got %d", len(rscNames)))
	}

	rd, err := cli.ResourceDefinitions.Get(ctx, rscNames[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch resource definition: %w", err)
	}

	vds, err := cli.ResourceDefinitions.GetVolumeDefinitions(ctx, rscNames[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch volume definition: %w", err)
	}

	resources, err := cli.Resources.GetResourceView(ctx, &client.ListOpts{Resource: rscNames})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch deployed resources: %w", err)
	}

	return &rd, vds, resources, nil
}

// PromoterResourceConfig is the configuration of a single promotable resource used by drbd-reactor's promoter.
type PromoterResourceConfig struct {
	Start              []ResourceAgent `toml:"start,omitempty"`
	Runner             string          `toml:"runner,omitempty"`
	OnStopFailure      string          `toml:"on-stop-failure,omitempty"`
	StopServicesOnExit bool            `toml:"stop-services-on-exit,omitempty"`
	TargetAs           string          `toml:"target-as,omitempty"`
}

// EnsureConfig ensures the given config is registered in LINSTOR and up-to-date.
func EnsureConfig(ctx context.Context, cli *client.Client, cfg *PromoterConfig) error {
	buffer := strings.Builder{}
	encoder := toml.NewEncoder(&buffer)

	err := encoder.Encode(&Config{Promoter: []PromoterConfig{*cfg}})
	if err != nil {
		return fmt.Errorf("error encoding promoter config: %w", err)
	}

	path := ConfigPath(cfg.ID)
	err = cli.Controller.ModifyExternalFile(ctx, path, client.ExternalFile{Path: path, Content: []byte(buffer.String())})
	if err != nil {
		return fmt.Errorf("error setting promoter config in linstor: %w", err)
	}

	return nil
}

// AttachConfig ensures the promoter config is attached to all referenced resources.
func AttachConfig(ctx context.Context, cli *client.Client, cfg *PromoterConfig) error {
	path := ConfigPath(cfg.ID)

	for rd := range cfg.Resources {
		err := cli.ResourceDefinitions.AttachExternalFile(ctx, rd, path)
		if err != nil {
			return fmt.Errorf("error attaching file to resource: %w", err)
		}
	}

	return nil
}

// DetachConfig detaches the promoter config from all resources.
func DetachConfig(ctx context.Context, cli *client.Client, cfg *PromoterConfig) error {
	path := ConfigPath(cfg.ID)

	for rd := range cfg.Resources {
		err := cli.ResourceDefinitions.DetachExternalFile(ctx, rd, path)
		if err != nil {
			return fmt.Errorf("error attaching file to resource: %w", err)
		}
	}

	return nil
}

// ListConfigs fetches all promoter configurations registered with LINSTOR.
func ListConfigs(ctx context.Context, cli *client.Client) ([]PromoterConfig, []string, error) {
	files, err := cli.Controller.GetExternalFiles(ctx, &client.ListOpts{Content: true})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch file list: %w", err)
	}

	result := make([]PromoterConfig, 0, len(files))
	paths := make([]string, 0, len(files))

	for _, file := range files {
		var name string
		n, _ := fmt.Sscanf(file.Path, gatewayConfigPath, &name)
		if n == 0 {
			continue
		}

		cfg := Config{}
		err := toml.Unmarshal(file.Content, &cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode promoter config: %w", err)
		}

		result = append(result, cfg.Promoter...)
		paths = append(paths, file.Path)
	}

	return result, paths, nil
}

// FindConfig fetches the promoter config with the given id.
//
// Returns nil if no config exists.
func FindConfig(ctx context.Context, cli *client.Client, id string) (*PromoterConfig, string, error) {
	// TODO: replace by directly looking up the config file once LINSTOR is fixed.
	all, paths, err := ListConfigs(ctx, cli)
	if err != nil {
		return nil, "", err
	}

	for i := range all {
		if all[i].ID == id {
			return &all[i], paths[i], nil
		}
	}

	return nil, "", nil
}

// DeleteConfig removes the promoter of the given id from LINSTOR.
//
// In case the config did not exist, no error is returned.
func DeleteConfig(ctx context.Context, cli *client.Client, id string) error {
	path := ConfigPath(id)

	err := cli.Controller.DeleteExternalFile(ctx, path)
	if err != nil && err != client.NotFoundError {
		return fmt.Errorf("error removing config file: %w", err)
	}

	return nil
}

// ConfigPath is the file system path of the promoter config with the given id once it is deployed.
func ConfigPath(id string) string {
	return fmt.Sprintf(gatewayConfigPath, id)
}
