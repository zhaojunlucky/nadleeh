package workflow

import (
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/plugin"
	"nadleeh/pkg/workflow/run_context"

	"github.com/zhaojunlucky/golib/pkg/env"
)

type PluginRunner struct {
	Config   map[string]string
	StepName string
	plug     plugin.Plugin
}

// Compile compiles the bash script
func (p *PluginRunner) Compile(runCtx run_context.WorkflowRunContext) error {
	return p.plug.Compile(runCtx)
}

// Do runs the bash script
func (p *PluginRunner) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	err := p.plug.Init(runCtx, p.Config)
	if err != nil {
		return core.NewRunnableResult(err)
	}
	return p.plug.Do(parent, runCtx, ctx)
}

func (p *PluginRunner) CanRun() bool {
	return !p.plug.CanRun()
}
