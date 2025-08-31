package workflow

import (
	"nadleeh/pkg/common"
	"nadleeh/pkg/script"

	"github.com/zhaojunlucky/golib/pkg/env"
)

func InterpretNadEnv(jsContext *script.JSContext, parent env.Env, envs map[string]string, variables map[string]interface{}) (*env.ReadWriteEnv, error) {
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

func InterpretWriteOnParentEnv(jsContext *script.JSContext, parent env.Env, envs map[string]string, variables map[string]interface{}) (*common.WriteOnParentEnv, error) {
	newEnvs := make(map[string]string, len(envs))
	for k, v := range envs {
		val, err := jsContext.EvalActionScriptStr(parent, v, variables)
		if err != nil {
			return nil, err
		}
		newEnvs[k] = val
	}

	return common.NewWriteOnParentEnv(parent, newEnvs), nil
}
