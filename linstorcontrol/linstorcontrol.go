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
	"io"
	"strconv"

	client "github.com/LINBIT/golinstor/client"
	"github.com/sirupsen/logrus"
)

const (
	DEBUG_LINSTOR_CONTROLLERS = "10.43.9.28:3370"
)

// Creates a LINSTOR resource definition, volume definition and associated resources on the selected nodes
func CreateVolume(
	iscsiTargetName string,
	lun uint8,
	vlmSizeKiB uint64,
	storageNodeList []string,
	clientNodeList []string,
	autoPlaceCount uint64,
	storageStorPool string,
	clientStorPool string,
	logStream io.Writer,
) (string, error) {
	if len(storageNodeList) < 1 {
		return "", errors.New("Invalid CreateVolume() call: Parameter storageNodeList is an empty list")
	}

	clientCtx := context.Background()
	logCfg := &client.LogCfg{Level: logrus.TraceLevel.String()}
	ctrlConn, err := client.NewClient(client.Log(logCfg))
	if err != nil {
		return "", err
	}

	// Create a resource definition
	rscDfnData := client.ResourceDefinitionCreate{
		ResourceDefinition: client.ResourceDefinition{Name: iscsiTargetName + "_lu" + strconv.Itoa(int(lun))},
	}
	err = ctrlConn.ResourceDefinitions.Create(clientCtx, rscDfnData)
	if err != nil {
		return "", err
	}

	// Create a volume definition
	vlmDfnData := client.VolumeDefinitionCreate{
		VolumeDefinition: client.VolumeDefinition{VolumeNumber: int32(0), SizeKib: vlmSizeKiB},
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
	for _, tgtNodeName := range storageNodeList {
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
	vlm, err := ctrlConn.Resources.GetVolumes(clientCtx, rscDfnData.ResourceDefinition.Name, storageNodeList[0], nil)
	if err != nil {
		return "", err
	}
	if len(vlm) < 1 {
		return "", errors.New("The volume list queried from the LINSTOR server contains no volumes")
	}

	return vlm[0].DevicePath, err
}

// Deletes the LINSTOR resource definition
func DeleteVolume(
	iscsiTargetName string,
	lun uint8,
) error {
	clientCtx := context.Background()
	logCfg := &client.LogCfg{Level: logrus.TraceLevel.String()}
	ctrlConn, err := client.NewClient(client.Log(logCfg))
	if err != nil {
		return err
	}

	luName := "lu" + strconv.Itoa(int(lun))
	return ctrlConn.ResourceDefinitions.Delete(clientCtx, iscsiTargetName+"_"+luName)
}

type debugFormatter struct {
	logrus.Formatter
}

// Implements the logrus.Formatter interface; provisional implementation; used for golinstor
func (formatter *debugFormatter) format(logEntry *logrus.Entry) ([]byte, error) {
	var output []byte = make([]byte, 0)
	output = append(output, (*logEntry).Message...)
	return output, nil
}
