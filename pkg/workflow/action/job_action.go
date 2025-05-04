package workflow

import (
	log "github.com/sirupsen/logrus"
	"nadleeh/pkg/env"
	"nadleeh/pkg/workflow/model"
	"nadleeh/pkg/workflow/run_context"
)

type JobAction struct {
	job         workflow.Job
	stepActions []*StepAction
	result      *ActionResult
}

func (action *JobAction) Run(ctx *run_context.WorkflowRunContext, parent *env.NadEnv, actionCtx *ActionContext) *ActionResult {
	log.Infof("Run job: %s", action.job.Name)

	jobEnv, err := InterpretEnv(&ctx.JSCtx, parent, action.job.Env, actionCtx.GenerateEnv())
	if err != nil {
		log.Errorf("Failed to interpret job env %v", err)
		action.result = NewActionResult(err, 1, "")
		return action.result
	}

	actionCtx.JobResult = &WorkflowJobResult{jobAction: action}
	failed := false
	for _, stepAction := range action.stepActions {

		ret := stepAction.Run(ctx, jobEnv, actionCtx, failed)

		if ret.ReturnCode != 0 {
			log.Errorf("Run job %s failed due to step %s failed", action.job.Name, stepAction.step.Name)
			if ret.If == IfEvalErr || ret.ContinueOnErr == ContinueOnErrEvalErr {
				log.Errorf("Failed to eval if or continue-on-error for step %s, fail fast", stepAction.step.Name)
				action.result = ret
				return action.result
			}
			if ret.ContinueOnErr == ContinueOnErrMatched {
				log.Infof("Continue on error matched for step %s", stepAction.step.Name)
				continue
			} else if ret.ContinueOnErr != NoContinueOnError {
				log.Errorf("set failed flag to true due to step %s failed, and it has no continue-on-error set", stepAction.step.Name)
				failed = true
			}
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
