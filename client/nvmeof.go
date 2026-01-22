package client

import (
	"context"
	"fmt"
	"time"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

type NvmeOfService struct {
	client *Client
}

func (s *NvmeOfService) GetAll(ctx context.Context) ([]nvmeof.ResourceConfig, error) {
	var configs []nvmeof.ResourceConfig
	_, err := s.client.doGET(ctx, "/api/v2/nvme-of", &configs)
	return configs, err
}

func (s *NvmeOfService) Create(ctx context.Context, config *nvmeof.ResourceConfig) (*nvmeof.ResourceConfig, error) {
	var ret *nvmeof.ResourceConfig
	_, err := s.client.doPOST(ctx, "/api/v2/nvme-of", config, &ret)
	return ret, err
}

func (s *NvmeOfService) Get(ctx context.Context, nqn nvmeof.Nqn) (*nvmeof.ResourceConfig, error) {
	var config *nvmeof.ResourceConfig
	_, err := s.client.doGET(ctx, "/api/v2/nvme-of/"+nqn.String(), &config)
	return config, err
}

func (s *NvmeOfService) Delete(ctx context.Context, nqn nvmeof.Nqn, resourceTimeout time.Duration) error {
	url := "/api/v2/nvme-of/" + nqn.String()
	if resourceTimeout > 0 {
		url += "?resource_timeout=" + resourceTimeout.String()
	}
	_, err := s.client.doDELETE(ctx, url, nil)
	return err
}

func (s *NvmeOfService) Start(ctx context.Context, nqn nvmeof.Nqn, resourceTimeout time.Duration) (*nvmeof.ResourceConfig, error) {
	var ret *nvmeof.ResourceConfig
	url := "/api/v2/nvme-of/" + nqn.String() + "/start"
	if resourceTimeout > 0 {
		url += "?resource_timeout=" + resourceTimeout.String()
	}
	_, err := s.client.doPOST(ctx, url, nil, &ret)
	return ret, err
}

func (s *NvmeOfService) Stop(ctx context.Context, nqn nvmeof.Nqn, resourceTimeout time.Duration) (*nvmeof.ResourceConfig, error) {
	var ret *nvmeof.ResourceConfig
	url := "/api/v2/nvme-of/" + nqn.String() + "/stop"
	if resourceTimeout > 0 {
		url += "?resource_timeout=" + resourceTimeout.String()
	}
	_, err := s.client.doPOST(ctx, url, nil, &ret)
	return ret, err
}

func (s *NvmeOfService) GetVolume(ctx context.Context, nqn nvmeof.Nqn, lun int) (*common.VolumeConfig, error) {
	var config *common.VolumeConfig
	_, err := s.client.doGET(ctx, fmt.Sprintf("/api/v2/nvme-of/%s/%d", nqn.String(), lun), &config)
	return config, err
}

func (s *NvmeOfService) AddVolume(ctx context.Context, nqn nvmeof.Nqn, volume *common.VolumeConfig) (*common.VolumeConfig, error) {
	var ret *common.VolumeConfig
	_, err := s.client.doPUT(ctx, fmt.Sprintf("/api/v2/nvme-of/%s/%d", nqn.String(), volume.Number), volume, &ret)
	return ret, err
}

func (s *NvmeOfService) DeleteVolume(ctx context.Context, nqn nvmeof.Nqn, volume int) error {
	_, err := s.client.doDELETE(ctx, fmt.Sprintf("/api/v2/nvme-of/%s/%d", nqn.String(), volume), nil)
	return err
}
