package main

import "os"
import "fmt"
import "errors"
import application "github.com/LINBIT/linstor-remote-storage/application"

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

	var exit_code int = application.EXIT_INV_PRM
	var err error = nil
	argCount := len(os.Args)
	if argCount >= 2 {
		action := os.Args[1]
		switch action {
		case application.ACTION_CREATE:
			exit_code, err = application.ActionCreate()
		case application.ACTION_DELETE:
			exit_code, err = application.ActionDelete()
		case application.ACTION_LIST:
			exit_code, err = application.ActionList()
		default:
			err = errors.New("Action '" + action + "' is not implemented")
			exit_code = application.EXIT_FAILED_ACTION
		}
	}

	if err != nil {
		fmt.Printf("Operation failed, error: %s\n", err.Error())
		os.Exit(application.EXIT_FAILED_ACTION)
	}

	os.Exit(exit_code)
}
