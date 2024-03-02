package script

import (
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"nadleeh/pkg/env"
)
import "github.com/dop251/goja"

type JSContext struct {
}

func (js *JSContext) Run(env env.Env, script string) (int, string, error) {
	vm := NewJSVm()
	vm.Set("env", env)
	val, err := vm.RunString(script)
	output := ""
	if val != goja.Undefined() && val != goja.Null() {
		output = val.String()
	}
	if err != nil {
		return 1, output, err
	}
	return 0, output, nil
}

func NewJSVm() *goja.Runtime {
	vm := goja.New()
	//vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	new(require.Registry).Enable(vm)
	console.Enable(vm)

	vm.GlobalObject().Set("file", &NJSFile{})
	vm.GlobalObject().Set("http", &NJSHttp{})

	return vm
}
