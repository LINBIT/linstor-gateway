package client

import (
	"context"
	"fmt"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

type ISCSIService struct {
	client *Client
}

func (s *ISCSIService) GetAll(ctx context.Context) ([]*iscsi.ResourceConfig, error) {
	var configs []*iscsi.ResourceConfig
	_, err := s.client.doGET(ctx, "/api/v2/iscsi", &configs)
	return configs, err
}

func (s *ISCSIService) Create(ctx context.Context, config *iscsi.ResourceConfig) (*iscsi.ResourceConfig, error) {
	var ret *iscsi.ResourceConfig
	_, err := s.client.doPOST(ctx, "/api/v2/iscsi", config, &ret)
	return ret, err
}

func (s *ISCSIService) Get(ctx context.Context, iqn iscsi.Iqn) (*iscsi.ResourceConfig, error) {
	var config *iscsi.ResourceConfig
	_, err := s.client.doGET(ctx, "/api/v2/iscsi/"+iqn.String(), &config)
	return config, err
}

func (s *ISCSIService) Delete(ctx context.Context, iqn iscsi.Iqn) error {
	_, err := s.client.doDELETE(ctx, "/api/v2/iscsi/"+iqn.String(), nil)
	return err
}

func (s *ISCSIService) Start(ctx context.Context, iqn iscsi.Iqn) (*iscsi.ResourceConfig, error) {
	var ret *iscsi.ResourceConfig
	_, err := s.client.doPOST(ctx, "/api/v2/iscsi/"+iqn.String()+"/start", nil, &ret)
	return ret, err
}

func (s *ISCSIService) Stop(ctx context.Context, iqn iscsi.Iqn) (*iscsi.ResourceConfig, error) {
	var ret *iscsi.ResourceConfig
	_, err := s.client.doPOST(ctx, "/api/v2/iscsi/"+iqn.String()+"/stop", nil, &ret)
	return ret, err
}

func (s *ISCSIService) GetLogicalUnit(ctx context.Context, iqn iscsi.Iqn, lun int) (*common.VolumeConfig, error) {
	var config *common.VolumeConfig
	_, err := s.client.doGET(ctx, fmt.Sprintf("/api/v2/iscsi/%s/%d", iqn.String(), lun), &config)
	return config, err
}

func (s *ISCSIService) AddLogicalUnit(ctx context.Context, iqn iscsi.Iqn, volume *common.VolumeConfig) (*common.VolumeConfig, error) {
	var ret *common.VolumeConfig
	_, err := s.client.doPUT(ctx, fmt.Sprintf("/api/v2/iscsi/%s/%d", iqn.String(), volume.Number), volume, &ret)
	return ret, err
}

func (s *ISCSIService) DeleteLogicalUnit(ctx context.Context, iqn iscsi.Iqn, lun int) error {
	_, err := s.client.doDELETE(ctx, fmt.Sprintf("/api/v2/iscsi/%s/%d", iqn.String(), lun), nil)
	return err
}
