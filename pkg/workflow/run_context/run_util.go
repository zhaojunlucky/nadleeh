package run_context

import (
	log "github.com/sirupsen/logrus"
	"nadleeh/pkg/env"
)

func InterpretPluginCfg(ctx *WorkflowRunContext, parent env.Env, config map[string]string, variables map[string]interface{}) (map[string]string, error) {
	newMap := make(map[string]string)
	for k, v := range config {
		val, err := ctx.JSCtx.EvalActionScriptStr(parent, v, variables)
		if err != nil {
			log.Errorf("Failed to eval %s: %v", v, err)
			return nil, err
		}
		newMap[k] = val
	}
	return newMap, nil
}
