// +build linux

package crmcontrol

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

const inouterr = `#!/bin/sh

echo "stdout: hi"
>&2 echo "stderr: hi"

while read -r line; do
	echo "stdout: $line"
	>&2 echo "stderr: $line"
done

echo "stdout: bye"
>&2 echo "stderr: bye"

exit "$1"
`

const expo = `stdout: hi
stdout: input
stdout: bye
`

const expe = `stderr: hi
stderr: input
stderr: bye
`

func genTempFile() (string, error) {
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	fname := tmpFile.Name()

	if err := tmpFile.Chmod(0700); err != nil { // because it got created with 600
		return fname, err
	}
	return fname, tmpFile.Close()
}

func testExecInOutErrZero(ret int) error {
	fname, err := genTempFile()
	if fname != "" {
		defer os.Remove(fname)
	}
	if err != nil {
		return err
	}

	ioutil.WriteFile(fname, []byte(inouterr), 0700)

	i := "input\n"

	o, e, err := execute(&i, fname, strconv.Itoa(ret))

	if ret == 0 && err != nil {
		return fmt.Errorf("I did not expect an error with a return code of 0")
	}
	if ret != 0 && err == nil {
		return fmt.Errorf("I expected an error with a return code of != 0")
	}

	if o != expo {
		return fmt.Errorf("I expected stout: '%s', but got '%s'", expo, o)
	}
	if e != expe {
		return fmt.Errorf("I expected stderr: '%s', but got '%s'", expe, e)
	}

	return nil
}

func TestExecInOutErrZero(t *testing.T) {
	if err := testExecInOutErrZero(0); err != nil {
		t.Fatalf("Did not expect this error: %v", err)
	}
}

func TestExecInOutErrNonZero(t *testing.T) {
	if err := testExecInOutErrZero(23); err != nil {
		t.Fatalf("Did not expect this error: %v", err)
	}
}
