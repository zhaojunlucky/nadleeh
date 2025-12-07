package core

import "github.com/akamensky/argparse"

type WorkflowArgs struct {
	File        *string
	Provider    *string
	Check       *bool
	PrivateFile *string
}

func NewWorkflowArgs(args map[string]argparse.Arg) *WorkflowArgs {
	wa := &WorkflowArgs{}
	file := args["file"]
	if file != nil && file.GetParsed() {
		wa.File = file.GetResult().(*string)
	}

	provider := args["provider"]
	if provider != nil && provider.GetParsed() {
		wa.Provider = provider.GetResult().(*string)
	}

	check := args["check"]
	if check != nil && check.GetParsed() {
		wa.Check = check.GetResult().(*bool)
	}

	return wa
}
