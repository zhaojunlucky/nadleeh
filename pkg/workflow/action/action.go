package workflow

import (
	"nadleeh/pkg/env"
	"nadleeh/pkg/workflow/run_context"
)

type Action interface {
	Run(ctx *run_context.WorkflowRunContext, env env.Env) *ActionResult
}
