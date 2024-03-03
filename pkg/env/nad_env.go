package env

import (
	"os"
	"strings"
)

type NadEnv struct {
	Parent Env
	envs   map[string]string
}

func NewEnv(parent Env, envs *map[string]string) *NadEnv {
	if parent == nil {
		parent = OSEnv
	}
	env := &NadEnv{
		Parent: parent,
	}
	if envs != nil {
		env.envs = *envs
	} else {
		env.envs = make(map[string]string)
	}
	return env
}

func (env *NadEnv) Get(key string) string {
	if val, ok := env.envs[key]; ok {
		return val
	}
	return os.Getenv(key)
}

func (env *NadEnv) Set(key, value string) {
	env.envs[key] = value
}

func (env *NadEnv) SetAll(envs map[string]string) {
	for key, value := range envs {
		env.envs[key] = value
	}
}

func (env *NadEnv) GetAll() map[string]string {
	envs := make(map[string]string)

	for key, value := range env.envs {
		envs[key] = value
	}
	for _, envStr := range os.Environ() {
		key, value, found := strings.Cut(envStr, "=")
		if !found {
			continue
		}

		if _, ok := env.envs[key]; !ok {
			envs[key] = value
		}
	}

	return envs

}

func (env *NadEnv) Expand(s string) string {
	return os.Expand(s, func(s string) string {
		return env.Get(s)
	})
}
