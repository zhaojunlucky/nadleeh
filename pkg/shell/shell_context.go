package shell

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"

	"io/fs"
	"os"
	"os/exec"
	"path"
)

type bashScript struct {
	err error
}

type ShellContext struct {
	TmpDir      string
	scriptCache map[string]*bashScript
}

func (sh *ShellContext) Compile(script string) error {
	script = strings.TrimSpace(script)
	bs := sh.scriptCache[script]
	if bs != nil {
		return bs.err
	}

	tmpShFile, err := sh.getShellTmpFile(script)
	defer os.Remove(tmpShFile)
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", "-n", tmpShFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("compile shell error: %s", string(out))
		compileErr := fmt.Errorf("compile shell error: %s: %w", string(out), err)
		sh.scriptCache[script] = &bashScript{
			err: compileErr,
		}
		return compileErr
	}
	sh.scriptCache[script] = &bashScript{
		err: nil,
	}
	return nil
}

func (sh *ShellContext) getShellTmpFile(script string) (string, error) {
	newUUID := uuid.New()

	tmpShFile := path.Join(sh.TmpDir, fmt.Sprintf("%s.sh", newUUID))
	err := os.WriteFile(tmpShFile, []byte(script), fs.ModePerm)
	if err != nil {
		return "Failed to write shell file", err
	}
	return tmpShFile, nil
}

func (sh *ShellContext) Run(env env.Env, shell string, needOutput bool) (int, string, error) {

	tmpShFile, err := sh.getShellTmpFile(shell)
	if err != nil {
		return 1, "", err
	}

	defer os.Remove(tmpShFile)
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
		TmpDir:      os.TempDir(),
		scriptCache: make(map[string]*bashScript),
	}
}
