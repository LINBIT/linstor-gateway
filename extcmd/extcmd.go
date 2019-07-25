// External command handling
package extcmd

// extcmd module
//
// This module manages piping input to external commands and piping
// stdout and stderr output back to this application.

import "strings"
import "os/exec"
import "io"
import "bufio"
import "sync"
import "errors"

type ExtCmdHandle struct {
	Command     *exec.Cmd
	stdinPipe   io.WriteCloser
	stdoutPipe  io.ReadCloser
	stderrPipe  io.ReadCloser
	stdoutLines []string
	stderrLines []string
	pipeThrGrp  sync.WaitGroup
	failedIo    bool
}

// Starts an external command and sets up anonymous pipes to/from that external command
func PipeToExtCmd(executable string, arguments []string) (*ExtCmdHandle, *bufio.Writer, error) {
	cmdObj := exec.Command(executable, arguments...)
	var handle *ExtCmdHandle = &ExtCmdHandle{cmdObj, nil, nil, nil, make([]string, 0), make([]string, 0), sync.WaitGroup{}, false}
	var err error
	handle.stdinPipe, handle.stdoutPipe, handle.stderrPipe, err = setupCmdPipes(cmdObj)
	if err != nil {
		return nil, nil, err
	}
	err = cmdObj.Start()
	if err != nil {
		return nil, nil, err
	}
	handle.pipeThrGrp.Add(2)
	go pipeCmdStream(handle.stdoutPipe, &handle.stdoutLines, &handle.failedIo, &handle.pipeThrGrp)
	go pipeCmdStream(handle.stderrPipe, &handle.stderrLines, &handle.failedIo, &handle.pipeThrGrp)
	bufStdinPipe := bufio.NewWriter(handle.stdinPipe)

	return handle, bufStdinPipe, nil
}

// Causes an error object indicating an I/O error to be returned upon completion of WaitForExtCmd()
func (handle *ExtCmdHandle) IoFailed() {
	handle.failedIo = true
}

// Waits for the external command to exit and returns collected stdout/stderr output from the external command
func (handle *ExtCmdHandle) WaitForExtCmd() ([]string, []string, error) {
	handle.stdinPipe.Close()
	handle.pipeThrGrp.Wait()
	err := handle.Command.Wait()
	if err != nil {
		return handle.stdoutLines, handle.stderrLines, err
	}
	if handle.failedIo {
		return handle.stdoutLines, handle.stderrLines, errors.New("CRM command: Interprocess communication failed: I/O error")
	}
	return handle.stdoutLines, handle.stderrLines, nil
}

// Fuses the elements of an array of strings to form a single string
func FuseStrings(linesArray []string) string {
	var dataBld strings.Builder
	for _, line := range linesArray {
		dataBld.WriteString(line)
	}
	return dataBld.String()
}

// Sets up stdin/stdout/stderr pipes for interprocess communication with the external process
func setupCmdPipes(cmdObj *exec.Cmd) (io.WriteCloser, io.ReadCloser, io.ReadCloser, error) {
	stdinPipe, err := cmdObj.StdinPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stdoutPipe, err := cmdObj.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stderrPipe, err := cmdObj.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	return stdinPipe, stdoutPipe, stderrPipe, nil
}

// Pipes data from the external process into an array of strings, one line per element
func pipeCmdStream(cmdStream io.ReadCloser, outputLines *[]string, failedIo *bool, pipeThrGrp *sync.WaitGroup) {
	cmdIn := bufio.NewReader(cmdStream)
	for {
		line, err := cmdIn.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				*failedIo = true
			}
			break
		}
		*outputLines = append(*outputLines, line)
	}
	cmdStream.Close()
	pipeThrGrp.Done()
}
