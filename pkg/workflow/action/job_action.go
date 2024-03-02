package workflow

import "nadleeh/pkg/env"
import "nadleeh/pkg/workflow/model"

type JobAction struct {
	job               workflow.Job
	stepActions       []*StepAction
	stepActionResults []*ActionResult
}

func (action JobAction) Run(ctx *WorkflowRunContext, parent env.Env) *ActionResult {

	for _, stepAction := range action.stepActions {
		action.stepActionResults = append(action.stepActionResults, stepAction.Run(ctx, parent))
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
