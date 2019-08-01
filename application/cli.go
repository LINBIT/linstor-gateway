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
import term "github.com/LINBIT/linstor-remote-storage/termcontrol"

// Application action command line parameters
const (
	ACTION_CREATE_CMD = "create"
	ACTION_DELETE_CMD = "delete"
	ACTION_START_CMD  = "start"
	ACTION_STOP_CMD   = "stop"
	ACTION_PROBE_CMD  = "probe"
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

// Default port for an iSCSI portal
const DFLT_ISCSI_PORTAL_PORT = 3260

type ParamDescription struct {
	Name     string
	Optional bool
}

type ActionDescription struct {
	Command string
	Params  []ParamDescription
}

type ParamValue struct {
	Optional bool
	Value    *string
}

var ACTION_CREATE ActionDescription = ActionDescription{
	ACTION_CREATE_CMD,
	[]ParamDescription{
		ParamDescription{Name: KEY_SVC_IP},
		ParamDescription{Name: KEY_IQN},
		ParamDescription{Name: KEY_LUN},
		ParamDescription{Name: KEY_USERNAME},
		ParamDescription{Name: KEY_PASSWORD},
		ParamDescription{Name: KEY_PORTALS, Optional: true},
		ParamDescription{Name: KEY_NODES},
		ParamDescription{Name: KEY_SIZE},
	},
}

var ACTION_START ActionDescription = ActionDescription{
	ACTION_START_CMD,
	[]ParamDescription{
		ParamDescription{Name: KEY_IQN},
		ParamDescription{Name: KEY_LUN},
	},
}

var ACTION_STOP ActionDescription = ActionDescription{
	ACTION_STOP_CMD,
	[]ParamDescription{
		ParamDescription{Name: KEY_IQN},
		ParamDescription{Name: KEY_LUN},
	},
}

var ACTION_DELETE ActionDescription = ActionDescription{
	ACTION_DELETE_CMD,
	[]ParamDescription{
		ParamDescription{Name: KEY_IQN},
		ParamDescription{Name: KEY_LUN},
	},
}

var ACTION_PROBE ActionDescription = ActionDescription{
	ACTION_PROBE_CMD,
	[]ParamDescription{
		ParamDescription{Name: KEY_IQN},
		ParamDescription{Name: KEY_LUN},
	},
}

var ACTION_LIST ActionDescription = ActionDescription{
	ACTION_LIST_CMD,
	[]ParamDescription{},
}

// List of available program actions
var APPLICATION_ACTIONS []ActionDescription = []ActionDescription{
	ACTION_CREATE,
	ACTION_START,
	ACTION_STOP,
	ACTION_DELETE,
	ACTION_PROBE,
	ACTION_LIST,
}

// Parses the required arguments for resource creation from the command line,
// then calls the application.CreateResource(...) high-level API
func CliCreateResource() (int, error) {
	if explainParams(ACTION_CREATE) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	err := parseArguments(ACTION_CREATE, &argMap)
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

	portals, found := argMap[KEY_PORTALS]
	if !found {
		portals = argMap[KEY_SVC_IP] + ":" + strconv.Itoa(DFLT_ISCSI_PORTAL_PORT)
	}

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
		portals,
	)
}

// Parses the required arguments for resource deletion from the command line,
// then calls the application.DeleteResource(...) high-level API
func CliDeleteResource() (int, error) {
	if explainParams(ACTION_DELETE) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	err := parseArguments(ACTION_DELETE, &argMap)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid command line: " + err.Error())
	}

	lun, err := parseLun(argMap[KEY_LUN])
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_LUN + "': Unparseable logical unit number")
	}

	return DeleteResource(argMap[KEY_IQN], lun)
}

// Parses the required arguments for starting resources from the command line,
// then calls the application.StartResource(...) high-level API
func CliStartResource() (int, error) {
	if explainParams(ACTION_START) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	err := parseArguments(ACTION_START, &argMap)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid command line: " + err.Error())
	}

	lun, err := parseLun(argMap[KEY_LUN])
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_LUN + "': Unparseable logical unit number")
	}

	return StartResource(argMap[KEY_IQN], lun)
}

// Parses the required arguments for stopping resources from the command line,
// then calls the application.StopResource(...) high-level API
func CliStopResource() (int, error) {
	if explainParams(ACTION_STOP) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	err := parseArguments(ACTION_STOP, &argMap)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid command line: " + err.Error())
	}

	lun, err := parseLun(argMap[KEY_LUN])
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_LUN + "': Unparseable logical unit number")
	}

	return StopResource(argMap[KEY_IQN], lun)
}

