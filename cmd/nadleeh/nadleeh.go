package main

import (
	"fmt"
	"nadleeh/pkg/env"
	workflow "nadleeh/pkg/workflow/action"
	workflowDef "nadleeh/pkg/workflow/model"
	"os"
	"path"
	"strings"
)

func main() {
	if len(os.Args) <= 1 {
		panic("usage: nadleeh worlflow.yml")
	}
	wYml := os.Args[1]
	ext := strings.ToLower(path.Ext(wYml))
	if ext != ".yaml" && ext != ".yml" {
		panic(fmt.Sprintf("%s must be a yaml file", wYml))
	}
	fi, err := os.Stat(wYml)
	if err != nil {
		panic(err)
	}
	if fi.IsDir() {
		panic(fmt.Sprintf("%s must be a file", wYml))
	}

	workflowDef, err := workflowDef.ParseWorkflow(wYml)
	if err != nil {
		panic(err)
	}

	wfa := workflow.NewWorkflowRunAction(workflowDef)
	wfa.Run(env.NewOSEnv())
}
