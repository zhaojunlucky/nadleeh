package script

import (
	"fmt"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"nadleeh/pkg/encrypt"
	"nadleeh/pkg/env"
)
import "github.com/dop251/goja"

var (
	printer = console.StdPrinter{
		StdoutPrint: func(s string) { fmt.Println(s) },
		StderrPrint: func(s string) { fmt.Println(s) },
	}
)

type JSContext struct {
	JSSecCtx JSSecureContext
}

func (js *JSContext) Run(env env.Env, script string) (int, string, error) {
	vm := NewJSVm()
	vm.Set("env", env)
	vm.Set("secure", &js.JSSecCtx)
	val, err := vm.RunString(script)
	output := ""

	if val != nil && val != goja.Undefined() && val != goja.Null() {
		output = val.String()
	}
	if err != nil {
		return 1, output, err
	}
	return 0, output, nil
}

func NewJSVm() *goja.Runtime {

	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	registry := new(require.Registry)
	registry.Enable(vm)
	registry.RegisterNativeModule(console.ModuleName, console.RequireWithPrinter(printer))
	console.Enable(vm)

	vm.GlobalObject().Set("file", &NJSFile{})
	vm.GlobalObject().Set("http", &NJSHttp{})

	return vm
}

func NewJSContext(secCtx *encrypt.SecureContext) JSContext {
	return JSContext{
		JSSecCtx: JSSecureContext{secureCtx: secCtx},
	}
}

type JSSecureContext struct {
	secureCtx *encrypt.SecureContext
}

func (js *JSSecureContext) IsEncrypted(str string) bool {
	return js.secureCtx.IsEncrypted(str)
}

func (js *JSSecureContext) Decrypt(str string) (*string, error) {
	val, err := js.secureCtx.DecryptStr(str)
	if err != nil {
		return nil, err
	} else {
		return &val, nil
	}
}
