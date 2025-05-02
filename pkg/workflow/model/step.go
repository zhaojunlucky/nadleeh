package workflow

import (
	"fmt"
	"nadleeh/pkg/util"
	"nadleeh/pkg/workflow/plugin"
	"slices"
)

type Step struct {
	Name            string
	Id              string
	Script          string
	Env             map[string]string
	ContinueOnError string `yaml:"continue-on-error"`
	If              string
	Run             string
	Uses            string
	With            map[string]string
}

func (step *Step) Validate() error {
	count := util.Bool2Int(len(step.Run) > 0) + util.Bool2Int(len(step.Script) > 0) + util.Bool2Int(len(step.Uses) > 0)
	if count > 1 {
		return fmt.Errorf("multiple script/run/uses specified in step %s", step.Name)
	}

	if step.RequirePlugin() && !slices.Contains(plugin.SupportedPlugins, step.Uses) {
		return fmt.Errorf("unsupported plugin %s in step %s", step.Uses, step.Name)
	}
	return nil
}

func (step *Step) HasScript() bool {
	return len(step.Script) > 0
}

func (step *Step) HasRun() bool {
	return len(step.Run) > 0
}

func (step *Step) RequirePlugin() bool {
	return len(step.Uses) > 0
}

func (step *Step) HasIf() bool {
	return len(step.If) > 0
}

func (step *Step) HasContinueOnError() bool {
	return len(step.ContinueOnError) > 0
}
