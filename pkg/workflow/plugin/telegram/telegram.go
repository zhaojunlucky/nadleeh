package telegram

import (
	"fmt"
	"nadleeh/pkg/workflow/core"

	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"

	"io"
	"nadleeh/pkg/workflow/run_context"
	"net/http"
)

type Telegram struct {
	Version    string
	PluginPath string
	ctx        *run_context.WorkflowRunContext
	config     map[string]string
	tgBotKey   string
	channel    string
	message    string
}

func (t *Telegram) GetName() string {
	return "telegram"
}

func (t *Telegram) CanRun() bool {
	return true
}

func (t *Telegram) Compile(runCtx run_context.WorkflowRunContext) error {
	return nil
}

func (t *Telegram) Resolve() error {
	return nil
}

func (t *Telegram) Init(ctx *run_context.WorkflowRunContext, config map[string]string) error {
	t.ctx = ctx
	t.config = config
	return nil
}

func (g *Telegram) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	log.Infof("Run telegram plugin")
	err := g.validate(parent, ctx.GenerateMap())
	if err != nil {
		return core.NewRunnableResult(err)
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", g.tgBotKey, g.channel, g.message)

	resp, err := http.Get(url)
	if err != nil {
		return core.NewRunnableResult(fmt.Errorf("Error sending GET request: %v", err))
	}
	// Ensure the response body is closed when the function exits
	// This is crucial to prevent resource leaks
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		return core.NewRunnableResult(fmt.Errorf("Received non-OK HTTP status: %d", resp.StatusCode))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return core.NewRunnableResult(fmt.Errorf("Error reading response body: %v", err))
	}

	// Print the response body as a string
	log.Infof("Response Body:\n%s", body)
	return core.NewRunnableResult(nil)
}

func (g *Telegram) validate(parent env.Env, variables map[string]interface{}) error {
	var err error
	g.config, err = run_context.InterpretPluginCfg(g.ctx, parent, g.config, variables)
	if err != nil {
		return err
	}
	g.tgBotKey = parent.Expand(g.config["key"])
	if len(g.tgBotKey) <= 0 {
		return fmt.Errorf("invalid tg-bot-key")
	}
	g.channel = parent.Expand(g.config["channel"])
	if len(g.channel) <= 0 {
		return fmt.Errorf("invalid channel")
	}
	g.message = parent.Expand(g.config["message"])
	if len(g.message) <= 0 {
		return fmt.Errorf("invalid message")
	}
	return nil
}
