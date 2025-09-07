package runner

import (
	"nadleeh/pkg/common"
	"nadleeh/pkg/workflow/core"
	workflow "nadleeh/pkg/workflow/model"
	"nadleeh/pkg/workflow/run_context"
	"os"
	"path/filepath"
	"strings"

	"github.com/akamensky/argparse"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
	"gopkg.in/yaml.v3"
)

func RunWorkflow(wa *core.WorkflowArgs, argEnv env.Env) {
	if wa.File == nil || len(*wa.File) == 0 {
		log.Fatalf("invalid workflow file")
	}
	yml := *wa.File
	var err error
	if !strings.HasPrefix(yml, "@") {
		yml, err = filepath.Abs(*wa.File)
		if err != nil {
			log.Fatalf("failed to get absolute path of workflow file: %v", err)
		}
	}

	log.Infof("load workflow file %s", yml)

	common.MustSetEnvs(map[string]string{
		"WORKFLOW_FILE":       yml,
		"WORKFLOW_DIR":        filepath.Dir(yml),
		"WORKFLOW_VERSION":    common.Version,
		"WORKFLOW_BUILD_DATE": common.BuildDate,
	})

	ymlFile, err := workflow.LoadWorkflowFile(yml, wa)
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("parse workflow file %s", yml)
	wf, err := workflow.ParseWorkflow(ymlFile)
	if err != nil {
		log.Fatalf("failed to parse workflow %v", err)
	}

	log.Infof("precheck workflow")
	if err = wf.Precheck(); err != nil {
		log.Fatalf("failed to precheck workflow: %v", err)
	}

	runCtx := run_context.NewWorkflowRunContext(wa.PrivateFile)

	log.Infof("preflight workflow")
	if err = wf.PreflightCheck(env.NewOSEnv(), argEnv, runCtx); err != nil {
		log.Fatalf("failed to PreflightCheck workflow: %v", err)
	}

	if wa.Check == nil || !*wa.Check {
		log.Infof("run workflow file: %s", yml)
		result := wf.Do(env.NewOSEnv(), runCtx, &core.RunnableContext{
			NeedOutput: false,
			Args:       argEnv,
		})
		if result.ReturnCode != 0 {
			log.Fatalf("run workflow failed, code %d, err %v", result.ReturnCode, result.Err)
		} else {
			log.Info("run workflow passed")
		}
	} else {
		log.Infof("workflow check completed")
	}
}

func RunWorkflowConfig(argsMap map[string]argparse.Arg, args env.Env) {
	allArgs := args.GetAll()
	cfgFileArg := argsMap["_positionalArg_wf_1"]

	if !cfgFileArg.GetParsed() {
		log.Fatalf("invalid workflow config file arg.")
	}
	cfgFile := cfgFileArg.GetResult().(*string)
	log.Infof("Loading workflow config file %s", *cfgFile)
	file, err := os.Open(*cfgFile)
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
