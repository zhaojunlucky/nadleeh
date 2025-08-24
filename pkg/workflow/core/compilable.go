package core

import "nadleeh/pkg/workflow/run_context"

type Compilable interface {
	Compile(runCtx run_context.WorkflowRunContext) error
}
