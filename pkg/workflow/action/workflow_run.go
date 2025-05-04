package workflow

import (
	"errors"
	"fmt"
	"github.com/akamensky/argparse"
	log "github.com/sirupsen/logrus"
	"nadleeh/internal/argument"
	"nadleeh/pkg/env"
	"nadleeh/pkg/util"
	workflow "nadleeh/pkg/workflow/model"
	"nadleeh/pkg/workflow/run_context"
	"os"
	"path"
	"regexp"
	"strings"
)

type WorkflowRunAction struct {
	workflow workflow.Workflow

	jobActions     []*JobAction
	result         *ActionResult
	workflowRunCtx *run_context.WorkflowRunContext
}

func (action *WorkflowRunAction) Run(parent env.Env, args env.Env) *ActionResult {
	workflowEnv, err := InterpretEnv(&action.workflowRunCtx.JSCtx, parent, action.workflow.Env, map[string]interface{}{"arg": args})
	if err != nil {
		log.Errorf("Failed to interpret env %v", err)
		return NewActionResult(err, 1, "")
	}

	action.changeWorkingDir(workflowEnv)

	log.Infof("Run workflow: %s", action.workflow.Name)
	for _, jobAction := range action.jobActions {
		actionCtx := &ActionContext{Args: args, WorkflowResult: &WorkflowResult{workflowRunAction: action}}
		jobEnv := env.NewEnv(workflowEnv, nil) // every job has its own env
		ret := jobAction.Run(action.workflowRunCtx, jobEnv, actionCtx)
		if ret.ReturnCode != 0 {
			action.result = ret
			log.Errorf("Run workflow %s failed due to job %s failed", action.workflow.Name, jobAction.job.Name)
			return action.result
		}
	}
	action.result = NewActionResult(nil, 0, "")
	return action.result
}

func (action *WorkflowRunAction) changeWorkingDir(workflowEnv *env.NadEnv) {
	if len(action.workflow.WorkingDir) > 0 {
		log.Infof("change working dir to: %s", action.workflow.WorkingDir)
		workflowEnv.Set("PWD", action.workflow.WorkingDir)
		workflowEnv.Set("HOME", action.workflow.WorkingDir)
		fi, err := os.Stat(action.workflow.WorkingDir)
		if err != nil {
			log.Panic(err)
		}
		if !fi.IsDir() {
			log.Fatalf("working directory must be a directory: %s", action.workflow.WorkingDir)
		}
		err = os.Chdir(action.workflow.WorkingDir)
		if err != nil {
			// Handle the error if the directory change fails
			log.Panic(err)
		}

	}
}

func (action *WorkflowRunAction) validate(env env.Env, checks []workflow.WorkflowArg) []error {
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

func (action *WorkflowRunAction) Validate(wfEnv env.Env, argEnv env.Env, checks *workflow.WorkflowCheck) error {
	var errs []error
	if checks.PrivateKey && !action.workflowRunCtx.SecureCtx.HasPrivateKey() {
		errs = append(errs, fmt.Errorf("no private key"))
	}
	argErrs := action.validate(argEnv, checks.Args)
	errs = append(errs, argErrs...)
	envErrs := action.validate(wfEnv, checks.Envs)
	errs = append(errs, envErrs...)
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil

}

func NewWorkflowRunAction(workflow *workflow.Workflow, pPriFile *string) *WorkflowRunAction {
	wfa := &WorkflowRunAction{
		workflow:       *workflow,
		workflowRunCtx: run_context.NewWorkflowRunContext(pPriFile),
	}

	for _, job := range workflow.Jobs {
		wfa.jobActions = append(wfa.jobActions, NewJobAction(job))
	}
	return wfa
}

func RunWorkflow(cmd *argparse.Command, args map[string]argparse.Arg, argEnv env.Env) {
	yml, err := argument.GetStringFromArg(args["file"], true)
	if err != nil {
		log.Fatalf("failed to get yaml file arg %v", err)
	}
	wfEnv := env.NewOSEnv()
	wYml := *yml
	log.Infof("run workflow file: %s", wYml)
	ext := strings.ToLower(path.Ext(wYml))
	if ext != ".yaml" && ext != ".yml" {
		log.Fatalf("%s must be a yaml file", wYml)
	}
	fi, err := os.Stat(wYml)
	if err != nil {
		log.Fatal("failed to get yaml file %v", err)
	}
	if fi.IsDir() {
		log.Fatalf("%s must be a file", wYml)
	}

	wfDef, checks, err := workflow.ParseWorkflow(wYml)
	if err != nil {
		log.Fatalf("failed to parse workflow %v", err)
	}
	pPriFile, err := argument.GetStringFromArg(args["private"], false)

	wfa := NewWorkflowRunAction(wfDef, pPriFile)
	err = wfa.Validate(wfEnv, argEnv, checks)
	if err != nil {
		log.Fatalf("failed to validate workflow %v", err)
	}
	result := wfa.Run(wfEnv, argEnv)
	log.Infof("run workflow end, status %d", result.ReturnCode)

	if result.ReturnCode != 0 {
		log.Fatalf("run workflow failed, code %d, err %v", result.ReturnCode, result.Err)
	}
}
