package js_plug

import (
	"fmt"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"
	"os"

	"github.com/dlclark/regexp2"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
	"gopkg.in/yaml.v3"
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
	manifest   manifest
	Config     map[string]string
	pm         *PluginMetadata
	hasError   int
}

func (j *JSPlug) Compile(runCtx run_context.WorkflowRunContext) error {
	_, err := runCtx.JSCtx.CompileFile(j.pm.MainFile)
	return err
}

func (j *JSPlug) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	var err error
	argMaps := ctx.GenerateMap()
	j.Config, err = run_context.InterpretPluginCfg(runCtx, parent, j.Config, argMaps)
	if err != nil {
		return core.NewRunnableResult(err)
	}

	plugEnv := env.NewReadEnv(parent, j.Config)
	ret, output, err := runCtx.JSCtx.RunFile(plugEnv, j.pm.MainFile, argMaps)
	return core.NewRunnable(err, ret, output)
}

func (j *JSPlug) CanRun() bool {
	return j.hasError > 1
}

func (j *JSPlug) GetName() string {
	return fmt.Sprintf("%s-%s", j.PluginName, j.Version)
}

func (j *JSPlug) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {
	file, err := os.Open(j.pm.ManifestFile)
	if err != nil {
		log.Errorf("failed to read manifest file %s", j.pm.ManifestFile)
	}
	err = yaml.NewDecoder(file).Decode(&j.manifest)
	if err != nil {
		log.Errorf("failed to parse manifest file %s", j.pm.ManifestFile)
	}
	plugArgs := j.manifest.runtime.args

	for _, pa := range plugArgs {
		if !args.Contains(pa.name) {
			if pa.required {
				return fmt.Errorf("required argument '%s' is not provided", pa.name)
			}
		} else if len(pa.pattern) > 0 {
			reg, err := regexp2.Compile(pa.pattern, regexp2.IgnoreCase)
			if err != nil {
				return err
			}
			v, err := reg.MatchString(args.Get(pa.name))
			if err != nil {
				return err
			}
			if !v {
				return fmt.Errorf("argument '%s' value doesn't match pattern '%s'", pa.name, pa.pattern)
			}
		}

	}

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
