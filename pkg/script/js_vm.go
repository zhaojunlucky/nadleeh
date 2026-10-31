package script

import (
	"nadleeh/pkg/common"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
)

type JSVm struct {
	Vm  *goja.Runtime
	ssh *NSSSHManager
}

func (vm *JSVm) Shutdown() {
	vm.ssh.Close()
}

func NewJSVm(printers ...console.Printer) *JSVm {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	var consolePrinter console.Printer = printer
	if len(printers) > 0 && printers[0] != nil {
		consolePrinter = printers[0]
	}

	registry := new(require.Registry)
	registry.Enable(vm)
	registry.RegisterNativeModule(console.ModuleName, console.RequireWithPrinter(consolePrinter))
	console.Enable(vm)

	vm.GlobalObject().Set("sys", common.Sys.GetInfo().GetAll())
	vm.GlobalObject().Set("file", &NJSFile{})
	vm.GlobalObject().Set("http", &NJSHttp{})
	vm.GlobalObject().Set("core", &NJSCore{})

	sshManager := &NSSSHManager{}
	vm.GlobalObject().Set("ssh", sshManager)

	return &JSVm{
		Vm:  vm,
		ssh: sshManager,
	}
}
