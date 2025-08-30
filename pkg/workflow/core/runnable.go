package core

import (
	"nadleeh/pkg/workflow/run_context"

	"github.com/zhaojunlucky/golib/pkg/env"
)

type RunnableResult struct {
	// Fields of the struct
	Err        error
	ReturnCode int
	Output     string
}

type RunnableContext struct {
	NeedOutput     bool
	Args           env.Env
	JobStatus      *RunnableStatus
	WorkflowStatus *RunnableStatus
}

func (r *RunnableContext) GenerateMap() map[string]any {
	return map[string]any{
		"args":     r.Args,
		"workflow": r.WorkflowStatus,
		"job":      r.JobStatus,
	}
}

type Runnable interface {
	Compilable
	Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *RunnableContext) *RunnableResult
	CanRun() bool
	PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error
}

func NewRunnableResult(err error) *RunnableResult {
	if err == nil {
		return &RunnableResult{}
	}
	return &RunnableResult{Err: err, ReturnCode: 1}
}

func NewRunnable(err error, code int, output string) *RunnableResult {
	return &RunnableResult{
		Err:        err,
		ReturnCode: code,
		Output:     output,
	}
}
