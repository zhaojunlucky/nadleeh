package workflow

import (
	"errors"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
)

type Job struct {
	Name  string
	Steps []*Step
	Env   map[string]string
}

// Precheck validates the job definition
func (job *Job) Precheck() error {
	var jobErrors []error

	for _, step := range job.Steps {
		err := step.Precheck()
		if err != nil {
			jobErrors = append(jobErrors, err)
		}
	}

	if len(jobErrors) > 0 {
		return errors.Join(jobErrors...)
	}
	return nil
}

func (job *Job) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {
	var errs []error

	for _, step := range job.Steps {
		err := step.PreflightCheck(parent, args, runCtx)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	} else {
		return nil
	}
}

// Compile compiles the workflow
func (job *Job) Compile(ctx run_context.WorkflowRunContext) error {
	var errs []error

	for _, step := range job.Steps {
		errs = append(errs, step.Compile(ctx))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (job *Job) HasSteps() bool {
	return len(job.Steps) > 0
}

func (job *Job) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	log.Infof("Run job: %s", job.Name)
	jobStatus := core.NewRunnableStatus(job.Name, "job")
	ctx.WorkflowStatus.AddChild(jobStatus)
	ctx.JobStatus = jobStatus
	jobStatus.Start()

	jobEnv, err := InterpretEnv(&runCtx.JSCtx, parent, job.Env, ctx.GenerateMap())
	if err != nil {
		log.Errorf("Failed to interpret job env %v", err)
		jobStatus.Finish(err)
		return core.NewRunnableResult(err)
	}

	var errResults []error
	for _, step := range job.Steps {
		ret := step.Do(jobEnv, runCtx, ctx)
		if ret.ReturnCode != 0 {
			log.Errorf("Run job %s failed due to step %s failed", job.Name, step.Name)
			errResults = append(errResults, ret.Err)
		}
	}
	if len(errResults) == 0 {
		jobStatus.Finish([]error{}...)
		return core.NewRunnableResult(nil)
	} else {
		jobStatus.Finish(errResults...)
		return core.NewRunnable(errors.Join(errResults...), 255, "")
	}

}
