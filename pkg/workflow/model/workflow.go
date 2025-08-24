package workflow

import (
	"errors"
	"fmt"
	"nadleeh/pkg/util"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
)

type Workflow struct {
	Name       string
	Version    string
	Env        map[string]string
	Jobs       []*Job
	WorkingDir string
	Checks     WorkflowCheck
}

type WorkflowArg struct {
	Name    string `yaml:"name"`
	Pattern string `yaml:"pattern"`
}

type WorkflowCheck struct {
	PrivateKey bool          `yaml:"private-key"`
	Args       []WorkflowArg `yaml:"args"`
	Envs       []WorkflowArg `yaml:"envs"`
}

// Precheck validates the workflow definition
func (w *Workflow) Precheck() error {
	var workflowErrs []error

	for _, job := range w.Jobs {
		workflowErrs = append(workflowErrs, job.Precheck())
	}
	if len(workflowErrs) > 0 {
		return errors.Join(workflowErrs...)
	}
	return nil
}

// Compile compiles the workflow
func (w *Workflow) Compile(ctx run_context.WorkflowRunContext) error {
	var errs []error

	for _, job := range w.Jobs {
		errs = append(errs, job.Compile(ctx))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// PreflightCheck validates the workflow arguments and environment variables
func (w *Workflow) PreflightCheck(parent env.Env, args env.Env,
	workflowRunCtx *run_context.WorkflowRunContext) error {
	var errs []error
	if w.Checks.PrivateKey && !workflowRunCtx.SecureCtx.HasPrivateKey() {
		errs = append(errs, fmt.Errorf("no private key"))
	}
	argErrs := w.preflightCheck(args, w.Checks.Args)
	errs = append(errs, argErrs...)
	envErrs := w.preflightCheck(env.NewReadEnv(parent, w.Env), w.Checks.Envs)
	errs = append(errs, envErrs...)
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (w *Workflow) preflightCheck(env env.Env, checks []WorkflowArg) []error {
	envMap := env.GetAll()
	var errs []error
	for _, check := range checks {
		if !util.HasKey(envMap, check.Name) {
			errs = append(errs, fmt.Errorf("env %s is required", check.Name))
			continue
		}
		if len(check.Pattern) > 0 {
			matched, err := regexp.MatchString(check.Pattern, envMap[check.Name])
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if !matched {
				errs = append(errs, fmt.Errorf("env %s does not match pattern %s", check.Name, check.Pattern))
			}

		}

	}
	return errs
}

func (w *Workflow) CanRun() bool {
	return true
}

func (w *Workflow) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {

	workflowStatus := core.NewRunnableStatus(w.Name, "workflow")
	workflowStatus.Start()
	ctx.WorkflowStatus = workflowStatus

	workflowEnv, err := InterpretEnv(&runCtx.JSCtx, parent, w.Env, map[string]interface{}{"arg": ctx.Args})
	if err != nil {
		log.Errorf("Failed to interpret env %v", err)
		workflowStatus.Finish(err)
		return core.NewRunnable(err, 1, "")
	}

	w.changeWorkingDir(workflowEnv)

	log.Infof("Run workflow: %s", w.Name)
	for _, job := range w.Jobs {

		ret := job.Do(workflowEnv, runCtx, ctx)
		if ret.ReturnCode != 0 {
			log.Errorf("Run workflow %s failed due to job %s failed", w.Name, job.Name)
			workflowStatus.Finish(ret.Err)
			return ret
		}
	}
	workflowStatus.Finish(nil)
	return core.NewRunnableResult(nil)
}

func (w *Workflow) changeWorkingDir(workflowEnv *env.ReadWriteEnv) {
	if len(w.WorkingDir) > 0 {
		log.Infof("change working dir to: %s", w.WorkingDir)
		workflowEnv.Set("PWD", w.WorkingDir)
		workflowEnv.Set("HOME", w.WorkingDir)
		fi, err := os.Stat(w.WorkingDir)
		if err != nil {
			log.Fatal(err)
		}
		if !fi.IsDir() {
			log.Fatalf("working directory must be a directory: %s", w.WorkingDir)
		}
		err = os.Chdir(w.WorkingDir)
		if err != nil {
			// Handle the error if the directory change fails
			log.Fatal(err)
		}
	}
}
