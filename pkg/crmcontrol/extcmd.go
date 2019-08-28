package crmcontrol

import (
	"errors"
	"io"
	"io/ioutil"
	"os/exec"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

// execute executes a command that optionally takes a string that is sent to the command's stdin
// The command returns stdout and stderr as strings.
func execute(forStdin *string, name string, arg ...string) (string, string, error) {
	cmd := exec.Command(name, arg...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}
	defer stderr.Close()

	if forStdin != nil {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return "", "", err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, *forStdin)
		}()
	}

	if err := cmd.Start(); err != nil {
		return "", "", err
	}

	ioFailed := uint32(0)
	var stdoutSlurp []byte
	var stderrSlurp []byte
	ioWaitGroup := &sync.WaitGroup{}
	ioWaitGroup.Add(2)
	go func(ioError *uint32) {
		var readErr error
		stdoutSlurp, readErr = ioutil.ReadAll(stdout)
		if readErr != nil {
			atomic.StoreUint32(ioError, uint32(1))
		}
		ioWaitGroup.Done()
	}(&ioFailed)
	go func(ioError *uint32) {
		var readErr error
		stderrSlurp, readErr = ioutil.ReadAll(stderr)
		if readErr != nil {
			atomic.StoreUint32(ioError, uint32(1))
		}
		ioWaitGroup.Done()
	}(&ioFailed)
	ioWaitGroup.Wait()

	// Ensure that the value change caused by the I/O threads is seen
	// in the current thread
	if atomic.LoadUint32(&ioFailed) != 0 {
		return "", "", errors.New("Command execution failed: I/O error while piping data")
	}

	if len(stdoutSlurp) >= 1 {
		log.Trace("CRM command stdout output:", string(stdoutSlurp))
	} else {
		log.Trace("No stdout output")
	}

	if len(stderrSlurp) >= 1 {
		log.Trace("CRM command stderr output:", string(stderrSlurp))
	} else {
		log.Trace("No stderr output")
	}

	return string(stdoutSlurp), string(stderrSlurp), cmd.Wait()
}
