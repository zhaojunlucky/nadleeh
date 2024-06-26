package workflow

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"nadleeh/pkg/env"
	"nadleeh/pkg/workflow/plugin"
	"nadleeh/pkg/workflow/run_context"
)
import "nadleeh/pkg/workflow/model"

type StepAction struct {
	step workflow.Step
}

func (action *StepAction) Run(ctx *run_context.WorkflowRunContext, parent env.Env) *ActionResult {
	parent.SetAll(action.step.Env)

	log.Infof("Run step %s", action.step.Name)
	var ret *ActionResult
	if action.step.RequirePlugin() {
		ret = action.runWithPlugin(ctx, parent)
	} else if action.step.HasRun() {
		ret = action.runWithShell(ctx, parent)
	} else if action.step.HasScript() {
		ret = action.runWithJS(ctx, parent)
	} else {
		panic(fmt.Sprintf("invalid step %s", action.step.Name))
	}
	if ret.ReturnCode != 0 {
		log.Errorf("Run step %s return code: %d, error: %s", action.step.Name, ret.ReturnCode, ret.Err)
	}
	return ret
}

func (action *StepAction) runWithPlugin(ctx *run_context.WorkflowRunContext, parent env.Env) *ActionResult {
	plug := plugin.NewPlugin(action.step.Uses)
	if plug == nil {
		return NewActionResult(fmt.Errorf("invalid plugin %s", action.step.Uses), 0, "")
	}
	err := plug.Init(ctx, action.step.With)
	if err != nil {
		return NewActionResult(err, 1, "")
	}
	err = plug.Run(parent)
	if err != nil {
		return NewActionResult(err, 1, "")
	}
	return NewActionResult(nil, 0, "")
}

func (action *StepAction) runWithShell(ctx *run_context.WorkflowRunContext, parent env.Env) *ActionResult {
	ret, output, err := ctx.ShellCtx.Run(parent, action.step.Run, action.needOutput())
	return NewActionResult(err, ret, output)
}

func (action *StepAction) needOutput() bool {
	return false
}

func (action *StepAction) runWithJS(ctx *run_context.WorkflowRunContext, parent env.Env) *ActionResult {
	ret, output, err := ctx.JSCtx.Run(parent, action.step.Script)
	return NewActionResult(err, ret, output)
}

func NewStepAction(step workflow.Step) *StepAction {
	return &StepAction{
		step: step,
	}
}
