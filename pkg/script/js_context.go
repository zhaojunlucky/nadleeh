package script

import (
	"fmt"

	"github.com/dop251/goja/parser"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"

	"nadleeh/pkg/encrypt"
	"nadleeh/pkg/util"
	"nadleeh/pkg/util/js_token"
	"reflect"
	"slices"
	"strings"
)
import "github.com/dop251/goja"

var (
	printer = console.StdPrinter{
		StdoutPrint: func(s string) { fmt.Println(s) },
		StderrPrint: func(s string) { fmt.Println(s) },
	}
)

type jsScriptProgram struct {
	program *goja.Program
	err     error
}

type JSContext struct {
	JSSecCtx      JSSecureContext
	scriptProgram map[string]*jsScriptProgram
	count         int
}

var unAllowedEnvKeys = []string{"secure", "env", "http", "core", "file"}

func (js *JSContext) Compile(script string) error {
	script = strings.TrimSpace(script)
	sp := js.scriptProgram[script]
	if sp != nil {
		return sp.err
	}
	program, err := goja.Compile(fmt.Sprintf("script_%d", js.count), script, true)
	js.count++
	if err != nil {
		js.scriptProgram[script] = &jsScriptProgram{
			err: err,
		}
		return err
	}
	js.scriptProgram[script] = &jsScriptProgram{
		program: program,
	}

	return nil

}

func (js *JSContext) getFileKey(jsFile string) (string, error) {
	sha256, err := util.CalculateFileSHA256(jsFile)
	if err != nil {
		log.Errorf("failed to calc hash for file %s: %v", jsFile, err)
		return "", err
	}
	return fmt.Sprintf("%s:%s", jsFile, sha256), nil
}

func (js *JSContext) CompileFile(jsFile string) (string, error) {
	fileKey, err := js.getFileKey(jsFile)
	if err != nil {
		js.scriptProgram[fileKey] = &jsScriptProgram{
			err: err,
		}
		return fileKey, err
	}

	sp := js.scriptProgram[fileKey]
	if sp != nil {
		if sp.err != nil {
			return fileKey, sp.err
		} else {
			return fileKey, nil
		}
	}

	prg, err := parser.ParseFile(nil, jsFile, nil, 0)
	if err != nil {
		log.Errorf("failed to parse javascript file %s: %v", jsFile, err)
		js.scriptProgram[fileKey] = &jsScriptProgram{
			err: err,
		}
		return fileKey, err
	}

	prog, err := goja.CompileAST(prg, true)

	if err != nil {
		log.Errorf("failed to compile javascript file %s: %v", jsFile, err)
		js.scriptProgram[fileKey] = &jsScriptProgram{
			err: err,
		}
		return fileKey, err
	}

	js.scriptProgram[fileKey] = &jsScriptProgram{
		program: prog,
	}

	return fileKey, nil
}

func (js *JSContext) runWithProgram(vm *goja.Runtime, script string) (goja.Value, error) {
	script = strings.TrimSpace(script)
	sp := js.scriptProgram[script]
	if sp != nil {
		if sp.err != nil {
			return nil, sp.err
		} else {
			return vm.RunProgram(sp.program)
		}
	}

	return vm.RunString(script)
}

func (js *JSContext) RunFile(env env.Env, jsFile string, variables map[string]interface{}) (int, string, error) {
	fileKey, err := js.CompileFile(jsFile)
	if err != nil {
		return 1, "", err
	}

	st := js.scriptProgram[fileKey]
	if st == nil || st.err != nil {
		return 2, "", fmt.Errorf("invalud file %s", jsFile)
	}

	vm := NewJSVm()
	vm.Set("env", env)
	vm.Set("secure", &js.JSSecCtx)

	for k, v := range variables {
		if slices.Contains(unAllowedEnvKeys, k) {
			return 1, "", fmt.Errorf("key %s is not allowed", k)
		}
		vm.Set(k, v)
	}

	val, err := vm.RunProgram(st.program)
	output := ""
	if val != nil && val != goja.Undefined() && val != goja.Null() {
		output = val.String()
	}
	if err != nil {
		return 1, output, err
	}
	return 0, output, nil

}

