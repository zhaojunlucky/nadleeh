package plugin

import "nadleeh/pkg/env"
import "nadleeh/pkg/workflow/plugin/googledrive"

var SupportedPlugins = []string{"google-drive"}

type Plugin interface {
	Init(config map[string]string) error
	Run(parent env.Env) error
}

func NewPlugin(name string) Plugin {
	if name == "google-drive" {
		return &googledrive.GoogleDrive{}
	}
	return nil
}
