package workflow

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
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
	Env        map[string]string
	WorkingDir string `yaml:"working-dir"`
	Jobs       yaml.Node
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
	workflow := &Workflow{
		Name:       rawWorkflow.Name,
		Version:    rawWorkflow.Version,
		Env:        rawWorkflow.Env,
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
