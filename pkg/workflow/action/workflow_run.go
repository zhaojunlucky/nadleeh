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
}

type WorkflowRunAction struct {
	workflow workflow.Workflow

	jobActions []*JobAction

	jobActionResults []*ActionResult

	workflowActionResult *ActionResult

	workflowRunCtx *WorkflowRunContext
}

func (action WorkflowRunAction) Run(parent env.Env) *ActionResult {
	workflowEnv := env.NewEnv(parent, &action.workflow.Env)
	for _, jobAction := range action.jobActions {
		action.jobActionResults = append(action.jobActionResults, jobAction.Run(action.workflowRunCtx, workflowEnv))
	}
	action.workflowActionResult = NewActionResult(nil, 0, "")
	return action.workflowActionResult
}

func NewWorkflowRunAction(workflow *workflow.Workflow) *WorkflowRunAction {
	wfa := &WorkflowRunAction{
		workflow:       *workflow,
		workflowRunCtx: NewWorkflowRunContext(),
	}

	for _, job := range workflow.Jobs {
		wfa.jobActions = append(wfa.jobActions, NewJobAction(job))
	}
	return wfa
}

func NewWorkflowRunContext() *WorkflowRunContext {
	return &WorkflowRunContext{
		JSCtx:    script.NewJSContext(),
		ShellCtx: shell.NewShellContext(),
	}
}
