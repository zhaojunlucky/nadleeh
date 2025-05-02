package workflow

import (
	"bufio"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"nadleeh/pkg/util"
	"os"
	"path/filepath"
	"strings"
)

type Workflow struct {
	Name       string
	Version    string
	Env        map[string]string
	Jobs       []Job
	WorkingDir string
}

type workflowDefinition struct {
	Name       string
	Version    string
	EnvFiles   []string `yaml:"env-files"`
	Env        map[string]string
	WorkingDir string `yaml:"working-dir"`
	Jobs       yaml.Node
}

func parseEnv(env map[string]string, envFiles []string) (map[string]string, error) {
	if env == nil {
		env = make(map[string]string)
	}
	for _, envFile := range envFiles {
		_, err := os.Stat(envFile)
		if os.IsNotExist(err) {
			if !strings.HasPrefix(envFile, "/") {
				envFile = filepath.Join(os.Getenv("HOME"), ".nadleeh", envFile)
				_, err = os.Stat(envFile)
				if os.IsNotExist(err) {
					log.Fatal(fmt.Sprintf("env file %s does not exist", envFile))
				}
			}
		}
		file, err := os.Open(envFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
				continue
			}
			kv := strings.SplitN(line, "=", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("env file line %s is not valid", line)
			}

			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			if util.HasKey(env, key) {
				log.Warnf("override env key %s from file %s", key, envFile)
			}
			env[key] = value
		}

	}

	return env, nil
}

func ParseWorkflow(filePath string) (*Workflow, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var rawWorkflow workflowDefinition
	if err := yaml.NewDecoder(file).Decode(&rawWorkflow); err != nil {
		return nil, err
	}
	wfEnv, err := parseEnv(rawWorkflow.Env, rawWorkflow.EnvFiles)
	if err != nil {
		return nil, err
	}
	workflow := &Workflow{
		Name:       rawWorkflow.Name,
		Version:    rawWorkflow.Version,
		Env:        wfEnv,
		WorkingDir: rawWorkflow.WorkingDir,
		Jobs:       []Job{},
	}

	for i, node := range rawWorkflow.Jobs.Content {
		if node.Tag != "!!str" {
			continue // Node is a map, so it is read out at key.
		}
		var job Job
		job.Name = node.Value

		err := rawWorkflow.Jobs.Content[i+1].Decode(&job)
		if err != nil {
			return nil, fmt.Errorf("failed to parse job %s: %w", job.Name, err)
		}

		workflow.Jobs = append(workflow.Jobs, job)
	}
	err = workflow.validate()
	if err != nil {
		return nil, err
	}
	return workflow, nil
}

func (w *Workflow) validate() error {
	var workflowErrs []error

	for _, job := range w.Jobs {
		workflowErrs = append(workflowErrs, job.Validate())
	}
	if len(workflowErrs) > 0 {
		return errors.Join(workflowErrs...)
	}
	return nil
}
