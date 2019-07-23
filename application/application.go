package application

import "os"
import "fmt"
import "strings"
import "strconv"
import "errors"
import "github.com/LINBIT/linstor-remote-storage/linstorcontrol"
import "github.com/LINBIT/linstor-remote-storage/crmcontrol"

const (
	EXIT_SUCCESS       = 0
	EXIT_INV_PRM       = 1
	EXIT_FAILED_ACTION = 2
)

const (
	ACTION_CREATE = "create"
	ACTION_DELETE = "delete"
	ACTION_LIST   = "list"
)

const (
	KEY_TARGET   = "target"
	KEY_SVC_IP   = "ip"
	KEY_IQN      = "iqn"
	KEY_LUN      = "lun"
	KEY_USERNAME = "username"
	KEY_PASSWORD = "password"
	KEY_PORTALS  = "portals"
	KEY_NODES    = "nodes"
	KEY_SIZE     = "size"
)

var CREATE_PARAMS []string = []string{
	KEY_SVC_IP,
	KEY_IQN,
	KEY_LUN,
	KEY_USERNAME,
	KEY_PASSWORD,
	KEY_PORTALS,
	KEY_NODES,
	KEY_SIZE,
}

var DELETE_PARAMS []string = []string{
	KEY_IQN,
	KEY_LUN,
}

func ActionCreate() (int, error) {
	if explainParams(ACTION_CREATE, CREATE_PARAMS) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	loadParams(CREATE_PARAMS, argMap)

	err := parseArguments(&argMap)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid command line: " + err.Error())
	}

	lun, err := strconv.ParseUint(argMap[KEY_LUN], 10, 8)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_LUN + "': Unparseable logical unit number")
	}

	sizeKiB, err := strconv.ParseUint(argMap[KEY_SIZE], 10, 64)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_SIZE + "': Unparseable volume size")
	}

	storageNodeList := strings.Split(argMap[KEY_NODES], ",")

	targetName, err := iqnExtractTarget(argMap[KEY_IQN])
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_IQN + "': Invalid IQN format: Missing ':' separator and target name")
	}

	// Read the current configuration from the CRM
	config, err := crmcontrol.ReadConfiguration()
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
		sizeKiB,
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
		argMap[KEY_SVC_IP],
		argMap[KEY_IQN],
		uint8(lun),
		devPath,
		argMap[KEY_USERNAME],
		argMap[KEY_PASSWORD],
		argMap[KEY_PORTALS],
		int16(freeTid),
	)
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}

	return EXIT_SUCCESS, nil
}

func ActionDelete() (int, error) {
	if explainParams(ACTION_DELETE, DELETE_PARAMS) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	loadParams(DELETE_PARAMS, argMap)

	err := parseArguments(&argMap)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid command line: " + err.Error())
	}

	lun, err := strconv.ParseUint(argMap[KEY_LUN], 10, 8)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_LUN + "': Unparseable logical unit number")
	}

	targetName, err := iqnExtractTarget(argMap[KEY_IQN])
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_IQN + "': Invalid IQN format: Missing ':' separator and target name")
	}

	err = crmcontrol.DeleteCrmLu(targetName, uint8(lun))
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}

	return EXIT_SUCCESS, nil
}

func ActionList() (int, error) {
	_, err := crmcontrol.ReadConfiguration()
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}

	return EXIT_SUCCESS, nil
}

func parseArguments(argMap *map[string]string) error {
	collectedArgs := make(map[string]*string)
	for key, _ := range *argMap {
		collectedArgs[key] = nil
	}

	argCount := len(os.Args)
	for idx := 2; idx < argCount; idx++ {
		arg := os.Args[idx]
		keyPtr, valuePtr, err := splitArg(arg)
		if err != nil {
			return err
		}
		key := *keyPtr
		value := *valuePtr
		mapValue, found := collectedArgs[key]
		if !found {
			return errors.New("Invalid argument key '" + key + "'")
		}
		if mapValue != nil {
			return errors.New("Duplicate argument '" + key + "'")
		}
		collectedArgs[key] = &value
	}

	for key, value := range collectedArgs {
		if value != nil {
			(*argMap)[key] = *value
		} else {
			return errors.New("Missing argument '" + key + "'")
		}
	}

	return nil
}

func splitArg(arg string) (*string, *string, error) {
	var key string
	var value string
	var err error = nil
	idx := strings.IndexByte(arg, '=')
	if idx != -1 {
		key = arg[:idx]
		value = arg[idx+1:]
	} else {
		err = errors.New("Malformed argument '" + arg + "'")
	}
	return &key, &value, err
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

func loadParams(paramList []string, argMap map[string]string) {
	for _, key := range paramList {
		argMap[key] = ""
	}
}

func explainParams(action string, paramList []string) bool {
	explainFlag := len(os.Args) < 3
	if explainFlag {
		fmt.Printf("Parameters required for action %s:\n", action)
		for _, entry := range paramList {
			fmt.Printf("    %s\n", entry)
		}
		fmt.Printf("\n")
	}
	return explainFlag
}
