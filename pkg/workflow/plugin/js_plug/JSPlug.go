package js_plug

import (
	"fmt"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
)

type metadata struct {
	workflowVersion string `yaml:"workflow_version"`
	version         string `yaml:"version"`
	name            string `yaml:"name"`
	description     string `yaml:"description"`
	platforms       []struct {
		os      string `yaml:"os"`
		arch    string `yaml:"arch"`
		id      string `yaml:"id"`
		version string `yaml:"version"`
	} `yaml:"platforms"`
}

type runtime struct {
	args []struct {
		name     string `yaml:"name"`
		pattern  string `yaml:"pattern"`
		required bool   `yaml:"required"`
	} `yaml:"args"`
}

type manifest struct {
	metadata metadata `yaml:"metadata"`
	runtime  runtime  `yaml:"runtime"`
}

type JSPlug struct {
	Version    string
	PluginPath string
	PluginName string
	manifest   *manifest
	Config     map[string]string
	pm         *PluginMetadata
}

func (j *JSPlug) Compile(runCtx run_context.WorkflowRunContext) error {
	//TODO implement me
	panic("implement me")
}

func (j *JSPlug) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	//TODO implement me
	panic("implement me")
}

func (j *JSPlug) CanRun() bool {
	//TODO implement me
	panic("implement me")
}

func (j *JSPlug) GetName() string {
	return fmt.Sprintf("%s-%s", j.PluginName, j.Version)
}

func (j *JSPlug) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {
	return nil
}

func (j *JSPlug) Resolve() error {
	var err error
	j.pm, err = PM.LoadPlugin(j.PluginName, j.Version, "", j.PluginPath)
	if err != nil {
		log.Errorf("failed to load plugin: %v", err)
		return err
	}
	return nil
}
