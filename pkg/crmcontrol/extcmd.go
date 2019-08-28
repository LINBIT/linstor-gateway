package crmcontrol

import (
	"io"
	"io/ioutil"
	"os/exec"
	"sync"

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

	var stdoutSlurp []byte
	var stderrSlurp []byte
	ioWaitGroup := &sync.WaitGroup{}
	ioWaitGroup.Add(2)
	go func() {
		stdoutSlurp, _ = ioutil.ReadAll(stdout)
		ioWaitGroup.Done()
	}()
	go func() {
		stderrSlurp, _ = ioutil.ReadAll(stderr)
		ioWaitGroup.Done()
	}()
	ioWaitGroup.Wait()

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
