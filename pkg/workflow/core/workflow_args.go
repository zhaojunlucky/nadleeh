package core

import "nadleeh/internal/argument"

type WorkflowArgs struct {
	File        *string
	Provider    *string
	Check       *bool
	Usage       *bool
	PrivateFile *string
}

// NewWorkflowArgsFromRunArgs creates WorkflowArgs from cobra RunArgs
func NewWorkflowArgsFromRunArgs(args *argument.RunArgs) *WorkflowArgs {
	wa := &WorkflowArgs{}

	if args.File != "" {
		wa.File = &args.File
	}

	if args.Provider != "" {
		wa.Provider = &args.Provider
	}

	wa.Check = &args.Check
	wa.Usage = &args.Usage

	if args.PrivateFile != "" {
		wa.PrivateFile = &args.PrivateFile
	}

	return wa
}
