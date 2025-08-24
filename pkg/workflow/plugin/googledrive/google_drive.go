package googledrive

import (
	"context"
	"encoding/json"
	"fmt"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	SCOPE = drive.DriveScope
)

type GoogleDrive struct {
	Version    string
	PluginPath string
	ctx        *run_context.WorkflowRunContext
	name       string
	path       string
	remotePath string
	config     map[string]string
	cred       string
}

func (g *GoogleDrive) GetName() string {
	return "google-drive"
}

func (g *GoogleDrive) CanRun() bool {
	return true
}

func (g *GoogleDrive) Compile(runCtx run_context.WorkflowRunContext) error {
	return nil
}

func (g *GoogleDrive) Resolve() error {
	return nil
}

func (g *GoogleDrive) Init(ctx *run_context.WorkflowRunContext, config map[string]string) error {
	g.ctx = ctx
	g.config = config

	return nil
}

func (g *GoogleDrive) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	log.Infof("Run Google Drive plugin")
	err := g.validate(parent, ctx.GenerateMap())
	if err != nil {
		return core.NewRunnableResult(err)
	}
	client := g.ServiceAccount(g.cred)
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client), option.WithScopes(SCOPE))
	if err != nil {
		return core.NewRunnableResult(err)
	}
	file, err := os.Open(g.path)
	if err != nil {
		return core.NewRunnableResult(err)
	}
	defer file.Close()
	f := &drive.File{Name: g.name, Parents: []string{g.remotePath}}
	res, err := srv.Files.
		Create(f).
		Media(file).
		ProgressUpdater(func(now, size int64) { log.Infof("%d, %d\r", now, size) }).
		Do()
	if err != nil {
		return core.NewRunnableResult(err)
	}
	log.Infof("https://drive.google.com/file/d/%s/view?usp=drive_link", res.Id)
	return core.NewRunnableResult(nil)
}

// ServiceAccount : Use Service account
func (g *GoogleDrive) ServiceAccount(credentialFile string) *http.Client {
	b, err := os.ReadFile(credentialFile)
	if err != nil {
		log.Fatal(err)
	}
	s := string(b)
	if g.ctx.SecureCtx.IsEncrypted(s) {
		b, err = g.ctx.SecureCtx.Decrypt(s)
		if err != nil {
			log.Fatal(err)
		}
	}
	var c = struct {
		Email      string `json:"client_email"`
		PrivateKey string `json:"private_key"`
	}{}
	json.Unmarshal(b, &c)
	config := &jwt.Config{
		Email:      c.Email,
		PrivateKey: []byte(c.PrivateKey),
		Scopes: []string{
			drive.DriveScope,
		},
		TokenURL: google.JWTTokenURL,
	}
	client := config.Client(context.TODO())
	return client
}

func (g *GoogleDrive) validate(parent env.Env, variables map[string]interface{}) error {
	var err error
	g.config, err = run_context.InterpretPluginCfg(g.ctx, parent, g.config, variables)
	if err != nil {
		return err
	}

	g.name = parent.Expand(g.config["name"])
	if len(g.name) <= 0 {
		return fmt.Errorf("invalid name")
	}
	g.path = parent.Expand(g.config["path"])
	if len(g.path) <= 0 {
		return fmt.Errorf("invalid path")
	}
	fi, err := os.Stat(g.path)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return fmt.Errorf("invalid path, it's not a file")
	}
	g.remotePath = parent.Expand(g.config["remote-path"])

	g.cred = parent.Expand(g.config["cred"])
	if len(g.cred) <= 0 {
		return fmt.Errorf("invalid cred")
	}
	return nil
}
