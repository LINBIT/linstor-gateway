package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/LINBIT/golinstor/client"
)

type UidGid struct {
	Uid int
	Gid int
}

func (u *UidGid) String() string {
	return fmt.Sprintf("%d:%d", u.Uid, u.Gid)
}

type VolumeConfig struct {
	Number              int    `json:"number"`
	SizeKiB             uint64 `json:"size_kib"`
	FileSystem          string `json:"file_system,omitempty"`
	FileSystemRootOwner UidGid `json:"file_system_root_owner,omitempty"`
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

func (l *ResourceState) UnmarshalJSON(text []byte) error {
	var raw string
	err := json.Unmarshal(text, &raw)
	if err != nil {
		return err
	}

	switch raw {
	case "OK":
		*l = ResourceStateOK
	case "Degraded":
		*l = ResourceStateDegraded
	case "Bad":
		*l = ResourceStateBad
	case "Unknown":
		*l = Unknown
	default:
		return errors.New(fmt.Sprintf("unknown resource state: %s", string(text)))
	}

	return nil
}

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
		if resource.State != nil && resource.State.InUse != nil && *resource.State.InUse {
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

func getInUseNode(ctx context.Context, cli *client.Client, name string) (string, error) {
	resources, err := cli.Resources.GetResourceView(ctx, &client.ListOpts{Resource: []string{name}})
	if err != nil {
		return "", err
	}
	for _, resource := range resources {
		if resource.State != nil && resource.State.InUse != nil && *resource.State.InUse {
			return resource.NodeName, nil
		}
	}

	return "", nil
}

// AssertResourceInUseStable records the node that a resource is running on, and then monitors the resource to check
// that it remains on the same node for a few seconds. This is useful for sanity-checking that a resource has started
// up and is healthy.
func AssertResourceInUseStable(ctx context.Context, cli *client.Client, name string) error {
	initialNode, err := getInUseNode(ctx, cli, name)
	if err != nil {
		return fmt.Errorf("failed to get InUse node for resource %s: %w", name, err)
	}
	if initialNode == "" {
		return fmt.Errorf("resource %s is not in use on any node", name)
	}

	count := 0
	for {
		node, err := getInUseNode(ctx, cli, name)
		if err != nil {
			return err
		}
		if node != initialNode {
			return fmt.Errorf("resource startup failed on node %s", initialNode)
		}

		time.Sleep(1 * time.Second)
		count++
		if count > 5 {
			return nil
		}
	}
}
