package workflow

import "nadleeh/pkg/env"

type Action interface {
	Run(ctx *WorkflowRunContext, env env.Env) *ActionResult
}
