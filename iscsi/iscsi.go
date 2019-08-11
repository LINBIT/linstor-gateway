// High-level API - entry points for resource creation/deletion/etc.
package iscsi

// iscsi module
//
// This module combines the LINSTOR operations (in package linstorcontrol)
// and the CRM operations (in package crmcontrol) to form a combined high-level API
// that performs each operation in both subsystems.

import (
	"errors"
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

type Target struct {
	IQN                string
	LUN                uint8
	ServiceIP          net.IP
	Username, Password string
	Portals            string
}

// Creates a new LINSTOR & iSCSI resource
//
// Returns: program exit code, error object
func CreateResource(target *Target, linstor *linstorcontrol.Linstor) error {
	targetName, err := target.iqnTarget()
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
	freeTid, haveFreeTid := crmcontrol.GetFreeTargetId(config.TidSet.SortedKeys())
	if !haveFreeTid {
		return errors.New("Failed to allocate a target ID for the new iSCSI target")
	}

	// Create a LINSTOR resource definition, volume definition and associated resources
	linstor.ResourceName = ResourceName(targetName, target.LUN)
	devPath, err := linstor.CreateVolume()
	if err != nil {
		return errors.New("LINSTOR volume operation failed, error: " + err.Error())
	}

	// Create CRM resources and constraints for the iSCSI services
	err = crmcontrol.CreateCrmLu(
		linstor.StorageNodeList,
		targetName,
		target.ServiceIP,
		target.IQN,
		target.LUN,
		devPath,
		target.Username,
		target.Password,
		target.Portals,
		freeTid,
	)
	if err != nil {
		return err
	}

	return nil
}

// Deletes existing LINSTOR & iSCSI resources
//
// Returns: program exit code, error object
func DeleteResource(target *Target, linstor *linstorcontrol.Linstor) error {
	targetName, err := target.iqnTarget()
	if err != nil {
		return err
	}

	// Delete the CRM resources for iSCSI LU, target, service IP addres, etc.
	err = crmcontrol.DeleteCrmLu(targetName, target.LUN)
	if err != nil {
		return err
	}

	// Delete the LINSTOR resource definition
	linstor.ResourceName = ResourceName(targetName, target.LUN)
	return linstor.DeleteVolume()
}

// Starts existing iSCSI resources
//
// Returns: program exit code, error object
func StartResource(target *Target) error {
	return modifyResourceTargetRole(target, true)
}

// Stops existing iSCSI resources
//
// Returns: program exit code, error object
func StopResource(target *Target) error {
	return modifyResourceTargetRole(target, false)
}

// Starts/stops existing iSCSI resources
//
// Returns: resource state map, program exit code, error object
func ProbeResource(target *Target) (*map[string]crmcontrol.LrmRunState, error) {
	targetName, err := target.iqnTarget()
	if err != nil {
		return nil, err
	}

	rscStateMap, err := crmcontrol.ProbeResource(targetName, target.LUN)
	if err != nil {
		return nil, err
	}

	return &rscStateMap, nil
}

// Extracts a list of existing CRM (Pacemaker) resources from the CIB XML
//
// Returns: CIB XML document tree, CrmConfiguration object, program exit code, error object
func ListResources() (*xmltree.Document, *crmcontrol.CrmConfiguration, error) {
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
func modifyResourceTargetRole(target *Target, startFlag bool) error {
	targetName, err := target.iqnTarget()
	if err != nil {
		return errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	// Stop the CRM resources for iSCSI LU, target, service IP addres, etc.
	err = crmcontrol.ModifyCrmLuTargetRole(targetName, target.LUN, startFlag)
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
