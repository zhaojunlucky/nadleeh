package plugin

import (
	"nadleeh/pkg/env"
	"nadleeh/pkg/workflow/plugin/githubaction"
	"nadleeh/pkg/workflow/run_context"
)
import "nadleeh/pkg/workflow/plugin/googledrive"

var SupportedPlugins = []string{"google-drive"}

type Plugin interface {
	Init(ctx *run_context.WorkflowRunContext, config map[string]string) error
	Run(parent env.Env) error
}

func NewPlugin(name string) Plugin {
	if name == "google-drive" {
		return &googledrive.GoogleDrive{}
	} else if name == "github-action" {
		return &githubaction.GitHubAction{}
	}
	return nil
}
