package runner

import (
	"nadleeh/internal/argument"
	"nadleeh/pkg/common"
	"nadleeh/pkg/file"
	"nadleeh/pkg/workflow/core"
	workflow "nadleeh/pkg/workflow/model"
	"nadleeh/pkg/workflow/run_context"
	"os"
	"os/user"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
	"gopkg.in/yaml.v3"
)

// detectSudoUser detects if running under sudo and returns the appropriate HOME and USER values
func detectSudoUser() (home string, username string, detected bool) {
	sudoHome := os.Getenv("SUDO_HOME")
	sudoUser := os.Getenv("SUDO_USER")
	currentHome := os.Getenv("HOME")

	// Check if SUDO_HOME is set
	if len(sudoHome) > 0 {
		log.Infof("sudo home detected, override HOME=%s to HOME=%s", currentHome, sudoHome)
		return sudoHome, sudoUser, true
	}

	// Check if SUDO_USER is set and lookup home directory
	if len(sudoUser) > 0 {
		if u, err := user.Lookup(sudoUser); err == nil {
			log.Infof("sudo user detected, override HOME=%s to HOME=%s", currentHome, u.HomeDir)
			return u.HomeDir, sudoUser, true
		} else {
			log.Warnf("sudo user %s detected but failed to lookup home directory: %v", sudoUser, err)
		}
	}

	return "", "", false
}

func RunWorkflow(wa *core.WorkflowArgs, argEnv env.Env) {
	if wa.File == nil || len(*wa.File) == 0 {
		log.Fatalf("invalid workflow file")
	}
	yml := *wa.File
	val, err := file.FileExists(yml)
	if err != nil {
		log.Fatalf("failed to check workflow file existence: %v", err)
	}

	if val {
		yml, err = filepath.Abs(*wa.File)
		if err != nil {
			log.Fatalf("failed to get absolute path of workflow file: %v", err)
		}
	}

	log.Debugf("load workflow file %s", yml)

	requiredEnvs := map[string]string{
		"WORKFLOW_FILE":       yml,
		"WORKFLOW_DIR":        filepath.Dir(yml),
		"WORKFLOW_VERSION":    common.Version,
		"WORKFLOW_BUILD_DATE": common.BuildDate,
	}

	// Detect sudo user and override HOME/USER if needed
	if home, username, detected := detectSudoUser(); detected {
		requiredEnvs["HOME"] = home
		requiredEnvs["USER"] = username
	}

	common.MustSetEnvs(requiredEnvs)

	ymlFile, err := workflow.LoadWorkflowFile(yml, wa)
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("parse workflow file %s", yml)
	wf, err := workflow.ParseWorkflow(ymlFile)
	if err != nil {
		log.Fatalf("failed to parse workflow %v", err)
	}

	if wa.Usage != nil && *wa.Usage {
		log.Infof("workflow usage")
		wf.Checks.Usage()

		return
	}

	log.Debugf("precheck workflow")
	if err = wf.Precheck(); err != nil {
		log.Fatalf("failed to precheck workflow: %v", err)
	}

	runCtx := run_context.NewWorkflowRunContext(wa.PrivateFile)

	log.Debugf("preflight workflow")
	if err = wf.PreflightCheck(env.NewOSEnv(), argEnv, runCtx); err != nil {
		log.Fatalf("failed to PreflightCheck workflow: %v", err)
	}

	if wa.Check != nil && *wa.Check {
		log.Infof("workflow check completed")
		return
	}

	log.Debugf("run workflow file: %s", yml)
	result := wf.Do(env.NewOSEnv(), runCtx, &core.RunnableContext{
		NeedOutput: false,
		Args:       argEnv,
	})
	if result.ReturnCode != 0 {
		log.Fatalf("run workflow failed, code %d, err %v", result.ReturnCode, result.Err)
	} else {
		log.Info("run workflow passed")
	}
}

func RunWorkflowConfig(wfArgs *argument.WorkflowArgs) {
	argEnv := argument.CreateArgsEnv(wfArgs.Args)
	allArgs := argEnv.GetAll()

	log.Infof("Loading workflow config file %s", wfArgs.ConfigFile)
	file, err := os.Open(wfArgs.ConfigFile)
	if err != nil {
		log.Fatalf("failed to get the workflow config file: %v", err)
	}
	var workflowCfg workflow.WorkflowConfig
	if err = yaml.NewDecoder(file).Decode(&workflowCfg); err != nil {
		log.Fatalf("invalid config file format: %v", err)
	}

	if len(workflowCfg.Workflow) == 0 {
		log.Fatal("workflow config file is invalid, workflow is required")
	}

	wa := &core.WorkflowArgs{
		File: &workflowCfg.Workflow,
	}
	if len(workflowCfg.Provider) > 0 {
		wa.Provider = &workflowCfg.Provider
	}
	if len(workflowCfg.Private) > 0 {
		wa.PrivateFile = &workflowCfg.Private
	}

	for k, v := range workflowCfg.Args {
		allArgs[k] = v
	}

	RunWorkflow(wa, env.NewReadEnv(env.NewEmptyRWEnv(), allArgs))
}
