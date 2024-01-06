package main

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"io"
	"os"
)

func ParamTest(str *string) {
	if str == nil {
		fmt.Println("nil str")
	} else {
		fmt.Println(*str)
	}
}
func main() {
	vm := goja.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
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
	_, err := vm.RunString(`
paramTest(null)
paramTest("hello world")

	`)
	if err != nil {
		fmt.Println(err)
	}
}
