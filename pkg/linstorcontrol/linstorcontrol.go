// Package linstorcontrol allows creating and deleting LINSTOR resources/volumes.
// It is a higher level abstraction to the low level golinstor REST package.
package linstorcontrol

import (
	"context"
	"errors"
	"fmt"
	"sort"

	apiconsts "github.com/LINBIT/golinstor"
	"github.com/LINBIT/golinstor/client"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/common"
)

// Linstor is a struct containing the configuration that is needed to create or delete a LINSTOR resource.
type Linstor struct {
	*client.Client
}

type Resource struct {
	Name          string                `json:"resource_name,omitempty"`
	Volumes       []common.VolumeConfig `json:"volumes,omitempty"`
	ResourceGroup string                `json:"resource_group_name,omitempty"`
	FileSystem    string                `json:"file_system,omitempty"`
}

// CreateResult is a struct than is used as the result of a successful create action.
// It already contains the data that is most likely used by a consumer of a CreateVolume() call.
type CreateResult struct {
	// Linux device path (e.g., /dev/drbd1001)
	DevicePath string
	// List of nodes where the actual data got places (i.e., after autoplace)
	StorageNodes []string
}

func StatusFromResources(serviceCfgPath string, definition *client.ResourceDefinition, group *client.ResourceGroup, resources []client.ResourceWithVolumes) common.ResourceStatus {
	resourceState := common.Unknown
	service := common.ServiceStateStopped
	primary := ""
	nodes := make([]string, 0, len(resources))

	volumeByNumber := make(map[int][]*client.Volume)
	for _, nodeRsc := range resources {
		nodes = append(nodes, nodeRsc.NodeName)

		if nodeRsc.State.InUse {
			primary = nodeRsc.NodeName
		}

		for _, vol := range nodeRsc.Volumes {
			volumeByNumber[int(vol.VolumeNumber)] = append(volumeByNumber[int(vol.VolumeNumber)], &vol)
		}
	}

	if definition.Props[fmt.Sprintf("files%s", serviceCfgPath)] == "True" {
		service = common.ServiceStateStarted
	}

	volumes := make([]common.VolumeState, 0, len(volumeByNumber))
	for nr, deployedVols := range volumeByNumber {
		upToDate := 0
		diskful := 0
		for _, vol := range deployedVols {
			if vol.State.DiskState == "UpToDate" {
				diskful++
			}
			if vol.State.DiskState == "UpToDate" || vol.State.DiskState == "Diskless" {
				upToDate++
			}
		}

		aggregateState := common.ResourceStateBad
		if upToDate == len(deployedVols) && diskful >= int(group.SelectFilter.PlaceCount) {
			aggregateState = common.ResourceStateOK
		} else if upToDate > 0 {
			aggregateState = common.ResourceStateDegraded
		}

		log.WithFields(log.Fields{
			"resource":       definition.Name,
			"wantPlaceCount": group.SelectFilter.PlaceCount,
			"haveDiskful":    diskful,
		}).Tracef("deciding aggregateState %s", aggregateState)

		volumes = append(volumes, common.VolumeState{
			Number: nr,
			State:  aggregateState,
		})

		if resourceState < aggregateState {
			resourceState = aggregateState
		}
	}

	sort.Slice(volumes, func(i, j int) bool {
		return volumes[i].Number < volumes[j].Number
	})

	return common.ResourceStatus{
		State:   resourceState,
		Service: service,
		Primary: primary,
		Nodes:   nodes,
		Volumes: volumes,
	}
}

func Default(controllers []string) (*Linstor, error) {
	cli, err := client.NewClient(client.Log(log.StandardLogger()), client.Controllers(controllers))
	if err != nil {
		return nil, err
	}

	return &Linstor{Client: cli}, nil
}

