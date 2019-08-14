// Package extcmd is a simple wrapper around os/exec.
package extcmd

import (
	"io"
	"io/ioutil"
	"os/exec"
)

// Execute executes a command that optionally takes a string that is sent to the command's stdin
// The command returns stdout and stderr as strings.
func Execute(forStdin *string, name string, arg ...string) (string, string, error) {
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

	stdoutSlurp, _ := ioutil.ReadAll(stdout)
	stderrSlurp, _ := ioutil.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		return "", "", err
	}

	return string(stdoutSlurp), string(stderrSlurp), nil
}
