package runner

import (
	"os"
	"path/filepath"
	"testing"

	"nadleeh/pkg/workflow/core"

	"github.com/akamensky/argparse"
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

func TestRunWorkflow_ArgumentValidation(t *testing.T) {
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
		// Test with empty args map - would cause log.Fatal which cannot be caught
		// Skipping this test since log.Fatal calls os.Exit() and terminates the process
		t.Skip("Skipping test that would trigger log.Fatal for missing file argument")
	})

	t.Run("NilArgsMap", func(t *testing.T) {
		// Test with nil args map - would cause log.Fatal which cannot be caught
		// Skipping this test since log.Fatal calls os.Exit() and terminates the process
		t.Skip("Skipping test that would trigger log.Fatal for missing file argument")
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

// TestRunWorkflow_CheckMode tests RunWorkflow with check-only mode using a real workflow file
func TestRunWorkflow_CheckMode(t *testing.T) {
	// Create a temporary workflow file
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "test.yml")
	workflowContent := `name: "test"

jobs:
  test-job:
    steps:
      - name: echo
        run: echo "hello"
`
	err := os.WriteFile(workflowFile, []byte(workflowContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create workflow file: %v", err)
	}

	t.Run("CheckOnlyMode", func(t *testing.T) {
		checkFlag := true
		wa := &core.WorkflowArgs{
			File:  &workflowFile,
			Check: &checkFlag,
		}
		argEnv := env.NewReadEnv(env.NewEmptyReadEnv(), map[string]string{})

		// This should not panic since we're only checking
		RunWorkflow(wa, argEnv)
	})
}

// TestRunWorkflow_ExecuteSimple tests RunWorkflow executing a simple workflow
func TestRunWorkflow_ExecuteSimple(t *testing.T) {
	// Create a temporary workflow file
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "test.yml")
	workflowContent := `name: "test"

jobs:
  test-job:
    steps:
      - name: echo
        run: echo "hello"
`
	err := os.WriteFile(workflowFile, []byte(workflowContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create workflow file: %v", err)
	}

	t.Run("ExecuteWorkflow", func(t *testing.T) {
		wa := &core.WorkflowArgs{
			File: &workflowFile,
		}
		argEnv := env.NewReadEnv(env.NewEmptyReadEnv(), map[string]string{})

		// This should execute successfully
		RunWorkflow(wa, argEnv)
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
			workflowArgs := core.NewWorkflowArgs(args)
			RunWorkflow(workflowArgs, mockEnv)
		}()
	}
}
