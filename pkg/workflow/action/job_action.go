package workflow

import (
	log "github.com/sirupsen/logrus"
	"nadleeh/pkg/env"
	"nadleeh/pkg/workflow/model"
	"nadleeh/pkg/workflow/run_context"
)

type JobAction struct {
	job               workflow.Job
	stepActions       []*StepAction
	stepActionResults []*ActionResult
}

func (action JobAction) Run(ctx *run_context.WorkflowRunContext, parent env.Env) *ActionResult {
	log.Infof("Run job: %s", action.job.Name)
	parent.SetAll(action.job.Env)
	for _, stepAction := range action.stepActions {
		action.stepActionResults = append(action.stepActionResults, stepAction.Run(ctx, parent))
		if action.stepActionResults[len(action.stepActionResults)-1].ReturnCode != 0 {
			return action.stepActionResults[len(action.stepActionResults)-1]
		}
	}
	return NewActionResult(nil, 0, "")
}

func NewJobAction(job workflow.Job) *JobAction {
	j := &JobAction{
		job: job,
	}
	for _, step := range job.Steps {
		j.stepActions = append(j.stepActions, NewStepAction(step))
	}
	return j
}
