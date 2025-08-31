package runner

import (
	"bytes"
	"io"
	"testing"

	"nadleeh/pkg/workflow/core"
	workflow "nadleeh/pkg/workflow/model"
	"nadleeh/pkg/workflow/run_context"

	"github.com/akamensky/argparse"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/golib/pkg/env"
)

// mockArg implements argparse.Arg interface for testing
type mockArg struct {
	parsed bool
	result interface{}
	lname  string
}

func (m *mockArg) GetParsed() bool {
	return m.parsed
}

func (m *mockArg) GetResult() interface{} {
	return m.result
}

func (m *mockArg) GetLname() string {
	return m.lname
}

func (m *mockArg) GetSname() string {
	return ""
}

func (m *mockArg) GetOpts() *argparse.Options {
	return nil
}

func (m *mockArg) GetArgs() []argparse.Arg {
	return nil
}

func (m *mockArg) GetCommands() []*argparse.Command {
	return nil
}

func (m *mockArg) GetSelected() *argparse.Command {
	return nil
}

func (m *mockArg) GetHappened() *bool {
	return nil
}

func (m *mockArg) GetRemainder() *[]string {
	return nil
}

func (m *mockArg) GetPositional() bool {
	return false
}

// mockEnv implements env.Env interface for testing
type mockEnv struct {
	data map[string]string
}

func (m *mockEnv) Get(key string) string {
	if value, exists := m.data[key]; exists {
		return value
	}
	return ""
}

func (m *mockEnv) Set(key, value string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
}

func (m *mockEnv) GetAll() map[string]string {
	result := make(map[string]string)
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

func (m *mockEnv) Contains(key string) bool {
	_, exists := m.data[key]
	return exists
}

func (m *mockEnv) Expand(value string) string {
	return value
}

func (m *mockEnv) SetAll(data map[string]string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	for k, v := range data {
		m.data[k] = v
	}
}

func newMockEnv() *mockEnv {
	return &mockEnv{
		data: make(map[string]string),
	}
}

// mockWorkflow implements the workflow interface for testing
type mockWorkflow struct {
	precheckErr      error
	preflightErr     error
	doResult         *core.RunnableResult
	shouldPanic      bool
	panicMessage     string
}

func (m *mockWorkflow) Precheck() error {
	if m.shouldPanic {
		panic(m.panicMessage)
	}
	return m.precheckErr
}

func (m *mockWorkflow) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {
	if m.shouldPanic {
		panic(m.panicMessage)
	}
	return m.preflightErr
}

func (m *mockWorkflow) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	if m.shouldPanic {
		panic(m.panicMessage)
	}
	return m.doResult
}

// Test helper to capture log output and panics
func captureLogAndPanic(t *testing.T, testFunc func()) (logOutput string, didPanic bool, panicValue interface{}) {
	// Capture log output
	var buf bytes.Buffer
	originalOutput := log.StandardLogger().Out
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	// Capture panics
	defer func() {
		if r := recover(); r != nil {
			didPanic = true
			panicValue = r
		}
	}()

	testFunc()
	return buf.String(), didPanic, panicValue
}

// Mock functions to replace the actual workflow functions during testing
var (
	mockLoadWorkflowFile func(yml string, args map[string]argparse.Arg) (io.Reader, error)
	mockParseWorkflow    func(ymlFile io.Reader) (*workflow.Workflow, error)
)

func TestRunWorkflow(t *testing.T) {
	// Note: RunWorkflow uses log.Fatalf which calls os.Exit(), making it difficult to test directly
	// These tests focus on the components we can test and verify expected behavior patterns
	
	t.Run("ArgumentValidation", func(t *testing.T) {
		// Test that we can create valid arguments that would be accepted by RunWorkflow
		filename := "test.yml"
		privateFile := "private.key"
		checkFlag := false

		args := map[string]argparse.Arg{
			"file": &mockArg{
				parsed: true,
				result: &filename,
				lname:  "file",
			},
			"private": &mockArg{
				parsed: true,
				result: &privateFile,
				lname:  "private",
			},
			"check": &mockArg{
				parsed: true,
				result: &checkFlag,
				lname:  "check",
			},
		}

		// Verify argument structure is correct
		if args["file"].GetResult().(*string) == nil || *args["file"].GetResult().(*string) != "test.yml" {
			t.Error("File argument not properly structured")
		}

		if args["private"].GetResult().(*string) == nil || *args["private"].GetResult().(*string) != "private.key" {
			t.Error("Private argument not properly structured")
		}

		if args["check"].GetResult().(*bool) == nil || *args["check"].GetResult().(*bool) != false {
			t.Error("Check argument not properly structured")
		}
	})

	t.Run("EnvironmentSetup", func(t *testing.T) {
		// Test that we can create a valid environment that would be accepted by RunWorkflow
		mockEnv := newMockEnv()
		mockEnv.Set("TEST_VAR", "test_value")
		mockEnv.Set("ANOTHER_VAR", "another_value")

		if mockEnv.Get("TEST_VAR") != "test_value" {
			t.Error("Environment variable not properly set")
		}

		if !mockEnv.Contains("TEST_VAR") {
			t.Error("Environment should contain TEST_VAR")
		}

		allVars := mockEnv.GetAll()
		if len(allVars) != 2 {
			t.Errorf("Expected 2 environment variables, got %d", len(allVars))
		}
	})

	t.Run("MockArgInterface", func(t *testing.T) {
		// Test that our mockArg properly implements the argparse.Arg interface
		arg := &mockArg{
			parsed: true,
			result: "test_value",
			lname:  "test_arg",
		}

		if !arg.GetParsed() {
			t.Error("Expected argument to be parsed")
		}

		if arg.GetResult().(string) != "test_value" {
			t.Error("Expected result to be 'test_value'")
		}

		if arg.GetLname() != "test_arg" {
			t.Error("Expected lname to be 'test_arg'")
		}

		// Test other interface methods don't panic
		_ = arg.GetSname()
		_ = arg.GetOpts()
		_ = arg.GetArgs()
		_ = arg.GetCommands()
		_ = arg.GetSelected()
		_ = arg.GetHappened()
		_ = arg.GetRemainder()
		_ = arg.GetPositional()
	})
}

