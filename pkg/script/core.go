package script

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"github.com/santhosh-tekuri/jsonschema/v5"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type NJSCore struct {
}

type CmdResult struct {
	Status int
	Stdout string
	Stderr string
}

func (n *NJSCore) ParseYaml(data string) (map[string]any, error) {
	var ret map[string]any
	err := yaml.Unmarshal([]byte(data), &ret)
	if err != nil {
		log.Errorf("failed to parse yaml: %v", err)
	}
	return ret, err
}

func (n *NJSCore) ValidateJSONSchema(data any, jsonSchema string) error {
	schema, err := jsonschema.CompileString("myschema.json", jsonSchema)
	if err != nil {
		log.Errorf("schema compilation error: %v", err)
		return fmt.Errorf("schema compilation error: %v", err)
	}
	if err = schema.Validate(data); err != nil {
		log.Errorf("validation failed: %v", err)
		return fmt.Errorf("validation failed: %v", err)
	} else {
		return nil
	}
}

func (n *NJSCore) RunCmd(name string, args []string) *CmdResult {
	log.Debugf("run cmd %s with args: %v", name, args)
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
