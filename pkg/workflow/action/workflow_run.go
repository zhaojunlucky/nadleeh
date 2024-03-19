package workflow

import (
	log "github.com/sirupsen/logrus"
	"nadleeh/pkg/env"
	"nadleeh/pkg/script"
	"nadleeh/pkg/shell"
	workflow "nadleeh/pkg/workflow/model"
	"os"
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
	workflowEnv := env.NewEnv(parent, action.workflow.Env)
	action.changeWorkingDir(workflowEnv)

	log.Infof("Run workflow: %s", action.workflow.Name)
	for _, jobAction := range action.jobActions {
		action.jobActionResults = append(action.jobActionResults, jobAction.Run(action.workflowRunCtx, workflowEnv))
		if action.jobActionResults[len(action.jobActionResults)-1].ReturnCode != 0 {
			action.workflowActionResult = action.jobActionResults[len(action.jobActionResults)-1]
			return action.workflowActionResult
		}
	}
	action.workflowActionResult = NewActionResult(nil, 0, "")
	return action.workflowActionResult
}

func (action WorkflowRunAction) changeWorkingDir(workflowEnv *env.NadEnv) {
	if len(action.workflow.WorkingDir) > 0 {
		log.Infof("change working dir to: %s", action.workflow.WorkingDir)
		workflowEnv.Set("PWD", action.workflow.WorkingDir)
		workflowEnv.Set("HOME", action.workflow.WorkingDir)
		fi, err := os.Stat(action.workflow.WorkingDir)
		if err != nil {
			log.Panic(err)
		}
		if !fi.IsDir() {
			log.Panicf("working directory must be a directory: %s", action.workflow.WorkingDir)
		}

	}
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
