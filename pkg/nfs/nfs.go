package nfs

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"path/filepath"
	"time"

	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
)

const (
	IDFormat       = "nfs-%s"
	FilenameFormat = "linstor-gateway-nfs-%s.toml"
)

type NFS struct {
	cli *linstorcontrol.Linstor
}

func New(controllers []string) (*NFS, error) {
	cli, err := linstorcontrol.Default(controllers)
	if err != nil {
		return nil, fmt.Errorf("failed to create linstor client: %w", err)
	}
	return &NFS{cli}, nil
}

func (n *NFS) Get(ctx context.Context, name string) (*ResourceConfig, error) {
	cfg, path, err := reactor.FindConfig(ctx, n.cli.Client, fmt.Sprintf(IDFormat, name))
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing config: %w", err)
	}

	if cfg == nil {
		return nil, nil
	}

	resourceDefinition, resourceGroup, volumeDefinitions, resources, err := cfg.DeployedResources(ctx, n.cli.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing deployment: %w", err)
	}

	deployedCfg, err := FromPromoter(cfg, resourceDefinition, volumeDefinitions)
	if err != nil {
		return nil, fmt.Errorf("unknown existing reactor config: %w", err)
	}

	deployedCfg.Status = linstorcontrol.StatusFromResources(path, resourceDefinition, resourceGroup, resources)

	return deployedCfg, nil
}

// getExistingDeployment returns the ResourceConfig for an existing reactor.PromoterConfig.
// If the corresponding LINSTOR resource does not exist, it returns nil (but also a nil error).
// If the LINSTOR resource does exist but is invalid, it returns an error.
func (n *NFS) getExistingDeployment(ctx context.Context, rsc *ResourceConfig, cfg *reactor.PromoterConfig, path string) (*ResourceConfig, error) {
	resourceDefinition, resourceGroup, volumeDefinitions, resources, err := cfg.DeployedResources(ctx, n.cli.Client)
	if err != nil {
		log.Warnf("Found an existing promoter config but no corresponding LINSTOR resource. Maybe left over from a previous deployment?")
		log.Warnf("Ignoring and overwriting the existing configuration at %s.", path)
		return nil, nil
	}

	deployedCfg, err := FromPromoter(cfg, resourceDefinition, volumeDefinitions)
	if err != nil {
		return nil, fmt.Errorf("unknown existing reactor config: %w", err)
	}

	if !rsc.Matches(deployedCfg) {
		log.Debugf("existing resource found that does not match config")
		log.Debugf("diff: %s", cmp.Diff(deployedCfg, rsc))
		return nil, errors.New("resource already exists with incompatible config")
	}

	deployedCfg.Status = linstorcontrol.StatusFromResources(path, resourceDefinition, resourceGroup, resources)
	return deployedCfg, nil
}

func (n *NFS) isNFSConfig(config reactor.PromoterConfig) bool {
	for _, r := range config.Resources {
		for _, s := range r.Start {
			if agent, ok := s.(*reactor.ResourceAgent); ok {
				if agent.Type == "ocf:heartbeat:nfsserver" {
					return true
				}
			}
		}
	}
	return false
}

// checkExistingConfigs goes through the list of currently deployed configs, and checks:
// - if an NFS config already exists; if so, abort with an error
// - if an existing config has the same name as the one we are trying to create; if so, just return that
// - if the IP address of the new config collides with an existing config; if so, abort with an error
func (n *NFS) checkExistingConfigs(ctx context.Context, newRsc *ResourceConfig) (*reactor.PromoterConfig, string, error) {
	configs, paths, err := reactor.ListConfigs(ctx, n.cli.Client)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check for existing NFS configs: %w", err)
	}

	var existingConfig *reactor.PromoterConfig
	var existingPath string
	for i := range configs {
		c := configs[i]
		p := paths[i]

		// filter out any existing config with the same name as the one we are trying to create.
		if name, _ := c.FirstResource(); name == newRsc.Name {
			existingConfig = &c
			existingPath = p
			continue
		}
		if n.isNFSConfig(c) {
			return nil, "", fmt.Errorf("an NFS config already exists in %s. Only one NFS resource is allowed", p)
		}

		if err := common.CheckIPCollision(c, newRsc.ServiceIP.IP()); err != nil {
			return nil, "", fmt.Errorf("invalid configuration: %w", err)
		}
	}
	return existingConfig, existingPath, nil
}