// EnsureResource creates or updates the given resource.
// It returns three values:
// - The newly created resource definition
// - A slice of all resources that have been spawned from this resource
//   definition on the respective nodes
// - An error if one occurred, or nil
func (l *Linstor) EnsureResource(ctx context.Context, res Resource, mayExist bool) (*client.ResourceDefinition, *client.ResourceGroup, []client.ResourceWithVolumes, error) {
	logger := log.WithField("resource", res.Name)

	logger.Trace("ensure resource group exists")

	err := l.ResourceGroups.Create(ctx, client.ResourceGroup{
		Name: res.ResourceGroup,
	})
	if err != nil && !isErrAlreadyExists(err) {
		return nil, nil, nil, fmt.Errorf("failed to create resource group: %w", err)
	}

	rgroup, err := l.ResourceGroups.Get(ctx, res.ResourceGroup)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get resource group: %w", err)
	}

	logger.Trace("ensure resource definition exists")

	props := map[string]string{}

	// XXX: currently, LINSTOR requires auto-promote=yes when a file system is
	// to be created because it does not try to promote the resource itself.
	// So we fix that up here: set auto-promote to "yes" initially then change
	// it back to "no" once the resource is created.
	// This will change in a future version, remove this hack then.
	if res.FileSystem != "" {
		props[apiconsts.NamespcDrbdResourceOptions+"/auto-promote"] = "yes"
	} else {
		props[apiconsts.NamespcDrbdResourceOptions+"/auto-promote"] = "no"
	}

	props[apiconsts.NamespcDrbdResourceOptions+"/quorum"] = "majority"
	props[apiconsts.NamespcDrbdResourceOptions+"/on-no-quorum"] = "io-error"

	err = l.ResourceDefinitions.Create(ctx, client.ResourceDefinitionCreate{
		ResourceDefinition: client.ResourceDefinition{
			Name:              res.Name,
			ResourceGroupName: res.ResourceGroup,
			Props:             props,
		},
	})
	if err != nil {
		if (!mayExist && isErrAlreadyExists(err)) || !isErrAlreadyExists(err) {
			return nil, nil, nil, fmt.Errorf("failed to create resource definition: %w", err)
		}
	}

	for _, vol := range res.Volumes {
		logger.WithField("volNr", vol.Number).Trace("ensure volume definition exists")

		volProps := map[string]string{}
		if vol.FileSystem != "" {
			volProps[apiconsts.NamespcFilesystem+"/Type"] = vol.FileSystem
			volProps[apiconsts.NamespcFilesystem+"/MkfsParams"] = "-E root_owner=" + vol.FileSystemRootOwner.String()
		}
		err := l.ResourceDefinitions.CreateVolumeDefinition(ctx, res.Name, client.VolumeDefinitionCreate{
			VolumeDefinition: client.VolumeDefinition{
				VolumeNumber: int32(vol.Number),
				SizeKib:      vol.SizeKiB,
				Props:        volProps,
			},
		})
		if err != nil && !isErrAlreadyExists(err) {
			return nil, nil, nil, fmt.Errorf("failed to ensure volume definition: %w", err)
		}
	}

	logger.Trace("ensure resource is placed")

	err = l.Resources.Autoplace(ctx, res.Name, client.AutoPlaceRequest{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to autoplace resources: %w", err)
	}

	// XXX: remove this when LINSTOR supports this (see comment above).
	if res.FileSystem != "" {
		err = l.ResourceDefinitions.Modify(ctx, res.Name, client.GenericPropsModify{
			OverrideProps: map[string]string{
				apiconsts.NamespcDrbdResourceOptions + "/auto-promote": "no",
			},
		})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to update properties of resource definition '%s': %w", res.Name, err)
		}
	}

	logger.Trace("fetch existing resource definition")

	rdef, err := l.ResourceDefinitions.Get(ctx, res.Name)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch existing resource definition: %w", err)
	}

	logger.Trace("fetch existing resources")

	view, err := l.Resources.GetResourceView(ctx, &client.ListOpts{Resource: []string{res.Name}})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch resource view: %w", err)
	}

	if len(view) == 0 {
		return nil, nil, nil, errors.New(fmt.Sprintf("failed to fetch resource '%s'", res.Name))
	}

	for _, existingVol := range view[0].Volumes {
		logger.WithField("volNr", existingVol.VolumeNumber).Trace("ensure existing volume is defined")

		expected := false
		for _, expectedVol := range res.Volumes {
			if int(existingVol.VolumeNumber) == expectedVol.Number {
				expected = true
				break
			}
		}
		if !expected {
			err := l.ResourceDefinitions.DeleteVolumeDefinition(ctx, res.Name, int(existingVol.VolumeNumber))
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to delete unexpected volume definition: %w", err)
			}
		}
	}

	return &rdef, &rgroup, view, nil
}

func isErrAlreadyExists(err error) bool {
	if err == nil {
		return false
	}

	apiErr, ok := err.(client.ApiCallError)
	if !ok {
		return false
	}

	possibleErrs := []uint64{
		apiconsts.FailExistsNode,
		apiconsts.FailExistsRscDfn,
		apiconsts.FailExistsRsc,
		apiconsts.FailExistsVlmDfn,
		apiconsts.FailExistsVlm,
		apiconsts.FailExistsNetIf,
		apiconsts.FailExistsNodeConn,
		apiconsts.FailExistsRscConn,
		apiconsts.FailExistsVlmConn,
		apiconsts.FailExistsStorPoolDfn,
		apiconsts.FailExistsStorPool,
		apiconsts.FailExistsStltConn,
		apiconsts.FailExistsCryptPassphrase,
		apiconsts.FailExistsWatch,
		apiconsts.FailExistsSnapshotDfn,
		apiconsts.FailExistsSnapshot,
		apiconsts.FailExistsExtName,
		apiconsts.FailExistsNvmeTargetPerRscDfn,
		apiconsts.FailExistsNvmeInitiatorPerRscDfn,
		apiconsts.FailLostStorPool,
		apiconsts.FailExistsRscGrp,
		apiconsts.FailExistsVlmGrp,
		apiconsts.FailExistsOpenflexTargetPerRscDfn,
		apiconsts.FailExistsSnapshotShipping,
		apiconsts.FailExistsExosEnclosure,
	}

	for _, e := range possibleErrs {
		if apiErr.Is(e) {
			return true
		}
	}

	return false
}
