package client

import (
	"context"
	"time"

	"github.com/LINBIT/linstor-gateway/pkg/nfs"
)

type NFSService struct {
	client *Client
}

func (s *NFSService) GetAll(ctx context.Context) ([]*nfs.ResourceConfig, error) {
	var configs []*nfs.ResourceConfig
	_, err := s.client.doGET(ctx, "/api/v2/nfs", &configs)
	return configs, err
}

func (s *NFSService) Create(ctx context.Context, config *nfs.ResourceConfig) (*nfs.ResourceConfig, error) {
	var ret *nfs.ResourceConfig
	_, err := s.client.doPOST(ctx, "/api/v2/nfs", config, &ret)
	return ret, err
}

func (s *NFSService) Get(ctx context.Context, name string) (*nfs.ResourceConfig, error) {
	var config *nfs.ResourceConfig
	_, err := s.client.doGET(ctx, "/api/v2/nfs/"+name, &config)
	return config, err
}

func (s *NFSService) Delete(ctx context.Context, name string, resourceTimeout time.Duration) error {
	url := "/api/v2/nfs/" + name
	if resourceTimeout > 0 {
		url += "?resource_timeout=" + resourceTimeout.String()
	}
	_, err := s.client.doDELETE(ctx, url, nil)
	return err
}

func (s *NFSService) Start(ctx context.Context, name string, resourceTimeout time.Duration) (*nfs.ResourceConfig, error) {
	var ret *nfs.ResourceConfig
	url := "/api/v2/nfs/" + name + "/start"
	if resourceTimeout > 0 {
		url += "?resource_timeout=" + resourceTimeout.String()
	}
	_, err := s.client.doPOST(ctx, url, nil, &ret)
	return ret, err
}

func (s *NFSService) Stop(ctx context.Context, name string, resourceTimeout time.Duration) (*nfs.ResourceConfig, error) {
	var ret *nfs.ResourceConfig
	url := "/api/v2/nfs/" + name + "/stop"
	if resourceTimeout > 0 {
		url += "?resource_timeout=" + resourceTimeout.String()
	}
	_, err := s.client.doPOST(ctx, url, nil, &ret)
	return ret, err
}
