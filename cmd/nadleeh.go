package main

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"io"
	"os"
)

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
	_, err := vm.RunString(`
try{
	console.log(readFile('/Users/jun/Downloads/ads.txt'))
console.log("normal")
} catch(e) {
console.log("catch")
	console.log(e)
}
	`)
	if err != nil {
		fmt.Println(err)
	}
}
