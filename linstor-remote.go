package main

import "os"
import "fmt"
import "strings"
import "strconv"
import "errors"
import "github.com/LINBIT/linstor-remote-storage/linstorcontrol"
import "github.com/LINBIT/linstor-remote-storage/crmcontrol"

const (
	KEY_TARGET   = "target"
	KEY_SVC_IP   = "ip"
	KEY_IQN      = "iqn"
	KEY_LUN      = "lun"
	KEY_USERNAME = "username"
	KEY_PASSWORD = "password"
	KEY_PORTALS  = "portals"
	KEY_TID      = "tid"
	KEY_NODES    = "nodes"
	KEY_SIZE     = "size"
)

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

var CREATE_PARAMS []string = []string{
	KEY_SVC_IP,
	KEY_IQN,
	KEY_LUN,
	KEY_USERNAME,
	KEY_PASSWORD,
	KEY_PORTALS,
	KEY_TID,
	KEY_NODES,
	KEY_SIZE,
}

var DELETE_PARAMS []string = []string{
	KEY_IQN,
	KEY_LUN,
}

func main() {
	fmt.Printf("\x1b[1;33m")
	fmt.Printf("linstor-remote experimental 2019-07-181 17:37")
	fmt.Printf("\x1b[0m\n")

	argsCount := len(os.Args)
	fmt.Printf("%d arguments:\n", argsCount)
	for idx, arg := range os.Args {
		fmt.Printf("    [%3d] %s\n", idx, arg)
	}
	fmt.Printf("\n")

	var err error = nil
	argCount := len(os.Args)
	if argCount >= 2 {
		action := os.Args[1]
		switch action {
		case ACTION_CREATE:
			err = actionCreate()
		case ACTION_DELETE:
			err = actionDelete()
		case ACTION_LIST:
			err = actionList()
		default:
			err = errors.New("Action '" + action + "' is not implemented")
		}
	}

	if err != nil {
		fmt.Printf("Operation failed, error: %s\n", err.Error())
		os.Exit(EXIT_FAILED_ACTION)
	}
}

func actionCreate() error {
	if explainParams(ACTION_CREATE, CREATE_PARAMS) {
		return nil
	}

	argMap := make(map[string]string)
	loadParams(CREATE_PARAMS, argMap)

	err := parseArguments(&argMap)
	if err != nil {
		fmt.Printf("Invalid command line: %s\n", err.Error())
		os.Exit(EXIT_INV_PRM)
	}

	lun, err := strconv.ParseUint(argMap[KEY_LUN], 10, 8)
	if err != nil {
		fmt.Printf("Argument '%s': Unparseable logical unit number", KEY_LUN)
		os.Exit(EXIT_INV_PRM)
	}

	sizeKiB, err := strconv.ParseUint(argMap[KEY_SIZE], 10, 64)
	if err != nil {
		fmt.Printf("Argument '%s': Unparseable volume size", KEY_SIZE)
		os.Exit(EXIT_INV_PRM)
	}

	storageNodeList := strings.Split(argMap[KEY_NODES], ",")

	targetName, err := iqnExtractTarget(argMap[KEY_IQN])
	if err != nil {
		fmt.Printf("Argument '%s': Invalid IQN format: Missing ':' separator and target name")
		os.Exit(EXIT_INV_PRM)
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
		fmt.Printf("LINSTOR volume operation failed, error: %s\n", err.Error())
		os.Exit(EXIT_FAILED_ACTION)
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
		argMap[KEY_TID],
	)

	return err
}

func actionDelete() error {
	if explainParams(ACTION_DELETE, DELETE_PARAMS) {
		return nil
	}

	argMap := make(map[string]string)
	loadParams(DELETE_PARAMS, argMap)

	err := parseArguments(&argMap)
	if err != nil {
		fmt.Printf("Invalid command line: %s\n", err.Error())
		os.Exit(EXIT_INV_PRM)
	}

	lun, err := strconv.ParseUint(argMap[KEY_LUN], 10, 8)
	if err != nil {
		fmt.Printf("Argument '%s': Unparseable logical unit number", KEY_LUN)
		os.Exit(EXIT_INV_PRM)
	}

	targetName, err := iqnExtractTarget(argMap[KEY_IQN])
	if err != nil {
		fmt.Printf("Argument '%s': Invalid IQN format: Missing ':' separator and target name")
		os.Exit(EXIT_INV_PRM)
	}

	return crmcontrol.DeleteCrmLu(targetName, uint8(lun))
}

func actionList() error {
	return crmcontrol.ReadConfiguration()
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
