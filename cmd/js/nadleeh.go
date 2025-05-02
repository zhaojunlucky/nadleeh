package main

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"io"
	"nadleeh/pkg/script"
	"os"
	"reflect"
)

func ParamTest(str *string) {
	if str == nil {
		fmt.Println("nil str")
	} else {
		fmt.Println(*str)
	}
}

type Response struct {
	Data map[string]any
}

func main() {
	vm := goja.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	vm.GlobalObject().Set("file", &script.NJSFile{})
	vm.GlobalObject().Set("http", &script.NJSHttp{})
	vm.Set("create", Response{
		Data: map[string]any{
			"id": 1,
		},
	})

	data := &map[string]any{
		"id": 1,
	}

	vm.GlobalObject().Set("data", data)

	vm.GlobalObject().Set("readFile", func(name string) (*string, error) {
		file, err := os.Open(name)
		if err != nil {
			return nil, err
		}
		bytes, err := io.ReadAll(file)
		text := string(bytes)
		return &text, nil
	})

	vm.GlobalObject().Set("paramTest", ParamTest)
	//	o, err := vm.RunString(`
	//console.log(http.Get)
	//const resp = http.Get("https://google.com")
	//console.log(JSON.stringify(resp))
	//	`)

	o, err := vm.RunString(`

 var a = 1

a == 1
`)
	fmt.Println("====")
	z := o.Export()
	fmt.Println(reflect.TypeOf(z))
	//fmt.Println(z)

	if err != nil {
		fmt.Println(err)
	}
}
