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

func (action *JobAction) Run(ctx *run_context.WorkflowRunContext, parent env.Env, workflowResult *WorkflowResult) *ActionResult {
	log.Infof("Run job: %s", action.job.Name)
	parent.SetAll(action.job.Env)
	jobResult := &WorkflowJobResult{jobAction: action}
	for _, stepAction := range action.stepActions {
		if stepAction.step.HasIf() {
			code, value, err := ctx.JSCtx.EvalBool(parent, stepAction.step.If, map[string]interface{}{
				"workflow": workflowResult,
				"job":      jobResult,
			})
			if err != nil {
				log.Errorf("Failed to eval if for job %s, step %s, code %d", action.job.Name, stepAction.step.Name, code)
				return NewActionResult(err, stepAction.result.ReturnCode, "")
			} else if !value {
				log.Infof("Skip step %s due to if condition", stepAction.step.Name)
				continue
			}
		}

		ret := stepAction.Run(ctx, parent)

		if ret.ReturnCode != 0 {
			if !stepAction.step.HasContinueOnError() {
				log.Errorf("Run job %s failed due to step %s failed", action.job.Name, stepAction.step.Name)
				return ret
			} else {
				code, value, err := ctx.JSCtx.EvalBool(parent, stepAction.step.ContinueOnError, map[string]interface{}{
					"workflow": workflowResult,
					"job":      jobResult,
				})
				if err != nil {
					log.Errorf("Failed to eval continue-on-error for job %s, step %s, code %d", action.job.Name,
						stepAction.step.Name, code)
					return NewActionResult(errors.Join(err, ret.Err), ret.ReturnCode, ret.Output)
				} else if value {
					log.Infof("Continue on error for job %s, step %s", action.job.Name, stepAction.step.Name)
				} else {
					return ret
				}
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
