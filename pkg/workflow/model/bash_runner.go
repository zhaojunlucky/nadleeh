package workflow

import (
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
)

type BashRunner struct {
	Name     string
	Script   string
	hasError int
}

// Compile compiles the bash script
func (r *BashRunner) Compile(runCtx run_context.WorkflowRunContext) error {
	err := runCtx.ShellCtx.Compile(r.Script)
	if err != nil {
		r.hasError = 1
	} else {
		r.hasError = 2
	}
	return err
}

// Do runs the bash script
func (r *BashRunner) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	run, err := runCtx.JSCtx.EvalActionScriptStr(parent, r.Script, ctx.GenerateMap())
	if err != nil {
		log.Errorf("Failed to eval run for step %s", r.Name)
		return core.NewRunnableResult(err)
	}

	retCode, output, err := runCtx.ShellCtx.Run(parent, run, ctx.NeedOutput)
	return &core.RunnableResult{
		Err:        err,
		ReturnCode: retCode,
		Output:     output,
	}
}

func (r *BashRunner) CanRun() bool {
	return r.hasError > 1
}

func (r *BashRunner) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {
	return nil
}
