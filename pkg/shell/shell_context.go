package shell

import (
	"context"
	"fmt"
	"nadleeh/pkg/common"
	"nadleeh/pkg/file"
	"strings"
	"time"

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
	Timeout     time.Duration
}

func (sh *ShellContext) Compile(script string) error {
	script = strings.TrimSpace(script)
	bs := sh.scriptCache[script]
	if bs != nil {
		return bs.err
	}

	tmpShFile, err := sh.getShellTmpFile(script)
	if err != nil {
		return err
	}
	defer os.Remove(tmpShFile)
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

	timeout := sh.Timeout
	if timeout <= 0 {
		timeout = time.Hour
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/bash", "-e", tmpShFile)

	for key, value := range common.Sys.GetInfo().GetAll() {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	for key, value := range env.GetAll() {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	var output string
	if needOutput {
		aow := NewStdOutputWriter(NewOutputMasker(env.GetAll()))
		cmd.Stdout = aow
		cmd.Stderr = aow
		err = cmd.Run()
		output = aow.String()
	} else {
		aow := NewStdOutputWriter(NewOutputMasker(env.GetAll()))
		cmd.Stderr = aow
		cmd.Stdout = aow
		cmd.Stdin = os.Stdin
		err = cmd.Run()
	}

	if err != nil {
		_ = file.LogFileWithLineNo("bash", tmpShFile)
		if ctx.Err() == context.DeadlineExceeded {
			return 124, output, fmt.Errorf("bash script timed out after %s: %w", timeout, ctx.Err())
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), output, err
		}
		return 1, output, err
	}
	return 0, output, nil
}

func NewShellContext() ShellContext {
	return ShellContext{
		TmpDir:      os.TempDir(),
		scriptCache: make(map[string]*bashScript),
		Timeout:     time.Hour,
	}
}
