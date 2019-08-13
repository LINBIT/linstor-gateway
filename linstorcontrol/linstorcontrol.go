// LINSTOR API - creating/deleting LINSTOR resources/volumes
package linstorcontrol

// linstorcontrol module
//
// The functions in this module are called by the high-level API in package application
// (module application.go) to perform operations in the LINSTOR subsystem, such
// as creating and deleting resources/volumes. The golinstor driver is used
// for communication with the LINSTOR Controller.

import (
	"context"
	"errors"
	"net"
	"net/url"

	client "github.com/LINBIT/golinstor/client"
)

type Linstor struct {
	ResourceName      string `json:"resource_name,omitempty"`
	VlmSizeKiB        uint64 `json:"size_kib,omitempty"`
	ResourceGroupName string `json:"resource_group_name,omitempty"`
	Loglevel          string `json:"loglevel,omitempty"`
	ControllerIP      net.IP `json:"controller_ip,omitempty"`
}

func ipToURL(ip net.IP) (*url.URL, error) {
	return url.Parse("http://" + ip.String() + ":3370")
}

type CreateResult struct {
	DevicePath      string
	StorageNodeList []string
}

// Creates a LINSTOR resource definition, volume definition and associated resources on the selected nodes
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

// Deletes the LINSTOR resource definition
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
