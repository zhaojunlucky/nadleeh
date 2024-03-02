package workflow

import (
	"nadleeh/pkg/env"
	"nadleeh/pkg/script"
	"nadleeh/pkg/shell"
	workflow "nadleeh/pkg/workflow/model"
)

type WorkflowRunContext struct {
	JSCtx    script.JSContext
	ShellCtx shell.ShellContext
	TmpDir   string
}

type WorkflowRunAction struct {
	workflow workflow.Workflow

	jobActions []JobAction

	jobActionResults []*ActionResult

	workflowActionResult *ActionResult

	workflowRunCtx *WorkflowRunContext
}

func (action WorkflowRunAction) Run(ctx *WorkflowRunContext, parent env.Env) *ActionResult {
	workflowEnv := env.NewEnv(parent, &action.workflow.Env)
	for _, jobAction := range action.jobActions {
		action.jobActionResults = append(action.jobActionResults, jobAction.Run(ctx, workflowEnv))
	}
	action.workflowActionResult = NewActionResult(nil, 0, "")
	return action.workflowActionResult
}
