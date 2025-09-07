package plugin

import (
	"errors"
	"fmt"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/plugin/githubaction"
	"nadleeh/pkg/workflow/plugin/js_plug"
	"nadleeh/pkg/workflow/plugin/minio"
	"nadleeh/pkg/workflow/plugin/telegram"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)
import "nadleeh/pkg/workflow/plugin/googledrive"

var SupportedPlugins = []string{"google-drive", "github-actions", "telegram", "minio"}

type Plugin interface {
	core.Compilable
	core.Runnable
	Resolve() error
	GetName() string
}

func NewPlugin(name string, pluginPath string, config map[string]string) (Plugin, error) {
	var version string
	i := strings.Index(name, "@")
	if i != -1 {
		version = name[i+1:]
		name = name[:i]
	}
	var plug Plugin
	if name == "google-drive" {
		plug = &googledrive.GoogleDrive{Version: version, Config: config}
	} else if name == "github-actions" {
		plug = &githubaction.GitHubAction{Version: version, Config: config}
	} else if name == "telegram" {
		plug = &telegram.Telegram{Version: version, Config: config}
	} else if name == "minio" {
		plug = &minio.Minio{Version: version, Config: config}
	} else if len(version) > 0 {
		if len(pluginPath) > 0 {
			pluginPath = os.ExpandEnv(pluginPath)
		}
		var err error
		pluginPath, err = filepath.Abs(pluginPath)
		if err != nil {
			log.Errorf("failed to resove abs path for pluginPath %s", pluginPath)
			return nil, fmt.Errorf("failed to resove abs path for pluginPath %s", pluginPath)
		}
		plug = &js_plug.JSPlug{Version: version, PluginPath: pluginPath, PluginName: name, Config: config}
	}
	if plug != nil {
		return plug, plug.Resolve()
	}
	return nil, errors.New(fmt.Sprintf("unknown plugin: %s", name))
}
