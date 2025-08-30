package script

import (
	"bytes"
	"errors"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type NJSCore struct {
}

type CmdResult struct {
	Status int
	Stdout string
	Stderr string
}

func (n *NJSCore) RunCmd(name string, args []string) *CmdResult {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run() // Runs the command and waits for it to complete

	ret := &CmdResult{
		Status: 0,
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		log.Error(err)
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			ret.Status = exitError.ExitCode()
		} else {
			ret.Status = 255
		}
	}
	return ret
}
