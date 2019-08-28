// Package iscsi combines LINSTOR operations and the CRM operations to create highly available iSCSI targets.
package iscsi

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/LINBIT/linstor-iscsi/pkg/crmcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/targetutil"
)

// Default port for an iSCSI portal
const DFLT_ISCSI_PORTAL_PORT = 3260

// ISCSI combines the information needed to create highly-available iSCSI targets.
// It contains a iSCSI target configuration and a LINSTOR configuration.
type ISCSI struct {
	Target  targetutil.Target      `json:"target,omitempty"`
	Linstor linstorcontrol.Linstor `json:"linstor,omitempty"`
}

// CreateResource creates a new highly available iSCSI target
func (i *ISCSI) CreateResource() error {
	targetName, err := targetutil.ExtractTargetName(i.Target.IQN)
	if err != nil {
		return err
	}

	for _, lu := range i.Target.LUNs {
		// Read the current configuration from the CRM
		docRoot, err := crmcontrol.ReadConfiguration()
		if err != nil {
			return err
		}
		// Find resources, allocated target IDs, etc.
		config, err := crmcontrol.ParseConfiguration(docRoot)
		if err != nil {
			return err
		}

		// Find a free target ID number using the set of allocated target IDs
		freeTid, ok := config.TIDs.GetFree(1, math.MaxInt16)
		if !ok {
			return errors.New("Failed to allocate a target ID for the new iSCSI target")
		}

		// Create a LINSTOR resource definition, volume definition and associated resources
		i.Linstor.ResourceName = linstorcontrol.ResourceNameFromLUN(targetName, lu.ID)
		i.Linstor.SizeKiB = lu.SizeKiB
		res, err := i.Linstor.CreateVolume()
		if err != nil {
			return fmt.Errorf("LINSTOR volume operation failed, error: %v", err)
		}

		// Create CRM resources and constraints for the iSCSI services
		err = crmcontrol.CreateCrmLu(i.Target, res.StorageNodeList,
			res.DevicePath, int16(freeTid))
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteResource deletes a highly available iSCSI target
func (i *ISCSI) DeleteResource() error {
	targetName, err := targetutil.ExtractTargetName(i.Target.IQN)
	if err != nil {
		return err
	}

	for _, lu := range i.Target.LUNs {
		var errs []string
		// Delete the CRM resources for iSCSI LU, target, service IP addres, etc.
		if err = crmcontrol.DeleteCrmLu(i.Target.IQN, lu.ID); err != nil {
			errs = append(errs, err.Error())
		}

		// Delete the LINSTOR resource definition
		i.Linstor.ResourceName = linstorcontrol.ResourceNameFromLUN(targetName, lu.ID)
		if err = i.Linstor.DeleteVolume(); err != nil {
			errs = append(errs, err.Error())
		}

		if len(errs) > 0 {
			return fmt.Errorf(strings.Join(errs, "\n"))
		}
	}
	return nil
}

// StartResource starts an existing iSCSI resource.
func (i *ISCSI) StartResource() error {
	return i.modifyResourceTargetRole(true)
}

// StopResource stops an existing iSCSI resource.
func (i *ISCSI) StopResource() error {
	return i.modifyResourceTargetRole(false)
}

// ProbeResource gets information about an existing iSCSI resource.
// It returns a resource state map and an error.
func (i *ISCSI) ProbeResource() (crmcontrol.ResourceRunState, error) {
	luns := make([]uint8, len(i.Target.LUNs))
	for i, lu := range i.Target.LUNs {
		luns[i] = lu.ID
	}

	return crmcontrol.ProbeResource(i.Target.IQN, luns)
}

// ListResources lists existing iSCSI targets.
//
// It returns a slice of Targets and an error object
func ListResources() ([]*targetutil.Target, error) {
	docRoot, err := crmcontrol.ReadConfiguration()
	if err != nil {
		return nil, err
	}

	config, err := crmcontrol.ParseConfiguration(docRoot)
	if err != nil {
		return nil, err
	}

	targets := make([]*targetutil.Target, 0)

	// first, "convert" all targets
	for _, t := range config.Targets {
		targetCfg := targetutil.TargetConfig{
			IQN:      t.IQN,
			LUNs:     make([]*targetutil.LUN, 0),
			Username: t.Username,
			Password: t.Password,
			Portals:  t.Portals,
		}

		target, err := targetutil.NewTarget(targetCfg)
		if err != nil {
			return nil, err
		}

		targets = append(targets, &target)
	}

	// then, "convert" and link LUs
	for _, l := range config.LUs {
		lun := &targetutil.LUN{
			ID: l.LUN,
		}

		// link to the correct target
		for _, t := range targets {
			if t.IQN == l.Target.IQN {
				t.LUNs = append(t.LUNs, lun)
				break
			}
		}
	}

	return targets, nil
}

// modifyResourceTargetRole modifies the role of an existing iSCSI resource.
func (i *ISCSI) modifyResourceTargetRole(startFlag bool) error {
	luns := make([]uint8, len(i.Target.LUNs))
	for i, lu := range i.Target.LUNs {
		luns[i] = lu.ID
	}
	if startFlag {
		return crmcontrol.StartCrmResource(i.Target.IQN, luns)
	} else {
		return crmcontrol.StopCrmResource(i.Target.IQN, luns)
	}
}
