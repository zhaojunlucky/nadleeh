package common

import (
	"fmt"

	"github.com/zhaojunlucky/golib/pkg/env"
)

type WriteOnParentEnv struct {
	parent env.Env

	comp env.Env
}

func (w *WriteOnParentEnv) Get(key string) string {
	return w.comp.Get(key)
}
func (w *WriteOnParentEnv) Set(key, value string) {
	w.parent.Set(key, value)

}
func (w *WriteOnParentEnv) SetAll(envs map[string]string) {
	w.parent.SetAll(envs)
}

func (w *WriteOnParentEnv) GetAll() map[string]string {
	return w.comp.GetAll()
}

func (w *WriteOnParentEnv) Expand(s string) string {
	return w.comp.Expand(s)
}
func (w *WriteOnParentEnv) Contains(s string) bool {
	return w.comp.Contains(s)
}

func NewWriteOnParentEnv(parent env.Env, envs map[string]string) *WriteOnParentEnv {
	return &WriteOnParentEnv{
		parent: parent,
		comp:   env.NewReadEnv(parent, envs),
	}
}

func CopyWriteOnParentEnvWithEmptyEnv(parent env.Env, envs map[string]string) (*WriteOnParentEnv, error) {
	other, ok := parent.(*WriteOnParentEnv)
	if !ok {
		return nil, fmt.Errorf("parent is not a WriteOnParentEnv")
	}
	return &WriteOnParentEnv{
		parent: other.parent,
		comp:   env.NewReadEnv(other.comp, envs),
	}, nil
}
