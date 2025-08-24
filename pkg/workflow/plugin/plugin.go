package plugin

import (
	"errors"
	"fmt"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/plugin/githubaction"
	"nadleeh/pkg/workflow/plugin/js_plug"
	"nadleeh/pkg/workflow/plugin/minio"
	"nadleeh/pkg/workflow/plugin/telegram"
	"nadleeh/pkg/workflow/run_context"
	"strings"
)
import "nadleeh/pkg/workflow/plugin/googledrive"

var SupportedPlugins = []string{"google-drive", "github-actions", "telegram", "minio"}

type Plugin interface {
	core.Compilable
	core.Runnable
	Init(ctx *run_context.WorkflowRunContext, config map[string]string) error
	Resolve() error
	GetName() string
}

func NewPlugin(name string, pluginPath string) (Plugin, error) {
	var version string
	i := strings.Index(name, "@")
	if i != -1 {
		version = name[i+1:]
		name = name[:i]
	}
	var plug Plugin
	if name == "google-drive" {
		plug = &googledrive.GoogleDrive{Version: version}
	} else if name == "github-actions" {
		plug = &githubaction.GitHubAction{Version: version}
	} else if name == "telegram" {
		plug = &telegram.Telegram{Version: version}
	} else if name == "minio" {
		plug = &minio.Minio{Version: version}
	} else if len(version) > 0 {
		plug = &js_plug.JSPlug{Version: version, PluginPath: pluginPath, PluginName: name}
	}
	if plug != nil {
		return plug, plug.Resolve()
	}
	return nil, errors.New(fmt.Sprintf("unknown plugin: %s", name))
}
