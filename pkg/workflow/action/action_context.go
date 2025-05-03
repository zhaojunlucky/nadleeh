package workflow

import (
	"nadleeh/pkg/env"
)

type ActionContext struct {
	Args           env.Env
	WorkflowResult *WorkflowResult
	JobResult      *WorkflowJobResult
}

func (a *ActionContext) GenerateEnv() map[string]interface{} {
	return map[string]interface{}{
		"arg":      a.Args.GetAll(), // arg are always a copy
		"job":      a.JobResult,
		"workflow": a.WorkflowResult,
	}
}
