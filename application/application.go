package application

import "strings"
import "errors"
import "github.com/LINBIT/linstor-remote-storage/linstorcontrol"
import "github.com/LINBIT/linstor-remote-storage/crmcontrol"
import xmltree "github.com/beevik/etree"

const (
	EXIT_SUCCESS       = 0
	EXIT_INV_PRM       = 1
	EXIT_FAILED_ACTION = 2
)

func CreateResource(
	iqn string,
	lun uint8,
	sizeKib uint64,
	storageNodeList []string,
	clientNodeList []string,
	serviceIp string,
	username string,
	password string,
	portals string,
) (int, error) {
	targetName, err := iqnExtractTarget(iqn)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	// Read the current configuration from the CRM
	docRoot, err := crmcontrol.ReadConfiguration()
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}
	config, err := crmcontrol.ParseConfiguration(docRoot)
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}

	freeTid, haveFreeTid := crmcontrol.GetFreeTargetId(config.TidSet.ToSortedArray())
	if !haveFreeTid {
		return EXIT_FAILED_ACTION, errors.New("Failed to allocate a target ID for the new iSCSI target")
	}

	devPath, err := linstorcontrol.CreateVolume(
		targetName,
		uint8(lun),
		sizeKib,
		storageNodeList,
		make([]string, 0),
		uint64(0),
		"",
		"",
		nil,
	)
	if err != nil {
		return EXIT_FAILED_ACTION, errors.New("LINSTOR volume operation failed, error: " + err.Error())
	}

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

func DeleteResource(
	iqn string,
	lun uint8,
) (int, error) {
	targetName, err := iqnExtractTarget(iqn)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	err = crmcontrol.DeleteCrmLu(targetName, lun)
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}

	err = linstorcontrol.DeleteVolume(targetName, lun)

	return EXIT_SUCCESS, nil
}

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