// Create creates an NFS export according to the resource configuration
// described in rsc. It automatically prepends a "cluster private volume" to the
// list of volumes, so volume numbers must start at 1.
func (n *NFS) Create(ctx context.Context, rsc *ResourceConfig) (*ResourceConfig, error) {
	rsc.FillDefaults()

	// prepend cluster private volume; it should always be the first volume and have number 0
	rsc.Volumes = append([]VolumeConfig{{VolumeConfig: common.ClusterPrivateVolume()}}, rsc.Volumes...)

	err := rsc.Valid()
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	existingConfig, existingPath, err := n.checkExistingConfigs(ctx, rsc)
	if err != nil {
		return nil, err
	}
	if existingConfig != nil {
		deployedCfg, err := n.getExistingDeployment(ctx, rsc, existingConfig, existingPath)
		if err != nil {
			return nil, err
		}
		if deployedCfg != nil {
			return deployedCfg, nil
		}
	}

	volumes := make([]common.VolumeConfig, len(rsc.Volumes))
	for i := range rsc.Volumes {
		volumes[i] = rsc.Volumes[i].VolumeConfig
	}

	resourceDefinition, resourceGroup, deployment, err := n.cli.EnsureResource(ctx, linstorcontrol.Resource{
		Name:          rsc.Name,
		ResourceGroup: rsc.ResourceGroup,
		Volumes:       volumes,
		GrossSize:     rsc.GrossSize,
	}, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create linstor resource: %w", err)
	}

	defer func() {
		// if we fail beyond this point, roll back by deleting the created resource definition
		if err != nil {
			log.Debugf("Rollback: deleting just created resource definition %s", rsc.Name)
			err := n.cli.ResourceDefinitions.Delete(ctx, rsc.Name)
			if err != nil {
				log.Warnf("Failed to roll back created resource definition: %v", err)
			}
		}
	}()

	existingConfig, err = rsc.ToPromoter(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resource to promoter configuration: %w", err)
	}

	err = reactor.EnsureConfig(ctx, n.cli.Client, existingConfig, rsc.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to register reactor config file: %w", err)
	}

	defer func() {
		// if we fail beyond this point, delete the just created reactor config
		if err != nil {
			log.Debugf("Rollback: deleting just created reactor config %s", rsc.ID())
			if err := reactor.DeleteConfig(ctx, n.cli.Client, rsc.ID()); err != nil {
				log.Warnf("Failed to roll back created reactor config: %v", err)
			}

			if err := common.WaitUntilResourceCondition(ctx, n.cli.Client, rsc.Name, common.NoResourcesInUse); err != nil {
				log.Warnf("Failed to wait for resource to become unused: %v", err)
			}
		}
	}()

	_, err = n.Start(ctx, rsc.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to start resources: %w", err)
	}

	rsc.Status = linstorcontrol.StatusFromResources(existingPath, resourceDefinition, resourceGroup, deployment)

	return rsc, nil
}

