package application

import "os"
import "fmt"
import "strings"
import "strconv"
import "errors"
import "github.com/LINBIT/linstor-remote-storage/crmcontrol"

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

func CliCreateResource() (int, error) {
	if explainParams(ACTION_CREATE, CREATE_PARAMS) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	loadParams(CREATE_PARAMS, argMap)

	err := parseArguments(&argMap)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid command line: " + err.Error())
	}

	lun, err := parseLun(argMap[KEY_LUN])
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_LUN + "': Unparseable logical unit number")
	}

	sizeKiB, err := strconv.ParseUint(argMap[KEY_SIZE], 10, 64)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_SIZE + "': Unparseable volume size")
	}

	storageNodeList := strings.Split(argMap[KEY_NODES], ",")

	return CreateResource(
		argMap[KEY_IQN],
		uint8(lun),
		sizeKiB,
		storageNodeList,
		// clientNodeList not supported yet
		make([]string, 0),
		argMap[KEY_SVC_IP],
		argMap[KEY_USERNAME],
		argMap[KEY_PASSWORD],
		argMap[KEY_PORTALS],
	)
}

func CliDeleteResource() (int, error) {
	if explainParams(ACTION_DELETE, DELETE_PARAMS) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	loadParams(DELETE_PARAMS, argMap)

	err := parseArguments(&argMap)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid command line: " + err.Error())
	}

	lun, err := parseLun(argMap[KEY_LUN])
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_LUN + "': Unparseable logical unit number")
	}

	return DeleteResource(argMap[KEY_IQN], lun)
}

func CliListResources() (int, error) {
	_, config, exit_code, err := ListResources()
	if err != nil {
		return exit_code, err
	}

	color(COLOR_YELLOW)
	fmt.Print("Cluster resources:")

	indent := 1
	color(COLOR_GREEN)
	IndentPrint(indent, "\x1b[1;32miSCSI resources:\x1b[0m\n")
	indent++
	IndentPrint(indent, "\x1b[1;32miSCSI targets:\x1b[0m\n")
	defaultColor()

	indent++
	if len(config.TargetList) > 0 {
		for _, rscName := range config.TargetList {
			IndentPrintf(indent, "%s\n", rscName)
		}
	} else {
		IndentPrint(indent, "No resources\n")
	}
	indent--

	color(COLOR_GREEN)
	IndentPrint(indent, "\x1b[1;32miSCSI logical units:\x1b[0m\n")
	defaultColor()

	indent++
	if len(config.LuList) > 0 {
		for _, rscName := range config.LuList {
			IndentPrintf(indent, "%s\n", rscName)
		}
	} else {
		IndentPrint(indent, "No resources\n")
	}
	indent -= 2

	color(COLOR_TEAL)
	IndentPrint(indent, "\x1b[1;32mOther cluster resources:\x1b[0m\n")
	defaultColor()

	indent++
	if len(config.OtherRscList) > 0 {
		for _, rscName := range config.OtherRscList {
			IndentPrintf(indent, "%s\n", rscName)
		}
	} else {
		IndentPrint(indent, "No resources\n")
	}
	indent = 0

	fmt.Print("\n")

	if config.TidSet.GetSize() > 0 {
		color(COLOR_GREEN)
		IndentPrint(indent, "\x1b[1;32mAllocated TIDs:\x1b[0m\n")
		defaultColor()

		indent++
		tidIter := config.TidSet.Iterator()
		for tid, isValid := tidIter.Next(); isValid; tid, isValid = tidIter.Next() {
			IndentPrintf(indent, "%d\n", tid)
		}
		indent--
	} else {
		color(COLOR_DARK_GREEN)
		IndentPrint(indent, "\x1b[1;32mNo TIDs allocated\x1b[0m\n")
		defaultColor()
	}
	fmt.Print("\n")

	freeTid, haveFreeTid := crmcontrol.GetFreeTargetId(config.TidSet.ToSortedArray())
	if haveFreeTid {
		color(COLOR_GREEN)
		IndentPrintf(indent, "\x1b[1;32mNext free TID:\x1b[0m\n    %d\n", int(freeTid))
	} else {
		color(COLOR_RED)
		IndentPrint(indent, "\x1b[1;31mNo free TIDs\x1b[0m\n")
	}
	defaultColor()
	fmt.Print("\n")

	return EXIT_SUCCESS, nil
}

func IndentPrint(indent int, text string) {
	for ctr := 0; ctr < indent; ctr++ {
		fmt.Print("    ")
	}
	fmt.Print(text)
}

func IndentPrintf(indent int, format string, arguments ...interface{}) {
	for ctr := 0; ctr < indent; ctr++ {
		fmt.Print("    ")
	}
	fmt.Printf(format, arguments...)
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

func parseLun(lunStr string) (uint8, error) {
	lunNum, err := strconv.ParseUint(lunStr, 10, 8)
	if err != nil {
		return 0, errors.New("Argument '" + KEY_LUN + "': Unparseable logical unit number")
	}
	return uint8(lunNum), nil
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
