package env

import (
	"nadleeh/pkg/util"
	"os"
	"strings"
)

var OSEnv = NewOSEnv()

type NadReadEnv struct {
	Parent Env
	envs   map[string]string
}

func NewOSEnv() Env {
	env := NadReadEnv{
		Parent: nil,
		envs:   make(map[string]string),
	}
	env.initOSEnv()
	return &env
}

func NewReadEnv(parent Env, envs map[string]string) Env {
	env := NadReadEnv{
		Parent: parent,
		envs:   envs,
	}
	return &env
}

func (env *NadReadEnv) initOSEnv() {
	for _, envStr := range os.Environ() {
		sepIndex := strings.Index(envStr, "=")
		if sepIndex < 0 {
			continue
		}
		env.envs[envStr[:sepIndex]] = envStr[sepIndex+1:]
	}
}

func (env *NadReadEnv) Get(key string) string {
	if val, ok := env.envs[key]; ok {
		return val
	}
	return os.Getenv(key)
}

func (env *NadReadEnv) GetAll() map[string]string {
	newEnv := util.CopyMap(env.envs)
	if env.Parent != nil {
		for key, value := range env.Parent.GetAll() {
			if _, ok := newEnv[key]; ok {
				continue
			}
			newEnv[key] = value
		}
	}
	return newEnv
}

func (env *NadReadEnv) Set(key, value string) {

}

func (env *NadReadEnv) SetAll(envs map[string]string) {

}

func (env *NadReadEnv) Expand(s string) string {
	return os.Expand(s, func(s string) string {
		return env.Get(s)
	})
}
