// Package iscsi combines LINSTOR operations and the CRM operations to create highly available iSCSI targets.
package iscsi

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
)

const IDFormat = "iscsi-%s"

type ISCSI struct {
	cli *linstorcontrol.Linstor
}

func New(controllers []string) (*ISCSI, error) {
	cli, err := linstorcontrol.Default(controllers)
	if err != nil {
		return nil, fmt.Errorf("failed to create linstor client: %w", err)
	}
	return &ISCSI{cli}, nil
}

func (i *ISCSI) Get(ctx context.Context, iqn Iqn) (*ResourceConfig, error) {
	cfg, path, err := reactor.FindConfig(ctx, i.cli.Client, fmt.Sprintf(IDFormat, iqn.WWN()))
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing config: %w", err)
	}

	if cfg == nil {
		return nil, nil
	}

	resourceDefinition, volumeDefinitions, resources, err := cfg.DeployedResources(ctx, i.cli.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing deployment: %w", err)
	}

	deployedCfg, err := FromPromoter(cfg, resourceDefinition, volumeDefinitions)
	if err != nil {
		return nil, fmt.Errorf("unknown existing reactor config: %w", err)
	}

	deployedCfg.Status = linstorcontrol.StatusFromResources(path, resourceDefinition, resources)

	return deployedCfg, nil
}

func (i *ISCSI) Create(ctx context.Context, rsc *ResourceConfig) (*ResourceConfig, error) {
	rsc.FillDefaults()

	err := rsc.Valid()
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	cfg, path, err := reactor.FindConfig(ctx, i.cli.Client, fmt.Sprintf(IDFormat, rsc.IQN.WWN()))
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing config: %w", err)
	}

	if cfg != nil {
		resourceDefinition, volumeDefinitions, resources, err := cfg.DeployedResources(ctx, i.cli.Client)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch existing deployment: %w", err)
		}

		deployedCfg, err := FromPromoter(cfg, resourceDefinition, volumeDefinitions)
		if err != nil {
			return nil, fmt.Errorf("unknown existing reactor config: %w", err)
		}

		if !rsc.Matches(deployedCfg) {
			return nil, errors.New("resource already exists with incompatible config")
		}

		deployedCfg.Status = linstorcontrol.StatusFromResources(path, resourceDefinition, resources)

		return deployedCfg, nil
	}

	resourceDefinition, deployment, err := i.cli.EnsureResource(ctx, linstorcontrol.Resource{
		Name:          rsc.IQN.WWN(),
		ResourceGroup: rsc.ResourceGroup,
		Volumes:       rsc.Volumes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create linstor resource: %w", err)
	}

	cfg, err = rsc.ToPromoter(deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resource to promoter configuration: %w", err)
	}

	err = reactor.EnsureConfig(ctx, i.cli.Client, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to register reactor config file: %w", err)
	}

	_, err = i.Start(ctx, rsc.IQN)
	if err != nil {
		return nil, fmt.Errorf("failed to start resources: %w", err)
	}

	rsc.Status = linstorcontrol.StatusFromResources(path, resourceDefinition, deployment)

	return rsc, nil
}

