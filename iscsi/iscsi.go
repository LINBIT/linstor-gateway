// High-level API - entry points for resource creation/deletion/etc.
package iscsi

// iscsi module
//
// This module combines the LINSTOR operations (in package linstorcontrol)
// and the CRM operations (in package crmcontrol) to form a combined high-level API
// that performs each operation in both subsystems.

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

func ResourceName(iscsiTargetName string, lun uint8) string {
	return iscsiTargetName + "_lu" + strconv.Itoa(int(lun))
}

type ISCSI struct {
	Target  Target                 `json:"target,omitempty"`
	Linstor linstorcontrol.Linstor `json:"linstor,omitempty"`
}

type Target struct {
	IQN       string `json:"iqn,omitempty"`
	LUN       uint8  `json:"lun,omitempty"`
	ServiceIP net.IP `json:"service_ip,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Portals   string `json:"portals,omitempty"`
}

// Creates a new LINSTOR & iSCSI resource
//
// Returns: program exit code, error object
func (i *ISCSI) CreateResource() error {
	targetName, err := i.Target.iqnTarget()
	if err != nil {
		return err
	}

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
	i.Linstor.ResourceName = ResourceName(targetName, i.Target.LUN)
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
		i.Target.LUN,
		res.DevicePath,
		i.Target.Username,
		i.Target.Password,
		i.Target.Portals,
		int16(freeTid),
	)
	if err != nil {
		return err
	}

	return nil
}

// Deletes existing LINSTOR & iSCSI resources
//
// Returns: program exit code, error object
func (i *ISCSI) DeleteResource() error {
	targetName, err := i.Target.iqnTarget()
	if err != nil {
		return err
	}

	// Delete the CRM resources for iSCSI LU, target, service IP addres, etc.
	err = crmcontrol.DeleteCrmLu(targetName, i.Target.LUN)
	if err != nil {
		return err
	}

	// Delete the LINSTOR resource definition
	i.Linstor.ResourceName = ResourceName(targetName, i.Target.LUN)
	return i.Linstor.DeleteVolume()
}

// Starts existing iSCSI resources
//
// Returns: program exit code, error object
func (i *ISCSI) StartResource() error {
	return i.modifyResourceTargetRole(true)
}

// Stops existing iSCSI resources
//
// Returns: program exit code, error object
func (i *ISCSI) StopResource() error {
	return i.modifyResourceTargetRole(false)
}

// Starts/stops existing iSCSI resources
//
// Returns: resource state map, program exit code, error object
func (i *ISCSI) ProbeResource() (*map[string]crmcontrol.LrmRunState, error) {
	targetName, err := i.Target.iqnTarget()
	if err != nil {
		return nil, err
	}

	rscStateMap, err := crmcontrol.ProbeResource(targetName, i.Target.LUN)
	if err != nil {
		return nil, err
	}

	return &rscStateMap, nil
}

// Extracts a list of existing CRM (Pacemaker) resources from the CIB XML
//
// Returns: CIB XML document tree, CrmConfiguration object, program exit code, error object
func (i *ISCSI) ListResources() (*xmltree.Document, *crmcontrol.CrmConfiguration, error) {
	docRoot, err := crmcontrol.ReadConfiguration()
	if err != nil {
		return nil, nil, err
	}

	config, err := crmcontrol.ParseConfiguration(docRoot)
	if err != nil {
		return nil, nil, err
	}

	return docRoot, config, nil
}

// Starts/stops existing iSCSI resources
//
// Returns: program exit code, error object
func (i *ISCSI) modifyResourceTargetRole(startFlag bool) error {
	targetName, err := i.Target.iqnTarget()
	if err != nil {
		return errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	// Stop the CRM resources for iSCSI LU, target, service IP addres, etc.
	err = crmcontrol.ModifyCrmLuTargetRole(targetName, i.Target.LUN, startFlag)
	if err != nil {
		return err
	}

	return nil
}

// Extracts the target name from an IQN string
//
// e.g., in "iqn.2019-07.org.demo.filserver:filestorage", the "filestorage" part
func (t *Target) iqnTarget() (string, error) {
	spl := strings.Split(t.IQN, ":")
	if len(spl) != 2 {
		return "", errors.New("Malformed argument '" + t.IQN + "'")
	} else {
		return spl[1], nil
	}
}
