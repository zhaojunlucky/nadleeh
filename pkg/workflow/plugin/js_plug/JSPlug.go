package js_plug

import (
	"nadleeh/pkg/workflow/plugin"
	"nadleeh/pkg/workflow/run_context"

	log "github.com/sirupsen/logrus"
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
	ctx        *run_context.WorkflowRunContext
	manifest   *manifest
	config     map[string]string
	pm         *plugin.PluginMetadata
}

func (j *JSPlug) Init(ctx *run_context.WorkflowRunContext, config map[string]string) error {
	j.ctx = ctx
	j.config = config
	return nil
}

func (j *JSPlug) Resolve() error {
	var err error
	j.pm, err = plugin.PM.LoadPlugin(j.PluginName, j.Version, "", j.PluginPath)
	if err != nil {
		log.Errorf("failed to load plugin: %v", err)
		return err
	}
	return nil
}
