package workflow

import (
	"nadleeh/pkg/script"

	"github.com/zhaojunlucky/golib/pkg/env"
)

func InterpretEnv(jsContext *script.JSContext, parent env.Env, envs map[string]string, variables map[string]interface{}) (*env.ReadWriteEnv, error) {
	nadEnv := env.NewReadWriteEnv(parent, nil)
	if len(envs) == 0 {
		return nadEnv, nil
	}
	for k, v := range envs {
		val, err := jsContext.EvalActionScriptStr(nadEnv, v, variables)
		if err != nil {
			return nil, err
		}
		nadEnv.Set(k, val)
	}

	return nadEnv, nil
}

func InterpretEnvSelf(jsContext *script.JSContext, parent *env.ReadWriteEnv, envs map[string]string, variables map[string]interface{}) error {
	if len(envs) == 0 {
		return nil
	}
	for k, v := range envs {
		val, err := jsContext.EvalActionScriptStr(parent, v, variables)
		if err != nil {
			return err
		}
		parent.Set(k, val)
	}

	return nil
}
