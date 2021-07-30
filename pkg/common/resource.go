package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/LINBIT/golinstor/client"
)

type VolumeConfig struct {
	Number  int    `json:"number"`
	SizeKiB uint64 `json:"size_kib"`
}

type ResourceStatus struct {
	State   ResourceState `json:"state"`
	Service ServiceState  `json:"service"`
	Primary string        `json:"primary"`
	Nodes   []string      `json:"nodes"`
	Volumes []VolumeState `json:"volumes"`
}

type Volume struct {
	Volume VolumeConfig `json:"volume"`
	Status VolumeState  `json:"status"`
}

type VolumeState struct {
	Number int           `json:"number"`
	State  ResourceState `json:"state"`
}

type ResourceState int

const (
	Unknown ResourceState = iota
	ResourceStateOK
	ResourceStateDegraded
	ResourceStateBad
)

func (l ResourceState) String() string {
	switch l {
	case ResourceStateOK:
		return "OK"
	case ResourceStateDegraded:
		return "Degraded"
	case ResourceStateBad:
		return "Bad"
	}
	return "Unknown"
}

func (l ResourceState) MarshalJSON() ([]byte, error) { return json.Marshal(l.String()) }

type ServiceState int

const (
	ServiceStateStopped ServiceState = iota
	ServiceStateStarted
)

func (s ServiceState) String() string {
	switch s {
	case ServiceStateStarted:
		return "Started"
	case ServiceStateStopped:
		return "Stopped"
	}

	return "Unknown"
}

func (s ServiceState) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }

func (s *ServiceState) UnmarshalJSON(text []byte) error {
	var raw string
	err := json.Unmarshal(text, &raw)
	if err != nil {
		return err
	}

	switch raw {
	case "Started":
		*s = ServiceStateStarted
	case "Stopped":
		*s = ServiceStateStopped
	default:
		return errors.New(fmt.Sprintf("unknown service state: %s", s))
	}

	return nil
}

func AnyResourcesInUse(resources []client.ResourceWithVolumes) bool {
	for _, resource := range resources {
		if resource.State.InUse {
			return true
		}
	}

	return false
}

func NoResourcesInUse(resources []client.ResourceWithVolumes) bool {
	return !AnyResourcesInUse(resources)
}

func WaitUntilResourceCondition(ctx context.Context, cli *client.Client, name string, condition func([]client.ResourceWithVolumes) bool) error {
	for {
		resources, err := cli.Resources.GetResourceView(ctx, &client.ListOpts{Resource: []string{name}})
		if err != nil {
			return err
		}

		if condition(resources) {
			return nil
		}

		time.Sleep(3 * time.Second)
	}
}
