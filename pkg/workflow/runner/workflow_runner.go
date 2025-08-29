package runner

import (
	"nadleeh/internal/argument"
	"nadleeh/pkg/workflow/core"
	workflow "nadleeh/pkg/workflow/model"
	"nadleeh/pkg/workflow/run_context"
	"os"
	"path"
	"strings"

	"github.com/akamensky/argparse"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
)

var (
	Run       = "run"
	Preflight = "preflight"
)

type WorkflowRunner struct {
}

func RunWorkflow(args map[string]argparse.Arg, argEnv env.Env) {
	yml, err := argument.GetStringFromArg(args["file"], true)
	if err != nil {
		log.Fatalf("failed to get yaml file arg %v", err)
	}
	wYml := *yml
	log.Infof("run workflow file: %s", wYml)
	ext := strings.ToLower(path.Ext(wYml))
	if ext != ".yaml" && ext != ".yml" {
		log.Fatalf("%s must be a yaml file", wYml)
	}
	fi, err := os.Stat(wYml)
	if err != nil {
		log.Fatalf("failed to get yaml file %v", err)
	}
	if fi.IsDir() {
		log.Fatalf("%s must be a file", wYml)
	}
	wf, err := workflow.ParseWorkflow(wYml)
	if err != nil {
		log.Fatalf("failed to parse workflow %v", err)
	}

	log.Infof("precheck workflow")
	if err = wf.Precheck(); err != nil {
		log.Fatalf("failed to precheck workflow: %v", err)
	}
	pPriFile, err := argument.GetStringFromArg(args["private"], false)

	runCtx := run_context.NewWorkflowRunContext(pPriFile)

	log.Infof("preflight workflow")
	if err = wf.PreflightCheck(env.OSEnv, argEnv, runCtx); err != nil {
		log.Fatalf("failed to PreflightCheck workflow: %v", err)
	}

	checkArg := args["check"]

	if !*checkArg.GetResult().(*bool) {
		log.Infof("start run workflow")
		result := wf.Do(env.OSEnv, runCtx, &core.RunnableContext{
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
