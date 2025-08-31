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
	Config     map[string]string
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

func (t *Telegram) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {

	return nil
}

func (t *Telegram) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	log.Infof("Run telegram plugin")
	err := t.validate(runCtx, parent, ctx.GenerateMap())
	if err != nil {
		return core.NewRunnableResult(err)
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", t.tgBotKey, t.channel, t.message)

	resp, err := http.Get(url)
	if err != nil {
		return core.NewRunnableResult(fmt.Errorf("error sending GET request: %v", err))
	}
	// Ensure the response body is closed when the function exits
	// This is crucial to prevent resource leaks
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		return core.NewRunnableResult(fmt.Errorf("received non-OK HTTP status: %d", resp.StatusCode))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return core.NewRunnableResult(fmt.Errorf("error reading response body: %v", err))
	}

	// Print the response body as a string
	log.Infof("Response Body:\n%s", body)
	return core.NewRunnableResult(nil)
}

func (t *Telegram) validate(runCtx *run_context.WorkflowRunContext, parent env.Env, variables map[string]interface{}) error {
	var err error
	t.Config, err = run_context.InterpretPluginCfg(runCtx, parent, t.Config, variables)
	if err != nil {
		return err
	}
	t.tgBotKey = parent.Expand(t.Config["key"])
	if len(t.tgBotKey) <= 0 {
		return fmt.Errorf("invalid tg-bot-key")
	}
	t.channel = parent.Expand(t.Config["channel"])
	if len(t.channel) <= 0 {
		return fmt.Errorf("invalid channel")
	}
	t.message = parent.Expand(t.Config["message"])
	if len(t.message) <= 0 {
		return fmt.Errorf("invalid message")
	}
	return nil
}
