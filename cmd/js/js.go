package main

import (
	"fmt"
	"nadleeh/pkg/script"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	log "github.com/sirupsen/logrus"
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
	//prog, err := goja.Compile("/Users/jun/magicworldz/github/nadleeh/examples/js_plug/main.js", "", true)
	//registry := new(require.Registry)
	//runtime := goja.New()
	//req := registry.Enable(runtime)
	//value, err := req.Require("/Users/jun/magicworldz/github/nadleeh/examples/js_plug/core.mjs")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(value)
	prg, err := parser.ParseFile(nil, "/Users/jun/magicworldz/github/nadleeh/examples/js_plug/main.js", nil, 0)
	if err != nil {
		log.Fatal(err)
		return
	}

	prog, err := goja.CompileAST(prg, true)

	if err != nil {
		fmt.Printf("Failed to compile script: %v\n", err)
		return
	}

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
	val, err := vm.RunProgram(prog)
	fmt.Println(val)
	if err != nil {
		log.Fatal(err)
	}

	//	vm.GlobalObject().Set("readFile", func(name string) (*string, error) {
	//		file, err := os.Open(name)
	//		if err != nil {
	//			return nil, err
	//		}
	//		bytes, err := io.ReadAll(file)
	//		text := string(bytes)
	//		return &text, nil
	//	})
	//
	//	vm.GlobalObject().Set("paramTest", ParamTest)
	//	o, err := vm.RunString(`
	//	const resp = http.get("https://api.github.com/repos/zhaojunlucky/mkdocs-cms/releases")
	//	console.log(resp.body)
	//		`)
	//
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//
	//	o, err = vm.RunString(`
	//
	// var a = 1
	//console.log(data.id)
	//a == data.id
	//data.id = 22
	//`)
	//	fmt.Println("====")
	//	z := o.Export()
	//	fmt.Println(reflect.TypeOf(z))
	//	//fmt.Println(z)
	//
	//	if err != nil {
	//		fmt.Println(err)
	//	}
}
