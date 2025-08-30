package workflow

import (
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
)

type JSRunner struct {
	Name     string
	Script   string
	hasError int
}

func (r *JSRunner) Compile(runCtx run_context.WorkflowRunContext) error {
	err := runCtx.JSCtx.Compile(r.Script)
	log.Errorf("js compile error: %v", err)
	if err != nil {
		r.hasError = 1
	} else {
		r.hasError = 2
	}
	return err
}

func (r *JSRunner) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	retCode, output, err := runCtx.JSCtx.Run(parent, r.Script, ctx.GenerateMap())
	if err != nil {
		log.Errorf("failed to run js: %v", err)
	}
	return &core.RunnableResult{
		Err:        err,
		ReturnCode: retCode,
		Output:     output,
	}
}

func (r *JSRunner) CanRun() bool {
	return r.hasError > 1
}

func (r *JSRunner) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {
	return nil
}
