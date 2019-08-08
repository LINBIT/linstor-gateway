// Command line interface (CLI) functionality
package application

// cli module
//
// This module implements CLI functionality, such as argument parsing, list display, usage help, etc.
import "fmt"
import "strings"

import "errors"

// Default port for an iSCSI portal
const DFLT_ISCSI_PORTAL_PORT = 3260

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
