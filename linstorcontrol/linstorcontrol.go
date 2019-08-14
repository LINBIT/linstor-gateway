// Package linstorcontrol allows creating and deleting LINSTOR resources/volumes.
// It is a higher level abstraction to the low level golinstor REST package.
package linstorcontrol

import (
	"context"
	"errors"
	"net"
	"net/url"

	client "github.com/LINBIT/golinstor/client"
)

// Linstor is a struct containing the the configuration that is needed to create or delete a LINSTOR resource.
type Linstor struct {
	ResourceName      string `json:"resource_name,omitempty"`
	VlmSizeKiB        uint64 `json:"size_kib,omitempty"`
	ResourceGroupName string `json:"resource_group_name,omitempty"`
	Loglevel          string `json:"loglevel,omitempty"`
	ControllerIP      net.IP `json:"controller_ip,omitempty"`
}

// CreateResult is a struct than is used as the result of a successful create action.
// It already contains the data that is most likely used by a consumer of a CreateVolume() call.
type CreateResult struct {
	// Linux device path (e.g., /dev/drbd1001)
	DevicePath string
	// List of nodes where the actual data got places (i.e., after autoplace)
	StorageNodeList []string
}

// CreateVolume creates a  LINSTOR resource based on a given resource group name.
func (l *Linstor) CreateVolume() (CreateResult, error) {
	result := CreateResult{}

	clientCtx := context.Background()
	loglevel := l.Loglevel
	if loglevel == "" {
		loglevel = "info"
	}
	logCfg := &client.LogCfg{Level: loglevel}
	u, err := ipToURL(l.ControllerIP)
	if err != nil {
		return result, err
	}
	ctrlConn, err := client.NewClient(client.BaseURL(u), client.Log(logCfg))
	if err != nil {
		return result, err
	}

	spawn := client.ResourceGroupSpawn{
		ResourceDefinitionName: l.ResourceName,
		VolumeSizes:            []int64{int64(l.VlmSizeKiB)},
	}
	err = ctrlConn.ResourceGroups.Spawn(clientCtx, l.ResourceGroupName, spawn)
	if err != nil {
		return result, err
	}

	var storageNodes []string
	lopt := client.ListOpts{Resource: []string{l.ResourceName}}
	resources, err := ctrlConn.Resources.GetResourceView(clientCtx, &lopt)
	if err != nil {
		return result, err
	}

	for _, r := range resources {
		if r.Name != l.ResourceName {
			continue
		}
		if len(r.Volumes) == 0 {
			return result, errors.New("The volume list queried from the LINSTOR server contains no volumes")
		}
		if r.Volumes[0].ProviderKind != client.DISKLESS {
			storageNodes = append(storageNodes, r.NodeName)
			if result.DevicePath == "" {
				result.DevicePath = r.Volumes[0].DevicePath
			}
		}
	}
	if len(storageNodes) == 0 {
		return result, errors.New("Resource successfully deployed, but now found on on 0 nodes")
	}
	result.StorageNodeList = storageNodes

	return result, nil
}

// DeleteVolume deletes a LINSTOR resource definition (and therefore all resources) identified by name.
func (l *Linstor) DeleteVolume() error {
	clientCtx := context.Background()
	loglevel := l.Loglevel
	if loglevel == "" {
		loglevel = "info"
	}
	logCfg := &client.LogCfg{Level: loglevel}
	u, err := ipToURL(l.ControllerIP)
	if err != nil {
		return err
	}
	ctrlConn, err := client.NewClient(client.BaseURL(u), client.Log(logCfg))
	if err != nil {
		return err
	}

	return ctrlConn.ResourceDefinitions.Delete(clientCtx, l.ResourceName)
}

func ipToURL(ip net.IP) (*url.URL, error) {
	return url.Parse("http://" + ip.String() + ":3370")
}
