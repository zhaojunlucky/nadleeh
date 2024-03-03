package workflow

import (
	"fmt"
	"nadleeh/pkg/env"
	"nadleeh/pkg/workflow/plugin"
)
import "nadleeh/pkg/workflow/model"

type StepAction struct {
	step workflow.Step
}

func (action *StepAction) Run(ctx *WorkflowRunContext, parent env.Env) *ActionResult {
	fmt.Printf("Run step %s\n", action.step.Name)
	if action.step.RequirePlugin() {
		return action.runWithPlugin(ctx, parent)
	} else if action.step.HasRun() {
		return action.runWithShell(ctx, parent)
	} else if action.step.HasScript() {
		return action.runWithJS(ctx, parent)
	} else {
		panic(fmt.Sprintf("invalid step %s", action.step.Name))
	}
}

func (action *StepAction) runWithPlugin(ctx *WorkflowRunContext, parent env.Env) *ActionResult {
	plug := plugin.NewPlugin(action.step.Uses)
	if plug == nil {
		return NewActionResult(fmt.Errorf("invalid plugin %s", action.step.Uses), 0, "")
	}
	err := plug.Init(action.step.With)
	if err != nil {
		return NewActionResult(err, 1, "")
	}
	err = plug.Run(parent)
	if err != nil {
		return NewActionResult(err, 1, "")
	}
	return NewActionResult(nil, 0, "")
}

func (action *StepAction) runWithShell(ctx *WorkflowRunContext, parent env.Env) *ActionResult {
	ret, output, err := ctx.ShellCtx.Run(parent, action.step.Run)
	return NewActionResult(err, ret, output)
}

func (action *StepAction) runWithJS(ctx *WorkflowRunContext, parent env.Env) *ActionResult {
	ret, output, err := ctx.JSCtx.Run(parent, action.step.Script)
	return NewActionResult(err, ret, output)
}

func NewStepAction(step workflow.Step) *StepAction {
	return &StepAction{
		step: step,
	}
}
