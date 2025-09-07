package js_plug

import (
	"fmt"
	workflow "nadleeh/pkg/common"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"
	"os"

	"github.com/dlclark/regexp2"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
	"gopkg.in/yaml.v3"
)

type metadata struct {
	WorkflowVersion string `yaml:"workflow_version"`
	Version         string `yaml:"version"`
	Name            string `yaml:"name"`
	Description     string `yaml:"description"`
	Platforms       []struct {
		OS      string `yaml:"os"`
		Arch    string `yaml:"arch"`
		ID      string `yaml:"id"`
		Version string `yaml:"version"`
	} `yaml:"platforms"`
}

type runtime struct {
	Args []struct {
		Name     string `yaml:"name"`
		Pattern  string `yaml:"pattern"`
		Required bool   `yaml:"required"`
	} `yaml:"args"`
}

type manifest struct {
	Metadata metadata `yaml:"metadata"`
	Runtime  runtime  `yaml:"runtime"`
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
	if err != nil {
		log.Errorf("failed to compile main js %s for plugin %s: %v", j.pm.MainFile, j.PluginName, err)
		j.hasError = 1
	} else {
		log.Debugf("plugin %s js compiled successfully", j.PluginName)
		j.hasError = 2
	}
	return err
}

func (j *JSPlug) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	var err error
	argMaps := ctx.GenerateMap()
	log.Infof("run plugin %s", j.PluginName)
	j.Config, err = run_context.InterpretPluginCfg(runCtx, parent, j.Config, argMaps)
	if err != nil {
		log.Errorf("failed to interpreset plugin %s env %v", j.PluginName, err)
		return core.NewRunnableResult(err)
	}
	j.Config["PLUGIN_PATH"] = j.PluginPath

	plugEnv := workflow.NewWriteOnParentEnv(parent, j.Config)
	ret, output, err := runCtx.JSCtx.RunFile(plugEnv, j.pm.MainFile, argMaps)
	if err != nil {
		log.Errorf("plugin %s failed %v", j.PluginName, err)
	}
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
		return err
	}
	err = yaml.NewDecoder(file).Decode(&j.manifest)
	if err != nil {
		log.Errorf("failed to parse manifest file %s", j.pm.ManifestFile)
		return err
	}
	plugArgs := j.manifest.Runtime.Args

	for _, pa := range plugArgs {
		if !args.Contains(pa.Name) {
			if pa.Required {
				return fmt.Errorf("required argument '%s' is not provided for plugin %s", pa.Name, j.PluginName)
			}
		} else if len(pa.Pattern) > 0 {
			reg, err := regexp2.Compile(pa.Pattern, regexp2.IgnoreCase)
			if err != nil {
				return err
			}
			v, err := reg.MatchString(args.Get(pa.Name))
			if err != nil {
				return err
			}
			if !v {
				return fmt.Errorf("argument '%s' value doesn't match pattern '%s' for plugin %s", pa.Name, pa.Pattern, j.PluginName)
			}
		}

	}

	return nil
}

func (j *JSPlug) Resolve() error {
	var err error
	log.Debugf("plugin path: %s", j.PluginPath)

	j.pm, err = PM.LoadPlugin(j.PluginName, j.Version, "", j.PluginPath)
	if err != nil {
		log.Errorf("failed to load plugin: %v", err)
		return err
	}
	return nil
}
