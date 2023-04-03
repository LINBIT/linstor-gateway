package client

import (
	"context"
)

type StatusService struct {
	client *Client
}

type Status struct {
	Status string `json:"status"`
}

func (s *StatusService) Get(ctx context.Context) (*Status, error) {
	var status *Status
	_, err := s.client.doGET(ctx, "/api/v2/status", &status)
	return status, err
}
