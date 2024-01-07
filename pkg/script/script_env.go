package script

import "os"

type NJSEnv struct {
	envs map[string]string
}

func (js *NJSEnv) Get(key string) string {
	if val, ok := js.envs[key]; ok {
		return val
	}
	return os.Getenv(key)
}

func (js *NJSEnv) Set(key, value string) {
	js.envs[key] = value
}
