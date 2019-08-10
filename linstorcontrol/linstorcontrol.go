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
	"fmt"
	"net"
	"net/url"
	"strconv"

	client "github.com/LINBIT/golinstor/client"
)

type Linstor struct {
	ResourceName    string
	VlmSizeKiB      uint64
	StorageNodeList []string
	ClientNodeList  []string
	AutoPlaceCount  uint64
	StorageStorPool string
	ClientStorPool  string
	Loglevel        string
	ControllerIP    net.IP
}

func ipToURL(ip net.IP) (*url.URL, error) {
	return url.Parse("http://" + ip.String() + ":3370")
}

// Creates a LINSTOR resource definition, volume definition and associated resources on the selected nodes
func (l *Linstor) CreateVolume() (string, error) {
	if len(l.StorageNodeList) < 1 {
		return "", errors.New("Invalid CreateVolume() call: Parameter storageNodeList is an empty list")
	}

	clientCtx := context.Background()
	logCfg := &client.LogCfg{Level: l.Loglevel}
	u, err := ipToURL(l.ControllerIP)
	if err != nil {
		return "", err
	}
	ctrlConn, err := client.NewClient(client.BaseURL(u), client.Log(logCfg))
	if err != nil {
		return "", err
	}

	// Create a resource definition
	rscDfnData := client.ResourceDefinitionCreate{
		ResourceDefinition: client.ResourceDefinition{Name: l.ResourceName},
	}
	err = ctrlConn.ResourceDefinitions.Create(clientCtx, rscDfnData)
	if err != nil {
		return "", err
	}

	// Create a volume definition
	vlmDfnData := client.VolumeDefinitionCreate{
		VolumeDefinition: client.VolumeDefinition{VolumeNumber: int32(0), SizeKib: l.VlmSizeKiB},
	}
	err = ctrlConn.ResourceDefinitions.CreateVolumeDefinition(
		clientCtx,
		rscDfnData.ResourceDefinition.Name,
		vlmDfnData,
	)
	if err != nil {
		return "", err
	}

	// Create resources on all selected nodes
	crtRscFailed := 0
	for _, tgtNodeName := range l.StorageNodeList {
		rscData := client.ResourceCreate{
			Resource: client.Resource{
				Name: rscDfnData.ResourceDefinition.Name, NodeName: tgtNodeName,
			},
		}
		err = ctrlConn.Resources.Create(clientCtx, rscData)
		if err != nil {
			crtRscFailed++
			fmt.Printf("%s\n", err.Error())
		}
	}
	rscLabel := "resource"
	if crtRscFailed > 1 {
		rscLabel = "resources"
	}
	if crtRscFailed > 0 {
		err = errors.New("The creation of " + strconv.Itoa(crtRscFailed) + " " + rscLabel +
			" of the resource definition " + rscDfnData.ResourceDefinition.Name + " failed")
	}

	// Get the volume for the first node back from the LINSTOR server to determine the
	// device path of the volume
	vlm, err := ctrlConn.Resources.GetVolumes(clientCtx, rscDfnData.ResourceDefinition.Name, l.StorageNodeList[0], nil)
	if err != nil {
		return "", err
	}
	if len(vlm) < 1 {
		return "", errors.New("The volume list queried from the LINSTOR server contains no volumes")
	}

	return vlm[0].DevicePath, err
}

// Deletes the LINSTOR resource definition
func (l *Linstor) DeleteVolume() error {
	clientCtx := context.Background()
	logCfg := &client.LogCfg{Level: l.Loglevel}
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
