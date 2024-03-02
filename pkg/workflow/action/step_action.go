package workflow

import (
	"fmt"
	"nadleeh/pkg/env"
)
import "nadleeh/pkg/workflow/model"

type StepAction struct {
	step workflow.Step
}

func (action *StepAction) Run(ctx *WorkflowRunContext, parent env.Env) *ActionResult {
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
	return nil
}

func (action *StepAction) runWithShell(ctx *WorkflowRunContext, parent env.Env) *ActionResult {
	ret, output, err := ctx.ShellCtx.Run(parent, action.step.Run)
	return NewActionResult(err, ret, output)
}

func (action *StepAction) runWithJS(ctx *WorkflowRunContext, parent env.Env) *ActionResult {
	ret, output, err := ctx.JSCtx.Run(parent, action.step.Script)
	return NewActionResult(err, ret, output)
}