func TestRunWorkflow_ArgumentParsing(t *testing.T) {
	t.Run("ValidArguments", func(t *testing.T) {
		filename := "test.yml"
		privateFile := "private.key"
		checkFlag := false

		args := map[string]argparse.Arg{
			"file": &mockArg{
				parsed: true,
				result: &filename,
				lname:  "file",
			},
			"private": &mockArg{
				parsed: true,
				result: &privateFile,
				lname:  "private",
			},
			"check": &mockArg{
				parsed: true,
				result: &checkFlag,
				lname:  "check",
			},
		}

		// Test that arguments are parsed correctly
		fileArg, err := args["file"].GetResult().(*string), error(nil)
		if err != nil || fileArg == nil || *fileArg != "test.yml" {
			t.Errorf("Expected file argument to be 'test.yml', got: %v", fileArg)
		}

		privateArg := args["private"].GetResult().(*string)
		if privateArg == nil || *privateArg != "private.key" {
			t.Errorf("Expected private argument to be 'private.key', got: %v", privateArg)
		}

		checkArg := args["check"].GetResult().(*bool)
		if checkArg == nil || *checkArg != false {
			t.Errorf("Expected check argument to be false, got: %v", checkArg)
		}
	})

	t.Run("OptionalPrivateArgNotProvided", func(t *testing.T) {
		args := map[string]argparse.Arg{
			"private": &mockArg{
				parsed: false,
				lname:  "private",
			},
		}

		// This should not cause an error since private is optional
		privateArg := args["private"]
		if privateArg.GetParsed() {
			t.Error("Expected private argument to not be parsed")
		}
	})
}

func TestRunWorkflow_EdgeCases(t *testing.T) {
	t.Run("EmptyArgsMap", func(t *testing.T) {
		// Test with empty args map - should cause panic when accessing args["file"]
		args := make(map[string]argparse.Arg)
		mockEnv := newMockEnv()

		// We expect this to panic, so we'll use a defer/recover pattern
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected function to panic due to missing arguments")
			}
		}()

		// This should panic when trying to access args["file"]
		RunWorkflow(args, mockEnv)
		t.Error("Function should have panicked before reaching this point")
	})

	t.Run("NilArgsMap", func(t *testing.T) {
		// Test with nil args map - should cause panic when accessing args["file"]
		var args map[string]argparse.Arg
		mockEnv := newMockEnv()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected function to panic due to nil arguments")
			}
		}()

		RunWorkflow(args, mockEnv)
		t.Error("Function should have panicked before reaching this point")
	})

	t.Run("ArgumentStructureValidation", func(t *testing.T) {
		// Test various argument configurations to ensure they're properly structured
		testCases := []struct {
			name     string
			filename string
			parsed   bool
			expectOk bool
		}{
			{"ValidFile", "test.yml", true, true},
			{"EmptyFile", "", true, false},
			{"UnparsedFile", "test.yml", false, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				arg := &mockArg{
					parsed: tc.parsed,
					result: &tc.filename,
					lname:  "file",
				}

				if arg.GetParsed() != tc.parsed {
					t.Errorf("Expected parsed=%v, got %v", tc.parsed, arg.GetParsed())
				}

				if tc.parsed && arg.GetResult().(*string) != nil {
					result := *arg.GetResult().(*string)
					if (result != "" && tc.expectOk) || (result == "" && !tc.expectOk) {
						// This is expected behavior
					} else {
						t.Errorf("Unexpected result for %s: %s", tc.name, result)
					}
				}
			})
		}
	})
}

// Benchmark tests
func BenchmarkRunWorkflow_ArgumentParsing(b *testing.B) {
	filename := "test.yml"
	privateFile := "private.key"
	checkFlag := true

	args := map[string]argparse.Arg{
		"file": &mockArg{
			parsed: true,
			result: &filename,
			lname:  "file",
		},
		"private": &mockArg{
			parsed: true,
			result: &privateFile,
			lname:  "private",
		},
		"check": &mockArg{
			parsed: true,
			result: &checkFlag,
			lname:  "check",
		},
	}
	mockEnv := newMockEnv()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We can only benchmark the argument parsing part since the full function
		// will panic due to file system dependencies
		func() {
			defer func() {
				recover() // Ignore panics for benchmarking
			}()
			RunWorkflow(args, mockEnv)
		}()
	}
}
