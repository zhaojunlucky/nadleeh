package script

import (
	"errors"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type NSSHSession struct {
	session *ssh.Session
}

// shellEscape escapes a string for safe use in shell commands
// It wraps the string in single quotes and escapes any single quotes within
func shellEscape(s string) string {
	// If string contains no special characters, return as-is
	if !strings.ContainsAny(s, " \t\n\"'\\$`!*?[](){};<>|&") {
		return s
	}
	// Use single quotes and escape any single quotes in the string
	// Replace ' with '\'' (end quote, escaped quote, start quote)
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func (s *NSSHSession) Close() {

	if s.session != nil {
		err := s.session.Close()
		// Ignore EOF errors - they're normal when session is already closed
		if err != nil && !errors.Is(err, io.EOF) {
			log.Warnf("failed to closed session: %v", err)
		}
		s.session = nil
	}
}

func (s *NSSHSession) SetEnv(envs map[string]string) error {
	for k, v := range envs {
		err := s.session.Setenv(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *NSSHSession) RunCmd(name string, args *[]string, options map[string]any) (*CmdResult, error) {

	if options != nil {
		if envs, ok := options["envs"].(map[string]string); ok {
			err := s.SetEnv(envs)
			if err != nil {
				return nil, err
			}
		}
	}
	// Build command string
	cmdStr := name
	if args != nil && len(*args) > 0 {
		for _, arg := range *args {
			cmdStr += " " + shellEscape(arg)
		}
	}

	// Prepend working directory change if specified
	if options != nil {
		if workdir, ok := options["workingDir"].(string); ok && workdir != "" {
			cmdStr = "cd " + shellEscape(workdir) + " && " + cmdStr
		}
	}

	log.Debugf("running SSH command: %s", cmdStr)

	// Capture output
	output, err := s.session.CombinedOutput(cmdStr)

	ret := &CmdResult{
		Status: 0,
		Stdout: string(output),
		Stderr: "",
	}

	if err != nil {
		log.Errorf("SSH command failed: %v", err)
		// SSH session.Run returns exit status in the error
		var exitErr *ssh.ExitError
		if errors.As(err, &exitErr) {
			ret.Status = exitErr.ExitStatus()
		} else {
			ret.Status = 255
			ret.Stderr = err.Error()
		}
	}

	return ret, nil
}
