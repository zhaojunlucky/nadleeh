package workflow

import (
	"errors"
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
		if stepAction.step.HasIf() {
			value, err := ctx.JSCtx.EvalBool(jobEnv, stepAction.step.If, actionCtx.GenerateEnv())
			if err != nil {
				log.Errorf("Failed to eval if for job %s, step %s", action.job.Name, stepAction.step.Name)
				return NewActionResult(err, stepAction.result.ReturnCode, "")
			} else if !value {
				log.Infof("Skip step %s due to if condition", stepAction.step.Name)
				continue
			}
		} else if failed {
			log.Infof("Skip step %s due to previous step failed, and no if condition", stepAction.step.Name)
			continue
		}

		ret := stepAction.Run(ctx, jobEnv, actionCtx)

		if ret.ReturnCode != 0 {

			log.Errorf("Run job %s failed due to step %s failed", action.job.Name, stepAction.step.Name)

			if stepAction.step.HasContinueOnError() {
				value, err := ctx.JSCtx.EvalBool(jobEnv, stepAction.step.ContinueOnError, actionCtx.GenerateEnv())
				if err != nil {
					log.Errorf("Failed to eval continue-on-error for job %s, step %s", action.job.Name,
						stepAction.step.Name)
					return NewActionResult(errors.Join(err, ret.Err), ret.ReturnCode, ret.Output)
				} else if value {
					log.Infof("Continue on error for job %s, step %s", action.job.Name, stepAction.step.Name)
				} else {
					failed = true
				}
			} else {
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
