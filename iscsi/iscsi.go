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

// Creates a new LINSTOR & iSCSI resource
//
// Returns: program exit code, error object
func CreateResource(
	iqn string,
	lun uint8,
	sizeKib uint64,
	storageNodeList []string,
	clientNodeList []string,
	serviceIp net.IP,
	username, password, portals, loglevel string, controllerIP net.IP) error {
	targetName, err := iqnExtractTarget(iqn)
	if err != nil {
		return errors.New("Invalid IQN format: Missing ':' separator and target name")
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
	freeTid, haveFreeTid := crmcontrol.GetFreeTargetId(config.TidSet.ToSortedArray())
	if !haveFreeTid {
		return errors.New("Failed to allocate a target ID for the new iSCSI target")
	}

	// Create a LINSTOR resource definition, volume definition and associated resources
	resourceName := ResourceName(targetName, lun)
	devPath, err := linstorcontrol.CreateVolume(
		resourceName,
		sizeKib,
		storageNodeList,
		make([]string, 0),
		uint64(0),
		"",
		"",
		loglevel, controllerIP)
	if err != nil {
		return errors.New("LINSTOR volume operation failed, error: " + err.Error())
	}

	// Create CRM resources and constraints for the iSCSI services
	err = crmcontrol.CreateCrmLu(
		storageNodeList,
		targetName,
		serviceIp,
		iqn,
		uint8(lun),
		devPath,
		username,
		password,
		portals,
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
func DeleteResource(iqn string, lun uint8, loglevel string, controllerIP net.IP) error {
	targetName, err := iqnExtractTarget(iqn)
	if err != nil {
		return errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	// Delete the CRM resources for iSCSI LU, target, service IP addres, etc.
	err = crmcontrol.DeleteCrmLu(targetName, lun)
	if err != nil {
		return err
	}

	// Delete the LINSTOR resource definition
	resourceName := ResourceName(targetName, lun)
	err = linstorcontrol.DeleteVolume(resourceName, loglevel, controllerIP)

	return nil
}

// Starts existing iSCSI resources
//
// Returns: program exit code, error object
func StartResource(iqn string, lun uint8) error {
	return modifyResourceTargetRole(iqn, lun, true)
}

// Stops existing iSCSI resources
//
// Returns: program exit code, error object
func StopResource(iqn string, lun uint8) error {
	return modifyResourceTargetRole(iqn, lun, false)
}

// Starts/stops existing iSCSI resources
//
// Returns: resource state map, program exit code, error object
func ProbeResource(iqn string, lun uint8) (*map[string]crmcontrol.LrmRunState, error) {
	targetName, err := iqnExtractTarget(iqn)
	if err != nil {
		return nil, errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	rscStateMap, err := crmcontrol.ProbeResource(targetName, lun)
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
func modifyResourceTargetRole(iqn string, lun uint8, startFlag bool) error {
	targetName, err := iqnExtractTarget(iqn)
	if err != nil {
		return errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	// Stop the CRM resources for iSCSI LU, target, service IP addres, etc.
	err = crmcontrol.ModifyCrmLuTargetRole(targetName, lun, startFlag)
	if err != nil {
		return err
	}

	return nil
}

// Extracts the target name from an IQN string
//
// e.g., in "iqn.2019-07.org.demo.filserver:filestorage", the "filestorage" part
func iqnExtractTarget(iqn string) (string, error) {
	var target string
	var err error = nil
	idx := strings.IndexByte(iqn, ':')
	if idx != -1 {
		target = iqn[idx+1:]
	} else {
		err = errors.New("Malformed argument '" + iqn + "'")
	}
	return target, err
}
