package run_context

import (
	"nadleeh/pkg/encrypt"
	"nadleeh/pkg/script"
	"nadleeh/pkg/shell"
)

type WorkflowRunContext struct {
	JSCtx     script.JSContext
	ShellCtx  shell.ShellContext
	SecureCtx encrypt.SecureContext
}

func NewWorkflowRunContext(pPriFile *string) *WorkflowRunContext {
	secCtx := encrypt.NewSecureContext(pPriFile)
	return &WorkflowRunContext{
		JSCtx:     script.NewJSContext(&secCtx),
		ShellCtx:  shell.NewShellContext(),
		SecureCtx: secCtx,
	}
}