func (i *ISCSI) Start(ctx context.Context, iqn Iqn) (*ResourceConfig, error) {
	cfg, _, err := reactor.FindConfig(ctx, i.cli.Client, fmt.Sprintf(IDFormat, iqn.WWN()))
	if err != nil {
		return nil, fmt.Errorf("failed to find the resource configuration: %w", err)
	}

	if cfg == nil {
		return nil, nil
	}

	err = reactor.AttachConfig(ctx, i.cli.Client, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to detach reactor configuration: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = common.WaitUntilResourceCondition(waitCtx, i.cli.Client, iqn.WWN(), common.AnyResourcesInUse)
	if err != nil {
		return nil, fmt.Errorf("error waiting for resource to become used: %w", err)
	}

	return i.Get(ctx, iqn)
}

func (i *ISCSI) Stop(ctx context.Context, iqn Iqn) (*ResourceConfig, error) {
	cfg, _, err := reactor.FindConfig(ctx, i.cli.Client, fmt.Sprintf(IDFormat, iqn.WWN()))
	if err != nil {
		return nil, fmt.Errorf("failed to find the resource configuration: %w", err)
	}

	if cfg == nil {
		return nil, nil
	}

	err = reactor.DetachConfig(ctx, i.cli.Client, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to detach reactor configuration: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = common.WaitUntilResourceCondition(waitCtx, i.cli.Client, iqn.WWN(), common.NoResourcesInUse)
	if err != nil {
		return nil, fmt.Errorf("error waiting for resource to become unused: %w", err)
	}

	return i.Get(ctx, iqn)
}

func (i *ISCSI) List(ctx context.Context) ([]*ResourceConfig, error) {
	cfgs, paths, err := reactor.ListConfigs(ctx, i.cli.Client)
	if err != nil {
		return nil, err
	}

	result := make([]*ResourceConfig, 0, len(cfgs))
	for j := range cfgs {
		cfg := &cfgs[j]
		path := paths[j]

		var rsc string
		n, _ := fmt.Sscanf(cfg.ID, IDFormat, &rsc)
		if n == 0 {
			log.WithField("id", cfg.ID).Trace("not a nvme resource config, skipping")
			continue
		}

		resourceDefinition, volumeDefinitions, resources, err := cfg.DeployedResources(ctx, i.cli.Client)
		if err != nil {
			log.WithError(err).Warn("failed to fetch deployed resources")
		}

		parsed, err := FromPromoter(cfg, resourceDefinition, volumeDefinitions)
		if err != nil {
			log.WithError(err).Warn("skipping error while parsing promoter config")
			continue
		}

		parsed.Status = linstorcontrol.StatusFromResources(path, resourceDefinition, resources)

		result = append(result, parsed)
	}

	return result, nil
}

func (i *ISCSI) Delete(ctx context.Context, iqn Iqn) error {
	err := reactor.DeleteConfig(ctx, i.cli.Client, fmt.Sprintf(IDFormat, iqn.WWN()))
	if err != nil {
		return fmt.Errorf("failed to delete reactor config: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = common.WaitUntilResourceCondition(waitCtx, i.cli.Client, iqn.WWN(), common.NoResourcesInUse)
	if err != nil {
		return fmt.Errorf("error waiting for resource to become unused: %w", err)
	}

	err = i.cli.ResourceDefinitions.Delete(ctx, iqn.WWN())
	if err != nil && err != client.NotFoundError {
		return fmt.Errorf("failed to delete resources: %w", err)
	}

	return nil
}

func (i *ISCSI) AddVolume(ctx context.Context, iqn Iqn, volCfg *common.VolumeConfig) (*ResourceConfig, error) {
	cfg, path, err := reactor.FindConfig(ctx, i.cli.Client, fmt.Sprintf(IDFormat, iqn.WWN()))
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing config: %w", err)
	}

	if cfg == nil {
		return nil, nil
	}

	resourceDefinition, volumeDefinitions, resources, err := cfg.DeployedResources(ctx, i.cli.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing deployment: %w", err)
	}

	deployedCfg, err := FromPromoter(cfg, resourceDefinition, volumeDefinitions)
	if err != nil {
		return nil, fmt.Errorf("unknown existing reactor config: %w", err)
	}

	exists := false
	for i := range deployedCfg.Volumes {
		if deployedCfg.Volumes[i].Number == volCfg.Number {
			if deployedCfg.Volumes[i].SizeKiB != volCfg.SizeKiB {
				return nil, errors.New(fmt.Sprintf("existing volume has differing size %d != %d", deployedCfg.Volumes[i].SizeKiB, volCfg.SizeKiB))
			}

			exists = true
			break
		}
	}

	if !exists {
		status := linstorcontrol.StatusFromResources(path, resourceDefinition, resources)
		if status.Service == common.ServiceStateStarted {
			return nil, errors.New("cannot add volume while service is running")
		}

		deployedCfg.Volumes = append(deployedCfg.Volumes, *volCfg)

		sort.Slice(deployedCfg.Volumes, func(i, j int) bool {
			return deployedCfg.Volumes[i].Number < deployedCfg.Volumes[j].Number
		})

		err = deployedCfg.Valid()
		if err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		resourceDefinition, resources, err = i.cli.EnsureResource(ctx, linstorcontrol.Resource{
			Name:          deployedCfg.IQN.WWN(),
			ResourceGroup: deployedCfg.ResourceGroup,
			Volumes:       deployedCfg.Volumes,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to reconcile linstor resource: %w", err)
		}
	}

	cfg, err = deployedCfg.ToPromoter(resources)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resource to promoter configuration: %w", err)
	}

	err = reactor.EnsureConfig(ctx, i.cli.Client, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to update config: %w", err)
	}

	deployedCfg.Status = linstorcontrol.StatusFromResources(path, resourceDefinition, resources)

	return deployedCfg, nil
}

func (i *ISCSI) DeleteVolume(ctx context.Context, iqn Iqn, lun int) (*ResourceConfig, error) {
	cfg, path, err := reactor.FindConfig(ctx, i.cli.Client, fmt.Sprintf(IDFormat, iqn.WWN()))
	if err != nil {
		return nil, fmt.Errorf("failed to delete reactor config: %w", err)
	}

	if cfg == nil {
		return nil, nil
	}

	resourceDefinition, volumeDefinition, resources, err := cfg.DeployedResources(ctx, i.cli.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployed resources: %w", err)
	}

	rscCfg, err := FromPromoter(cfg, resourceDefinition, volumeDefinition)
	if err != nil {
		return nil, fmt.Errorf("failed to convert volume definition to resource: %w", err)
	}

	status := linstorcontrol.StatusFromResources(path, resourceDefinition, resources)
	if status.Service == common.ServiceStateStarted {
		return nil, errors.New("cannot delete volume while service is running")
	}

	for j := range rscCfg.Volumes {
		if rscCfg.Volumes[j].Number == lun {
			err = i.cli.ResourceDefinitions.DeleteVolumeDefinition(ctx, iqn.WWN(), lun)
			if err != nil && err != client.NotFoundError {
				return nil, fmt.Errorf("failed to delete volume definition")
			}

			rscCfg.Volumes = append(rscCfg.Volumes[:j], rscCfg.Volumes[j+1:]...)
			// Manually delete the resources from the current resource config
			for k := range resources {
				resources[k].Volumes = append(resources[k].Volumes[:j], resources[k].Volumes[j+1:]...)
			}

			cfg, err = rscCfg.ToPromoter(resources)
			if err != nil {
				return nil, fmt.Errorf("failed to convert resource to promoter configuration: %w", err)
			}

			err = reactor.EnsureConfig(ctx, i.cli.Client, cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to update config")
			}

			break
		}
	}

	return i.Get(ctx, iqn)
}
