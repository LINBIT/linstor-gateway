// Package iscsi combines LINSTOR operations and the CRM operations to create highly available iSCSI targets.
package iscsi

import (
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/LINBIT/linstor-remote-storage/crmcontrol"
	"github.com/LINBIT/linstor-remote-storage/linstorcontrol"

	xmltree "github.com/beevik/etree"
)

// Default port for an iSCSI portal
const DFLT_ISCSI_PORTAL_PORT = 3260

func resourceName(iscsiTargetName string, lun uint8) string {
	return iscsiTargetName + "_lu" + strconv.Itoa(int(lun))
}

// ISCSI combines the information needed to create highly-available iSCSI targets.
// It contains a iSCSI target configuration and a LINSTOR configuration.
type ISCSI struct {
	Target  Target                 `json:"target,omitempty"`
	Linstor linstorcontrol.Linstor `json:"linstor,omitempty"`
}

type LUN struct {
	ID      uint8  `json:"id,omitempty"`
	SizeKiB uint64 `json:"size_kib,omitempty"`
}

// Target contains the information necessary for iSCSI targets.
type Target struct {
	Name      string `json:"name,omitempty"`
	IQN       string `json:"iqn,omitempty"`
	LUNs      []*LUN `json:"luns,omitempty"`
	ServiceIP net.IP `json:"service_ip,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Portals   string `json:"portals,omitempty"`
}

// CreateResource creates a new highly available iSCSI target
func (i *ISCSI) CreateResource() error {
	targetName, err := i.Target.iqnTarget()
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
		freeTid, ok := config.TidSet.GetFree(1, math.MaxInt16)
		if !ok {
			return errors.New("Failed to allocate a target ID for the new iSCSI target")
		}

		// Create a LINSTOR resource definition, volume definition and associated resources
		i.Linstor.ResourceName = resourceName(targetName, lu.ID)
		i.Linstor.SizeKiB = lu.SizeKiB
		res, err := i.Linstor.CreateVolume()
		if err != nil {
			return fmt.Errorf("LINSTOR volume operation failed, error: %v", err)
		}

		// Create CRM resources and constraints for the iSCSI services
		err = crmcontrol.CreateCrmLu(
			res.StorageNodeList,
			targetName,
			i.Target.ServiceIP,
			i.Target.IQN,
			lu.ID,
			res.DevicePath,
			i.Target.Username,
			i.Target.Password,
			i.Target.Portals,
			int16(freeTid),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteResource deletes a highly available iSCSI target
func (i *ISCSI) DeleteResource() error {
	targetName, err := i.Target.iqnTarget()
	if err != nil {
		return err
	}

	for _, lu := range i.Target.LUNs {
		// Delete the CRM resources for iSCSI LU, target, service IP addres, etc.
		err = crmcontrol.DeleteCrmLu(targetName, lu.ID)
		if err != nil {
			return err
		}

		// Delete the LINSTOR resource definition
		i.Linstor.ResourceName = resourceName(targetName, lu.ID)
	}
	return i.Linstor.DeleteVolume()
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
func (i *ISCSI) ProbeResource() (map[string]crmcontrol.LrmRunState, error) {
	targetName, err := i.Target.iqnTarget()
	if err != nil {
		return nil, err
	}

	rscStateMap := make(map[string]crmcontrol.LrmRunState)

	for _, lu := range i.Target.LUNs {
		tmpMap, err := crmcontrol.ProbeResource(targetName, lu.ID)
		if err != nil {
			return nil, err
		}

		// HACK: combine all the maps into one. the real solution would
		// be a more sensible data structure
		for k, v := range tmpMap {
			rscStateMap[k] = v
		}
	}

	return rscStateMap, nil
}

// ListResources lists existing iSCSI targets.
//
// Returns: CIB XML document tree, slice of Targets, error object
func ListResources() (*xmltree.Document, []*Target, error) {
	docRoot, err := crmcontrol.ReadConfiguration()
	if err != nil {
		return nil, nil, err
	}

	config, err := crmcontrol.ParseConfiguration(docRoot)
	if err != nil {
		return nil, nil, err
	}

	targets := make([]*Target, 0)

	// first, "convert" all targets
	for _, t := range config.TargetList {
		target := &Target{
			Name:     t.ID,
			IQN:      t.IQN,
			LUNs:     make([]*LUN, 0),
			Username: t.Username,
			Password: t.Password,
			Portals:  t.Portals,
		}

		targets = append(targets, target)
	}

	// then, "convert" and link LUs
	for _, l := range config.LuList {
		lun := &LUN{
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

	return docRoot, targets, nil
}

// modifyResourceTargetRole modifies the role of an existing iSCSI resource.
func (i *ISCSI) modifyResourceTargetRole(startFlag bool) error {
	targetName, err := i.Target.iqnTarget()
	if err != nil {
		return errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	for _, lu := range i.Target.LUNs {
		// Stop the CRM resources for iSCSI LU, target, service IP addres, etc.
		err = crmcontrol.ModifyCrmLuTargetRole(targetName, lu.ID, startFlag)
		if err != nil {
			return err
		}
	}

	return nil
}

// iqnTarget extracts the target name from an IQN string.
// e.g., in "iqn.2019-07.org.demo.filserver:filestorage", the "filestorage" part.
func (t *Target) iqnTarget() (string, error) {
	spl := strings.Split(t.IQN, ":")
	if len(spl) != 2 {
		return "", errors.New("Malformed argument '" + t.IQN + "'")
	}
	return spl[1], nil
}
