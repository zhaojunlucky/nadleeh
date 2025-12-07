package workflow

import (
	"fmt"
	"nadleeh/pkg/util"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/plugin"
	"nadleeh/pkg/workflow/run_context"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
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
	PluginPath      string `yaml:"plugin-path"`

	runner core.Runnable
}

// Precheck validates the step definition
func (step *Step) Precheck() error {
	count := util.Bool2Int(len(step.Run) > 0) + util.Bool2Int(len(step.Script) > 0) + util.Bool2Int(len(step.Uses) > 0)
	if count > 1 {
		err := fmt.Errorf("multiple script/run/uses specified in step %s", step.Name)
		log.Error(err)
		return err
	}

	if step.HasScript() {
		step.runner = &JSRunner{Script: step.Script, Name: step.Name}
	} else if step.HasRun() {
		step.runner = &BashRunner{Script: step.Run, Name: step.Name}
	} else if step.RequirePlugin() {
		plug, err := plugin.NewPlugin(step.Uses, step.PluginPath, step.With)
		if err != nil {
			log.Errorf("failed to create plugin %s for step %s", step.Uses, step.Name)
			return err
		}
		step.runner = &PluginRunner{plug: plug, StepName: step.Name, Config: step.With}
	} else {
		return fmt.Errorf("no script/run/uses specified in step %s", step.Name)
	}
	return nil
}

// Compile compiles the workflow
func (step *Step) Compile(ctx run_context.WorkflowRunContext) error {
	return step.runner.Compile(ctx)
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

func (step *Step) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {

	stepStatus := core.NewRunnableStatus(step.Name, "step")
	ctx.JobStatus.AddChild(stepStatus)

	futureStatus := ctx.JobStatus.FutureStatus()

	var ifVal bool
	var err error
	if step.HasIf() {
		ifVal, err = step.evalIf(runCtx, parent, ctx)
		if err != nil {
			log.Errorf("failed to eval if for step %s", step.Name)
			stepStatus.Finish(err)
			return core.NewRunnable(err, -1, err.Error())
		} else if !ifVal {
			log.Errorf("step %s if evaluated as false, skip it", step.Name)
			stepStatus.Skipped()
			return core.NewRunnableResult(nil)
		}
	} else if futureStatus == core.Fail {
		log.Warnf("step %s skipped due to previous error", step.Name)
		stepStatus.Skipped()
		return core.NewRunnableResult(nil)
	}

	stepStatus.Start()
	log.Infof("start step %s", step.Name)

	stepEnv, err := InterpretWriteOnParentEnv(&runCtx.JSCtx, parent, step.Env, ctx.GenerateMap())
	if err != nil {
		log.Errorf("Failed to interpret job env %v", err)
		stepStatus.Finish(err)
		return core.NewRunnableResult(err)
	}

	result := step.runner.Do(stepEnv, runCtx, ctx)

	if result.ReturnCode != 0 {
		log.Errorf("step %s failed %v", step.Name, result.Err)
		stepStatus.Finish(result.Err)
		if step.HasContinueOnError() {
			log.Debugf("step %s failed, check continue on error", step.Name)
			value, err := step.evalContinueOnError(runCtx, stepEnv, ctx)
			if err != nil {
				stepStatus.Finish(err)

				return core.NewRunnable(err, -1, "")
			}
			stepStatus.ContinueOnErr = value
		}

	}
	return result

}

func (step *Step) evalContinueOnError(runCtx *run_context.WorkflowRunContext, parent env.Env, ctx *core.RunnableContext) (bool, error) {

	value, err := runCtx.JSCtx.EvalActionScriptBool(parent, step.ContinueOnError, ctx.GenerateMap())
	if err != nil {
		log.Errorf("Failed to eval continue-on-error for step %s, error: %v", step.Name, err)
		return false, err
	}
	log.Infof("Continue on error is %v for step %s", value, step.Name)
	return value, err
}

func (step *Step) evalIf(runCtx *run_context.WorkflowRunContext, parent env.Env, ctx *core.RunnableContext) (bool, error) {

	value, err := runCtx.JSCtx.EvalActionScriptBool(parent, step.If, ctx.GenerateMap())
	if err != nil {
		log.Errorf("Failed to eval if for step %s, error: %v", step.Name, err)
		return false, err
	}

	log.Infof("If is %v for step %s", value, step.Name)
	return value, nil
}

func (step *Step) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {
	return step.runner.PreflightCheck(parent, args, runCtx)
}
