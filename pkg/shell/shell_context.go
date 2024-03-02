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

func (sh *ShellContext) Run(env env.Env, shell string) (int, string, error) {
	newUUID := uuid.New()

	tmpShFile := path.Join(sh.TmpDir, fmt.Sprintf("%s.sh", newUUID))
	err := os.WriteFile(tmpShFile, []byte(shell), fs.ModePerm)
	if err != nil {
		return -1, "Failed to write shell file", err
	}
	cmd := exec.Command("/bin/sh", tmpShFile)

	for key, value := range env.GetAll() {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	aow := NewStdOutputWriter()
	cmd.Stdout = *aow
	cmd.Stderr = *aow
	err = cmd.Run()

	if err != nil {
		return 1, aow.String(), err
	}
	return 0, aow.String(), nil
}

func NewShellContext() ShellContext {
	return ShellContext{
		TmpDir: os.TempDir(),
	}
}