// Shows the run status of the specified resource
func CliProbeResource() (int, error) {
	if explainParams(ACTION_PROBE) {
		return EXIT_SUCCESS, nil
	}

	argMap := make(map[string]string)
	err := parseArguments(ACTION_PROBE, &argMap)
	if err != nil {
		return EXIT_INV_PRM, errors.New("Invalid command line: " + err.Error())
	}

	lun, err := parseLun(argMap[KEY_LUN])
	if err != nil {
		return EXIT_INV_PRM, errors.New("Argument '" + KEY_LUN + "': Unparseable logical unit number")
	}

	rscStateMap, _, err := ProbeResource(argMap[KEY_IQN], lun)
	if err != nil {
		return EXIT_FAILED_ACTION, err
	}

	fmt.Printf("Current state of CRM resources\niSCSI resource %s, logical unit #%d:\n", argMap[KEY_IQN], lun)
	for rscName, runState := range *rscStateMap {
		label := term.COLOR_YELLOW + "Unknown" + term.COLOR_RESET
		if runState.HaveState {
			if runState.Running {
				label = term.COLOR_GREEN + "Running" + term.COLOR_RESET
			} else {
				label = term.COLOR_RED + "Stopped" + term.COLOR_RESET
			}
		}
		fmt.Printf("    %-40s %s\n", rscName, label)
	}

	return EXIT_SUCCESS, nil
}

// Lists existing CRM resources, output goes to stdout
func CliListResources() (int, error) {
	_, config, exit_code, err := ListResources()
	if err != nil {
		return exit_code, err
	}

	term.Color(term.COLOR_YELLOW)
	fmt.Print("Cluster resources:")

	indent := 1
	term.Color(term.COLOR_GREEN)
	IndentPrint(indent, "\x1b[1;32miSCSI resources:\x1b[0m\n")
	indent++
	IndentPrint(indent, "\x1b[1;32miSCSI targets:\x1b[0m\n")
	term.DefaultColor()

	indent++
	if len(config.TargetList) > 0 {
		for _, rscName := range config.TargetList {
			IndentPrintf(indent, "%s\n", rscName)
		}
	} else {
		IndentPrint(indent, "No resources\n")
	}
	indent--

	term.Color(term.COLOR_GREEN)
	IndentPrint(indent, "\x1b[1;32miSCSI logical units:\x1b[0m\n")
	term.DefaultColor()

	indent++
	if len(config.LuList) > 0 {
		for _, rscName := range config.LuList {
			IndentPrintf(indent, "%s\n", rscName)
		}
	} else {
		IndentPrint(indent, "No resources\n")
	}
	indent -= 2

	term.Color(term.COLOR_TEAL)
	IndentPrint(indent, "\x1b[1;32mOther cluster resources:\x1b[0m\n")
	term.DefaultColor()

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
		term.Color(term.COLOR_GREEN)
		IndentPrint(indent, "\x1b[1;32mAllocated TIDs:\x1b[0m\n")
		term.DefaultColor()

		indent++
		tidIter := config.TidSet.Iterator()
		for tid, isValid := tidIter.Next(); isValid; tid, isValid = tidIter.Next() {
			IndentPrintf(indent, "%d\n", tid)
		}
		indent--
	} else {
		term.Color(term.COLOR_DARK_GREEN)
		IndentPrint(indent, "\x1b[1;32mNo TIDs allocated\x1b[0m\n")
		term.DefaultColor()
	}
	fmt.Print("\n")

	freeTid, haveFreeTid := crmcontrol.GetFreeTargetId(config.TidSet.ToSortedArray())
	if haveFreeTid {
		term.Color(term.COLOR_GREEN)
		IndentPrintf(indent, "\x1b[1;32mNext free TID:\x1b[0m\n    %d\n", int(freeTid))
	} else {
		term.Color(term.COLOR_RED)
		IndentPrint(indent, "\x1b[1;31mNo free TIDs\x1b[0m\n")
	}
	term.DefaultColor()
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
func parseArguments(action ActionDescription, argMap *map[string]string) error {
	collectedArgs := make(map[string]ParamValue)
	for _, paramDescr := range action.Params {
		collectedArgs[paramDescr.Name] = ParamValue{Optional: paramDescr.Optional, Value: nil}
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
		paramEntry, found := collectedArgs[key]
		if !found {
			return errors.New("Invalid argument key '" + key + "'")
		}
		if paramEntry.Value != nil {
			return errors.New("Duplicate argument '" + key + "'")
		}
		paramEntry.Value = &value
		collectedArgs[key] = paramEntry
	}

	for paramName, paramEntry := range collectedArgs {
		if paramEntry.Value != nil {
			(*argMap)[paramName] = *paramEntry.Value
		} else if !paramEntry.Optional {
			return errors.New("Missing argument '" + paramName + "'")
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

// Prints a simple usage description to stdout, indicating the application
// action to perform and the parameters required for it
func explainParams(action ActionDescription) bool {
	explainFlag := len(os.Args) < 3
	if explainFlag {
		if len(action.Params) > 0 {
			fmt.Printf("Parameters required for action %s:\n", action.Command)
			for _, entry := range action.Params {
				fmt.Printf("    %s", entry.Name)
				if entry.Optional {
					fmt.Print(" [optional]")
				}
				fmt.Print("\n")
			}
			fmt.Printf("\n")
		} else {
			fmt.Printf("Action %s does not require any parameters", action.Command)
		}
	}
	return explainFlag
}
