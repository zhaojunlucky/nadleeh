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
	step   workflow.Step
	result *ActionResult
}

func (action *StepAction) Run(ctx *run_context.WorkflowRunContext, parent *env.NadEnv, actionCtx *ActionContext) *ActionResult {
	err := InterpretEnvSelf(&ctx.JSCtx, parent, action.step.Env, actionCtx.GenerateEnv())
	if err != nil {
		log.Errorf("Failed to interpret step env %v", err)
		action.result = NewActionResult(err, 1, "")
		return action.result
	}

	log.Infof("Run step %s", action.step.Name)
	if action.step.RequirePlugin() {
		action.result = action.runWithPlugin(ctx, parent, actionCtx)
	} else if action.step.HasRun() {
		action.result = action.runWithShell(ctx, parent, actionCtx)
	} else if action.step.HasScript() {
		action.result = action.runWithJS(ctx, parent, actionCtx)
	} else {
		panic(fmt.Sprintf("invalid step %s", action.step.Name))
	}
	if action.result.ReturnCode != 0 {
		log.Errorf("Run step %s return code: %d, error: %s", action.step.Name, action.result.ReturnCode, action.result.Err)
	}
	return action.result
}

func (action *StepAction) runWithPlugin(ctx *run_context.WorkflowRunContext, parent env.Env, actionCtx *ActionContext) *ActionResult {
	plug := plugin.NewPlugin(action.step.Uses)
	if plug == nil {
		return NewActionResult(fmt.Errorf("invalid plugin %s", action.step.Uses), 0, "")
	}
	err := plug.Init(ctx, action.step.With)
	if err != nil {
		return NewActionResult(err, 1, "")
	}
	err = plug.Run(parent, actionCtx.GenerateEnv())
	if err != nil {
		return NewActionResult(err, 1, "")
	}
	return NewActionResult(nil, 0, "")
}

func (action *StepAction) runWithShell(ctx *run_context.WorkflowRunContext, parent env.Env, actionCtx *ActionContext) *ActionResult {
	run, err := ctx.JSCtx.EvalActionScript(parent, action.step.Run, actionCtx.GenerateEnv())
	if err != nil {
		log.Errorf("Failed to eval run for step %s", action.step.Name)
		return NewActionResult(err, 1, "")
	}

	ret, output, err := ctx.ShellCtx.Run(parent, run, action.needOutput())
	return NewActionResult(err, ret, output)
}

func (action *StepAction) needOutput() bool {
	return false
}

func (action *StepAction) runWithJS(ctx *run_context.WorkflowRunContext, parent env.Env, actionCtx *ActionContext) *ActionResult {
	ret, output, err := ctx.JSCtx.Run(parent, action.step.Script, actionCtx.GenerateEnv())
	return NewActionResult(err, ret, output)
}

func NewStepAction(step workflow.Step) *StepAction {
	return &StepAction{
		step: step,
	}
}
