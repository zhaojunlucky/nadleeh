package githubaction

import (
	"context"
	"fmt"
	"github.com/google/go-github/v60/github"
	"nadleeh/pkg/env"
	"nadleeh/pkg/script"
	"nadleeh/pkg/workflow/run_context"
	"path"
	"strings"
)

const DOWNLOAD = "download-artifact"

type GitHubAction struct {
	ctx          *run_context.WorkflowRunContext
	organization string
	repository   string
	branch       string
	path         string
	token        string
	action       string
	config       map[string]string
}

func (g *GitHubAction) Init(ctx *run_context.WorkflowRunContext, config map[string]string) error {
	g.ctx = ctx
	g.config = config
	return nil
}

func (g *GitHubAction) Run(parent env.Env) error {
	err := g.initConfig(parent)
	if err != nil {
		return err
	}
	fmt.Printf("Run GitHub Action plugin, action %s\n", g.action)
	client := github.NewClient(nil)
	client = client.WithAuthToken(g.token)

	//opt := &github.ListOptions{
	//	Page:    100,
	//	PerPage: 1,
	//}
	artifacts, _, err := client.Actions.ListArtifacts(context.Background(), g.organization, g.repository, nil)
	if err != nil {
		return err
	}
	fmt.Printf("found %d artifacts for repo %s/%s first page\n", len(artifacts.Artifacts), g.organization, g.repository)
	for _, arti := range artifacts.Artifacts {
		if *arti.Expired {
			fmt.Printf("artifact %s expired\n", arti.GetName())
			continue
		}

		if *arti.GetWorkflowRun().HeadBranch == g.branch {
			fmt.Printf("found artifact %s\n", arti.GetName())

			_, _, err := g.ctx.ShellCtx.Run(parent, fmt.Sprintf("mkdir -p %s", g.path), false)
			if err != nil {
				return err
			}
			artiIrl := arti.GetArchiveDownloadURL()
			if len(artiIrl) <= 0 {
				fmt.Printf("invalid download url for artifact %s\n", arti.GetName())
				continue
			}
			jsHttp := script.NJSHttp{}
			headers := map[string]string{"Accept": "application/vnd.github+json", "Authorization": fmt.Sprintf("Bearer %s", g.token), "X-GitHub-Api-Version": "2022-11-28"}

			err = jsHttp.DownloadFile("GET", artiIrl, path.Join(g.path, arti.GetName()), &headers, nil)
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (g *GitHubAction) initConfig(env env.Env) error {
	g.repository = env.Expand(g.config["repository"])
	if len(g.repository) <= 0 {
		return fmt.Errorf("invalid repository")
	}

	orgRepo := strings.Split(g.repository, "/")
	if len(orgRepo) != 2 {
		return fmt.Errorf("invalid repository, should be org/repo")
	}
	g.organization = orgRepo[0]
	g.repository = orgRepo[1]

	g.branch = env.Expand(g.config["branch"])
	if len(g.branch) <= 0 {
		return fmt.Errorf("invalid branch")
	}

	g.path = env.Expand(g.config["path"])
	if len(g.path) <= 0 {
		return fmt.Errorf("invalid path")
	}

	g.token = env.Expand(g.config["token"])
	if len(g.token) > 0 && g.ctx.SecureCtx.IsEncrypted(g.token) {
		var err error
		g.token, err = g.ctx.SecureCtx.DecryptStr(g.token)
		if err != nil {
			return err
		}
	}

	g.action = env.Expand(g.config["action"])
	if len(g.action) <= 0 {
		return fmt.Errorf("invalid action")
	}

	if g.action != DOWNLOAD {
		return fmt.Errorf("invalid action, only support %s", DOWNLOAD)
	}

	return nil
}
