// LINSTOR - iSCSI administration utility
package main

import "os"
import "fmt"
import "errors"
import application "github.com/LINBIT/linstor-remote-storage/application"

func main() {
	fmt.Printf("\x1b[1;33m")
	fmt.Printf("linstor-remote experimental v0.1")
	fmt.Printf("\x1b[0m\n")

	programName := "linstor-remote"

	var exit_code int = application.EXIT_INV_PRM
	var err error = nil
	argCount := len(os.Args)
	if argCount == 1 {
		programName = os.Args[0]
	}

	if argCount >= 2 {
		action := os.Args[1]
		switch action {
		case application.ACTION_CREATE.Command:
			exit_code, err = application.CliCreateResource()
		case application.ACTION_DELETE.Command:
			exit_code, err = application.CliDeleteResource()
		case application.ACTION_LIST.Command:
			exit_code, err = application.CliListResources()
		default:
			err = errors.New("Action '" + action + "' is not implemented")
			exit_code = application.EXIT_FAILED_ACTION
		}
	} else {
		exit_code = application.CliCommands(programName)
	}

	if err != nil {
		fmt.Printf("%sOperation failed!%s Error: %s\n", application.COLOR_RED, application.COLOR_RESET, err.Error())
	}

	os.Exit(exit_code)
}