func (js *JSContext) Run(env env.Env, script string, variables map[string]interface{}) (int, string, error) {
	vm := NewJSVm()
	vm.Set("env", env)
	vm.Set("secure", &js.JSSecCtx)

	for k, v := range variables {
		if slices.Contains(unAllowedEnvKeys, k) {
			return 1, "", fmt.Errorf("key %s is not allowed", k)
		}
		vm.Set(k, v)
	}

	val, err := js.runWithProgram(vm, script)
	output := ""
	if val != nil && val != goja.Undefined() && val != goja.Null() {
		output = val.String()
	}
	if err != nil {
		return 1, output, err
	}
	return 0, output, nil
}

func (js *JSContext) Eval(env env.Env, expression string, variables map[string]interface{}) (goja.Value, error) {
	vm := NewJSVm()
	vm.Set("env", env.GetAll())
	vm.Set("secure", &js.JSSecCtx)

	for k, v := range variables {
		if slices.Contains(unAllowedEnvKeys, k) {
			return nil, fmt.Errorf("key %s is not allowed", k)
		}
		vm.Set(k, v)
	}

	return js.runWithProgram(vm, expression)
}

func (js *JSContext) EvalBool(env env.Env, expression string, variables map[string]interface{}) (bool, error) {
	val, err := js.Eval(env, expression, variables)
	if val != nil && val != goja.Undefined() && val != goja.Null() {
		raw := val.Export()
		rawType := reflect.ValueOf(raw)
		switch rawType.Kind() {
		case reflect.Bool:
			return rawType.Bool(), nil
		case reflect.String:
			return util.Str2Bool(rawType.String()), err
		case reflect.Int:
			return util.Int2Bool(rawType.Int()), nil
		default:
			return false, fmt.Errorf("invalid output %v of expression %s", raw, expression)
		}
	}
	return false, fmt.Errorf("invalid expression %s, no output", expression)
}
func (js *JSContext) EvalStr(env env.Env, expression string, variables map[string]interface{}) (string, error) {

	val, err := js.Eval(env, expression, variables)
	if val != nil && val != goja.Undefined() && val != goja.Null() {
		raw := val.Export()
		rawType := reflect.ValueOf(raw)
		switch rawType.Kind() {
		case reflect.Bool:
			return fmt.Sprintf("%v", rawType.Bool()), nil
		case reflect.String:
			return rawType.String(), err
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fmt.Sprintf("%d", rawType.Int()), nil
		case reflect.Float32, reflect.Float64:
			return fmt.Sprintf("%f", rawType.Float()), nil
		default:
			return "", fmt.Errorf("invalid output %v of expression %s", raw, expression)
		}
	}
	return "", fmt.Errorf("invalid expression %s, no output", expression)
}

func (js *JSContext) EvalActionScriptBool(env env.Env, expression string, variables map[string]interface{}) (bool, error) {
	scanner := js_token.JSTokenScanner{}

	tokens, err := scanner.Scan(expression)
	if err != nil {
		log.Errorf("Failed to scan %s: %v", expression, err)
		return false, err
	}

	if len(tokens) == 0 {
		return false, fmt.Errorf("invalid empty expression %s", expression)
	}

	if len(tokens) > 1 || tokens[0].Type == js_token.RawString {
		return false, fmt.Errorf("invalid expression %s, only one expression is allowed", expression)
	}
	val, err := js.EvalBool(env, tokens[0].Value, variables)
	if err != nil {
		log.Errorf("Failed to eval %s: %v", tokens[0].Value, err)
		return false, err
	}
	return val, nil
}

func (js *JSContext) EvalActionScriptStr(env env.Env, expression string, variables map[string]interface{}) (string, error) {
	scanner := js_token.JSTokenScanner{}

	tokens, err := scanner.Scan(expression)
	if err != nil {
		log.Errorf("Failed to scan %s: %v", expression, err)
		return "", err
	}
	var data []string

	for _, token := range tokens {
		if token.Type == js_token.RawString {
			data = append(data, token.Value)
		} else {
			val, err := js.EvalStr(env, token.Value, variables)
			if err != nil {
				log.Errorf("Failed to eval %s: %v", token.Value, err)
				return "", err
			}
			data = append(data, val)
		}
	}
	if len(data) == 0 {
		return "", nil
	}
	return strings.Join(data, ""), nil

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
	vm.GlobalObject().Set("core", &NJSCore{})

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
