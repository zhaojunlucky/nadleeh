package githubaction

import (
	"context"
	"fmt"
	"nadleeh/pkg/script"
	"nadleeh/pkg/util"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
)

const Download = "download-artifact"
const GhActionDownloadArtifact = "GH_ACTION_DOWNLOAD_ARTIFACT"

type GitHubAction struct {
	Version         string
	PluginPath      string
	ctx             *run_context.WorkflowRunContext
	organization    string
	repository      string
	branch          string
	path            string
	token           string
	action          string
	pr              int
	artifactPathEnv string
	config          map[string]string
}

func (g *GitHubAction) GetName() string {
	return "github-action"
}

func (g *GitHubAction) Init(ctx *run_context.WorkflowRunContext, config map[string]string) error {
	g.ctx = ctx
	g.config = config
	return nil
}

func (g *GitHubAction) Resolve() error {
	return nil
}

func (g *GitHubAction) CanRun() bool {
	return true
}

func (g *GitHubAction) Compile(runCtx run_context.WorkflowRunContext) error {
	return nil
}

func (g *GitHubAction) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	err := g.initConfig(parent, ctx.GenerateMap())
	if err != nil {
		return core.NewRunnableResult(err)
	}
	fmt.Printf("Run GitHub Action plugin, action %s", g.action)
	client := github.NewClient(nil)
	client = client.WithAuthToken(g.token)
	var pr *github.PullRequest
	if g.pr > 0 {
		pr, _, err = client.PullRequests.Get(context.Background(), g.organization, g.repository, g.pr)
		if err != nil {
			return core.NewRunnableResult(err)
		}
	}

	artifacts, _, err := client.Actions.ListArtifacts(context.Background(), g.organization, g.repository, nil)
	if err != nil {
		return core.NewRunnableResult(err)
	}
	log.Infof("found %d artifacts for repo %s/%s first page", len(artifacts.Artifacts), g.organization, g.repository)
	for _, arti := range artifacts.Artifacts {
		if *arti.Expired {
			log.Warnf("artifact %s expired", arti.GetName())
			continue
		}

		if len(g.branch) > 0 && *arti.GetWorkflowRun().HeadBranch != g.branch {
			log.Warnf("artifact %s not for branch %s", arti.GetName(), g.branch)
			continue
		}
		workflowRun := *arti.GetWorkflowRun()

		if pr != nil {
			if workflowRun.GetHeadBranch() != pr.Head.GetRef() || workflowRun.GetHeadSHA() != pr.Head.GetSHA() {
				log.Warnf("artifact %s not for pr %d, head sha or branch not match", arti.GetName(), g.pr)
				continue
			}
		}

		log.Infof("found artifact %s", arti.GetName())

		_, _, err := g.ctx.ShellCtx.Run(parent, fmt.Sprintf("mkdir -p %s", g.path), false)
		if err != nil {
			return core.NewRunnableResult(err)
		}
		artiIrl := arti.GetArchiveDownloadURL()
		if len(artiIrl) <= 0 {
			log.Errorf("invalid download url for artifact %s", arti.GetName())
			continue
		}
		jsHttp := script.NJSHttp{}
		headers := map[string]string{"Accept": "application/vnd.github+json", "Authorization": fmt.Sprintf("Bearer %s", g.token), "X-GitHub-Api-Version": "2022-11-28"}
		artifactPath := path.Join(g.path, fmt.Sprintf("%s.zip", arti.GetName()))
		err = jsHttp.DownloadFile("GET", artiIrl, artifactPath, &headers, nil)
		if err != nil {
			return core.NewRunnableResult(err)
		}
		artiEnv := GhActionDownloadArtifact
		if len(g.artifactPathEnv) > 0 {
			artiEnv = g.artifactPathEnv
		}
		log.Infof("set downloaded artifact path as env %s=%s", artiEnv, artifactPath)
		parent.Set(artiEnv, artifactPath)
		break
	}
	return core.NewRunnableResult(nil)
}

func (g *GitHubAction) initConfig(env env.Env, variables map[string]interface{}) error {
	var err error
	g.config, err = run_context.InterpretPluginCfg(g.ctx, env, g.config, variables)
	if err != nil {
		return err
	}

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

	if util.HasKey(g.config, "branch") {
		g.branch = env.Expand(g.config["branch"])
	}

	if util.HasKey(g.config, "artifact-path-env") {
		g.artifactPathEnv = env.Expand(g.config["artifact-path-env"])
	}

	if util.HasKey(g.config, "pr") {
		pr := env.Expand(g.config["pr"])

		if len(pr) > 0 {
			prRegex := regexp.MustCompile("\\d+")
			if !prRegex.MatchString(pr) {
				return fmt.Errorf("invalid pr, should be number")
			}
			prNum, err := strconv.ParseInt(pr, 10, 64)
			if err != nil {
				log.Errorf("invalid pr: %s", pr)
				return err
			}
			g.pr = int(prNum)
		}
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

	if g.action != Download {
		return fmt.Errorf("invalid action, only support %s", Download)
	}

	return nil
}
