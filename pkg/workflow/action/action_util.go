package workflow

import (
	"nadleeh/pkg/env"
	"nadleeh/pkg/script"
)

func InterpretEnv(jsContext *script.JSContext, parent env.Env, envs map[string]string, variables map[string]interface{}) (*env.NadEnv, error) {
	nadEnv := env.NewEnv(parent, nil)
	if len(envs) == 0 {
		return nadEnv, nil
	}
	for k, v := range envs {
		val, err := jsContext.EvalActionScript(nadEnv, v, variables)
		if err != nil {
			return nil, err
		}
		nadEnv.Set(k, val)
	}

	return nadEnv, nil
}

func InterpretEnvSelf(jsContext *script.JSContext, parent *env.NadEnv, envs map[string]string, variables map[string]interface{}) error {
	if len(envs) == 0 {
		return nil
	}
	for k, v := range envs {
		val, err := jsContext.EvalActionScript(parent, v, variables)
		if err != nil {
			return err
		}
		parent.Set(k, val)
	}

	return nil
}
