package shell

import (
	"fmt"
	"github.com/google/uuid"
	"io/fs"
	"nadleeh/pkg/env"
	"os"
	"os/exec"
	"path"
)

type ShellContext struct {
	TmpDir string
}

func (sh *ShellContext) Run(env env.Env, shell string, needOutput bool) (int, string, error) {
	newUUID := uuid.New()

	tmpShFile := path.Join(sh.TmpDir, fmt.Sprintf("%s.sh", newUUID))
	err := os.WriteFile(tmpShFile, []byte(shell), fs.ModePerm)
	if err != nil {
		return -1, "Failed to write shell file", err
	}
	cmd := exec.Command("/bin/bash", "-e", tmpShFile)

	for key, value := range env.GetAll() {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	var output string
	if needOutput {
		aow := NewStdOutputWriter()
		cmd.Stdout = *aow
		cmd.Stderr = *aow
		err = cmd.Run()
		output = aow.String()
	} else {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		err = cmd.Run()
	}

	if err != nil {
		return 1, output, err
	}
	return 0, output, nil
}

func NewShellContext() ShellContext {
	return ShellContext{
		TmpDir: os.TempDir(),
	}
}
