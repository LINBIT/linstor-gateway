// High-level API - entry points for resource creation/deletion/etc.
package application

// application module
//
// This module combines the LINSTOR operations (in package linstorcontrol)
// and the CRM operations (in package crmcontrol) to form a combined high-level API
// that performs each operation in both subsystems.

import (
	"errors"
	"net"
	"strings"

	"github.com/LINBIT/linstor-remote-storage/crmcontrol"
	"github.com/LINBIT/linstor-remote-storage/linstorcontrol"

	xmltree "github.com/beevik/etree"
)

const (
	// Indicates successful completion of the application
	EXIT_SUCCESS = 0
	// Indicates failure due to an invalid parameter
	EXIT_INV_PRM = 1
	// Indicates failure due to a failed action, e.g. failure to create a volume
	EXIT_FAILED_ACTION = 2
)

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
	username, password, portals, loglevel string, controllerIP net.IP) (int, error) {
	targetName, err := iqnExtractTarget(iqn)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	// Read the current configuration from the CRM
	docRoot, err := crmcontrol.ReadConfiguration()
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}
	// Find resources, allocated target IDs, etc.
	config, err := crmcontrol.ParseConfiguration(docRoot)
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}

	// Find a free target ID number using the set of allocated target IDs
	freeTid, haveFreeTid := crmcontrol.GetFreeTargetId(config.TidSet.ToSortedArray())
	if !haveFreeTid {
		return EXIT_FAILED_ACTION, errors.New("Failed to allocate a target ID for the new iSCSI target")
	}

	// Create a LINSTOR resource definition, volume definition and associated resources
	devPath, err := linstorcontrol.CreateVolume(
		targetName,
		uint8(lun),
		sizeKib,
		storageNodeList,
		make([]string, 0),
		uint64(0),
		"",
		"",
		loglevel, controllerIP)
	if err != nil {
		return EXIT_FAILED_ACTION, errors.New("LINSTOR volume operation failed, error: " + err.Error())
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
		return EXIT_FAILED_ACTION, err
	}

	return EXIT_SUCCESS, nil
}

// Deletes existing LINSTOR & iSCSI resources
//
// Returns: program exit code, error object
func DeleteResource(iqn string, lun uint8, loglevel string, controllerIP net.IP) (int, error) {
	targetName, err := iqnExtractTarget(iqn)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	// Delete the CRM resources for iSCSI LU, target, service IP addres, etc.
	err = crmcontrol.DeleteCrmLu(targetName, lun)
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}

	// Delete the LINSTOR resource definition
	err = linstorcontrol.DeleteVolume(targetName, lun, loglevel, controllerIP)

	return EXIT_SUCCESS, nil
}

// Starts existing iSCSI resources
//
// Returns: program exit code, error object
func StartResource(
	iqn string,
	lun uint8,
) (int, error) {
	return modifyResourceTargetRole(iqn, lun, true)
}

// Stops existing iSCSI resources
//
// Returns: program exit code, error object
func StopResource(
	iqn string,
	lun uint8,
) (int, error) {
	return modifyResourceTargetRole(iqn, lun, false)
}

// Starts/stops existing iSCSI resources
//
// Returns: resource state map, program exit code, error object
func ProbeResource(
	iqn string,
	lun uint8,
) (*map[string]crmcontrol.LrmRunState, int, error) {
	targetName, err := iqnExtractTarget(iqn)
	if err != nil {
		return nil, EXIT_INV_PRM, errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	rscStateMap, err := crmcontrol.ProbeResource(targetName, lun)
	if err != nil {
		return nil, EXIT_FAILED_ACTION, err
	}

	return &rscStateMap, EXIT_SUCCESS, nil
}

// Extracts a list of existing CRM (Pacemaker) resources from the CIB XML
//
// Returns: CIB XML document tree, CrmConfiguration object, program exit code, error object
func ListResources() (*xmltree.Document, *crmcontrol.CrmConfiguration, int, error) {
	docRoot, err := crmcontrol.ReadConfiguration()
	if err != nil {
		return nil, nil, EXIT_FAILED_ACTION, err
	}

	config, err := crmcontrol.ParseConfiguration(docRoot)
	if err != nil {
		return nil, nil, EXIT_FAILED_ACTION, err
	}

	return docRoot, config, EXIT_SUCCESS, nil
}

// Starts/stops existing iSCSI resources
//
// Returns: program exit code, error object
func modifyResourceTargetRole(
	iqn string,
	lun uint8,
	startFlag bool,
) (int, error) {
	targetName, err := iqnExtractTarget(iqn)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	// Stop the CRM resources for iSCSI LU, target, service IP addres, etc.
	err = crmcontrol.ModifyCrmLuTargetRole(targetName, lun, startFlag)
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}

	return EXIT_SUCCESS, nil
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