func (n *NFS) Start(ctx context.Context, name string) (*ResourceConfig, error) {
	cfg, path, err := reactor.FindConfig(ctx, n.cli.Client, fmt.Sprintf(IDFormat, name))
	if err != nil {
		return nil, fmt.Errorf("failed to find the resource configuration: %w", err)
	}

	if cfg == nil {
		return nil, nil
	}

	err = reactor.AttachConfig(ctx, n.cli.Client, cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to attach reactor configuration: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = common.WaitUntilResourceCondition(waitCtx, n.cli.Client, name, common.AnyResourcesInUse)
	if err != nil {
		return nil, fmt.Errorf("error waiting for resource to become used: %w", err)
	}

	err = common.AssertResourceInUseStable(waitCtx, n.cli.Client, name)
	if err != nil {
		return nil, fmt.Errorf("error waiting for resource to become stable: %w", err)
	}

	return n.Get(ctx, name)
}

func (n *NFS) Stop(ctx context.Context, name string) (*ResourceConfig, error) {
	cfg, path, err := reactor.FindConfig(ctx, n.cli.Client, fmt.Sprintf(IDFormat, name))
	if err != nil {
		return nil, fmt.Errorf("failed to find the resource configuration: %w", err)
	}

	if cfg == nil {
		return nil, nil
	}

	err = reactor.DetachConfig(ctx, n.cli.Client, cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to detach reactor configuration: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = common.WaitUntilResourceCondition(waitCtx, n.cli.Client, name, common.NoResourcesInUse)
	if err != nil {
		return nil, fmt.Errorf("error waiting for resource to become unused: %w", err)
	}

	return n.Get(ctx, name)
}

func (n *NFS) List(ctx context.Context) ([]*ResourceConfig, error) {
	cfgs, paths, err := reactor.ListConfigs(ctx, n.cli.Client)
	if err != nil {
		return nil, err
	}

	result := make([]*ResourceConfig, 0, len(cfgs))
	for i := range cfgs {
		cfg := &cfgs[i]
		path := paths[i]
		filename := filepath.Base(path)

		var rsc string
		num, _ := fmt.Sscanf(filename, FilenameFormat, &rsc)
		if num == 0 {
			log.WithField("filename", filename).Trace("not an nfs resource config, skipping")
			continue
		}

		resourceDefinition, resourceGroup, volumeDefinitions, resources, err := cfg.DeployedResources(ctx, n.cli.Client)
		if err != nil {
			log.WithError(err).Warn("failed to fetch deployed resources")
		}

		parsed, err := FromPromoter(cfg, resourceDefinition, volumeDefinitions)
		if err != nil {
			log.WithError(err).Warn("skipping error while parsing promoter config")
			continue
		}

		parsed.Status = linstorcontrol.StatusFromResources(path, resourceDefinition, resourceGroup, resources)

		result = append(result, parsed)
	}

	return result, nil
}

func (n *NFS) Delete(ctx context.Context, name string) error {
	err := reactor.DeleteConfig(ctx, n.cli.Client, fmt.Sprintf(IDFormat, name))
	if err != nil {
		return fmt.Errorf("failed to delete reactor config: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = common.WaitUntilResourceCondition(waitCtx, n.cli.Client, name, common.NoResourcesInUse)
	if err != nil {
		return fmt.Errorf("error waiting for resource to become unused: %w", err)
	}

	err = n.cli.ResourceDefinitions.Delete(ctx, name)
	if err != nil && err != client.NotFoundError {
		return fmt.Errorf("failed to delete resources: %w", err)
	}

	return nil
}

func (n *NFS) DeleteVolume(ctx context.Context, name string, lun int) (*ResourceConfig, error) {
	cfg, path, err := reactor.FindConfig(ctx, n.cli.Client, fmt.Sprintf(IDFormat, name))
	if err != nil {
		return nil, fmt.Errorf("failed to delete reactor config: %w", err)
	}

	if cfg == nil {
		return nil, nil
	}

	resourceDefinition, resourceGroup, volumeDefinition, resources, err := cfg.DeployedResources(ctx, n.cli.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployed resources: %w", err)
	}

	rscCfg, err := FromPromoter(cfg, resourceDefinition, volumeDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to convert volume definition to resource: %w", err)
	}

	status := linstorcontrol.StatusFromResources(path, resourceDefinition, resourceGroup, resources)
	if status.Service == common.ServiceStateStarted {
		return nil, errors.New("cannot delete volume while service is running")
	}

	for i := range rscCfg.Volumes {
		if rscCfg.Volumes[i].Number == lun {
			err = n.cli.ResourceDefinitions.DeleteVolumeDefinition(ctx, name, lun)
			if err != nil && err != client.NotFoundError {
				return nil, fmt.Errorf("failed to delete volume definition")
			}

			rscCfg.Volumes = append(rscCfg.Volumes[:i], rscCfg.Volumes[i+1:]...)
			// Manually delete the resources from the current resource config
			for j := range resources {
				resources[j].Volumes = append(resources[j].Volumes[:i], resources[j].Volumes[i+1:]...)
			}

			cfg, err = rscCfg.ToPromoter(resources)
			if err != nil {
				return nil, fmt.Errorf("failed to convert resource to promoter configuration: %w", err)
			}

			err = reactor.EnsureConfig(ctx, n.cli.Client, cfg, rscCfg.ID())
			if err != nil {
				return nil, fmt.Errorf("failed to update config")
			}

			break
		}
	}

	return n.Get(ctx, name)
}
