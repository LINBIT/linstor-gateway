// Command line interface (CLI) functionality
package application

// cli module
//
// This module implements CLI functionality, such as argument parsing, list display, usage help, etc.

import "os"
import "fmt"
import "strings"
import "strconv"
import "errors"
import "github.com/LINBIT/linstor-remote-storage/crmcontrol"

// Application action command line parameters
const (
	ACTION_CREATE_CMD = "create"
	ACTION_DELETE_CMD = "delete"
	ACTION_LIST_CMD   = "list"
)

// Command line parameter keys for key=value parameters
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

type ActionDescription struct {
	Command string
	Params  []string
}

var ACTION_CREATE ActionDescription = ActionDescription{
	ACTION_CREATE_CMD,
	[]string{
		KEY_SVC_IP,
		KEY_IQN,
		KEY_LUN,
		KEY_USERNAME,
		KEY_PASSWORD,
		KEY_PORTALS,
		KEY_NODES,
		KEY_SIZE,
	},
}

var ACTION_DELETE ActionDescription = ActionDescription{
	ACTION_DELETE_CMD,
	[]string{
		KEY_IQN,
		KEY_LUN,
	},
}

var ACTION_LIST ActionDescription = ActionDescription{
	ACTION_LIST_CMD,
	[]string{},
}

// List of available program actions
var APPLICATION_ACTIONS []ActionDescription = []ActionDescription{
	ACTION_CREATE,
	ACTION_DELETE,
	ACTION_LIST,
}

// Parses the required arguments for resource creation from the command line,
// then calls the application.CreateResource(...) high-level API
func CliCreateResource() (int, error) {
	if explainParams(ACTION_CREATE) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	loadParams(ACTION_CREATE, argMap)

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

// Parses the required arguments for resource deletion from the command line,
// then calls the application.DeleteResource(...) high-level API
func CliDeleteResource() (int, error) {
	if explainParams(ACTION_DELETE) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	loadParams(ACTION_DELETE, argMap)

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

// Lists existing CRM resources, output goes to stdout
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

func CliCommands(programName string) int {
	fmt.Print("Syntax:\n")
	for _, action := range APPLICATION_ACTIONS {
		IndentPrintf(1, "%s %s", programName, action.Command)
		if len(action.Params) > 0 {
			fmt.Print(" [ parameters ]")
		}
		fmt.Print("\n")
	}
	fmt.Print("\n")
	fmt.Print(
		"To display a list of parameters for actions that require parameters,\n" +
			"enter that action without specifying any further parameters.\n\n",
	)
	return EXIT_SUCCESS
}

// Prints text with an indent, 4 spaces per indent level
func IndentPrint(indent int, text string) {
	for ctr := 0; ctr < indent; ctr++ {
		fmt.Print("    ")
	}
	fmt.Print(text)
}

// Prints a formatted text with an indent, 4 spaces per indent level
func IndentPrintf(indent int, format string, arguments ...interface{}) {
	for ctr := 0; ctr < indent; ctr++ {
		fmt.Print("    ")
	}
	fmt.Printf(format, arguments...)
}

// Parses all arguments that have a key in the supplied argMap
//
// Errors are generated for missing arguments, duplicate arguments,
// argument keys that do not have a corresponding key in the supplied
// argMap, and argument keys without values.
// An empty string is an allowed value.
//
// The parsed argument values are stored as value entries with their
// respective keys in the supplied argMap.
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

// Parses a logical unit number (LUN)
//
// A LUN, as parsed in this function, must be in the range [0, 255].
//
// Note that 0 is not a valid LUN for a volume (it is reserved for the SCSI controller),
// therefore, other parts of the application should check the validity of the LUN for
// the respective purpose.
func parseLun(lunStr string) (uint8, error) {
	lunNum, err := strconv.ParseUint(lunStr, 10, 8)
	if err != nil {
		return 0, errors.New("Argument '" + KEY_LUN + "': Unparseable logical unit number")
	}
	return uint8(lunNum), nil
}

// Splits key=value arguments
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

// Loads an array of strings into the supplied map as key entries
func loadParams(action ActionDescription, argMap map[string]string) {
	for _, key := range action.Params {
		argMap[key] = ""
	}
}

// Prints a simple usage description to stdout, indicating the application
// action to perform and the parameters required for it
func explainParams(action ActionDescription) bool {
	explainFlag := len(os.Args) < 3
	if explainFlag {
		if len(action.Params) > 0 {
			fmt.Printf("Parameters required for action %s:\n", action.Command)
			for _, entry := range action.Params {
				fmt.Printf("    %s\n", entry)
			}
			fmt.Printf("\n")
		} else {
			fmt.Printf("Action %s does not require any parameters", action.Command)
		}
	}
	return explainFlag
}
