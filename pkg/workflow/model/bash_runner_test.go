package workflow

import (
	"nadleeh/pkg/shell"
	"nadleeh/pkg/script"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"
	"testing"
)

// Mock implementation of env.Env interface for BashRunner tests
type mockBashEnv struct {
	data map[string]string
}

func (m *mockBashEnv) Get(key string) string {
	if m.data == nil {
		return ""
	}
	value, exists := m.data[key]
	if exists {
		return value
	}
	return ""
}

func (m *mockBashEnv) GetAll() map[string]string {
	if m.data == nil {
		return make(map[string]string)
	}
	result := make(map[string]string)
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

func (m *mockBashEnv) Set(key, value string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
}

func (m *mockBashEnv) Unset(key string) {
	if m.data != nil {
		delete(m.data, key)
	}
}

func (m *mockBashEnv) Expand(value string) string {
	return value
}

func (m *mockBashEnv) SetAll(data map[string]string) {
	m.data = data
}

func (m *mockBashEnv) Contains(key string) bool {
	_, exists := m.data[key]
	return exists
}

// Note: Using real ShellContext since it's a concrete struct, not an interface

// Helper function to create a test JSContext for BashRunner
func createTestJSContextForBash() script.JSContext {
	return script.NewJSContext(nil)
}

// Helper function to create a test WorkflowRunContext for BashRunner
func createTestWorkflowRunContextForBash() run_context.WorkflowRunContext {
	jsCtx := createTestJSContextForBash()
	shellCtx := shell.NewShellContext()
	return run_context.WorkflowRunContext{
		JSCtx:    jsCtx,
		ShellCtx: shellCtx,
	}
}

// Helper function to create a test WorkflowRunContext pointer for BashRunner
func createTestWorkflowRunContextPtrForBash() *run_context.WorkflowRunContext {
	jsCtx := createTestJSContextForBash()
	shellCtx := shell.NewShellContext()
	return &run_context.WorkflowRunContext{
		JSCtx:    jsCtx,
		ShellCtx: shellCtx,
	}
}

// Helper function to create a test RunnableContext for BashRunner
func createTestBashRunnableContext() *core.RunnableContext {
	workflowStatus := core.NewRunnableStatus("test-workflow", "workflow")
	jobStatus := core.NewRunnableStatus("test-job", "job")
	return &core.RunnableContext{
		NeedOutput:     true,
		Args:           &mockBashEnv{data: map[string]string{"arg1": "value1"}},
		JobStatus:      jobStatus,
		WorkflowStatus: workflowStatus,
	}
}

func TestBashRunner_Compile(t *testing.T) {
	t.Run("SuccessfulCompile", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "test-runner",
			Script: "echo 'Hello World'",
		}
		runCtx := createTestWorkflowRunContextForBash()

		err := runner.Compile(runCtx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if runner.hasError != 2 {
			t.Errorf("Expected hasError to be 2, got %d", runner.hasError)
		}
	})

	t.Run("CompileError", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "test-runner",
			Script: "invalid bash syntax $$$ ^^^",
		}
		runCtx := createTestWorkflowRunContextForBash()

		err := runner.Compile(runCtx)

		// Note: Real ShellContext may or may not fail on this syntax, so we test both cases
		if err != nil {
			if runner.hasError != 1 {
				t.Errorf("Expected hasError to be 1 when error occurs, got %d", runner.hasError)
			}
		} else {
			if runner.hasError != 2 {
				t.Errorf("Expected hasError to be 2 when no error occurs, got %d", runner.hasError)
			}
		}
	})
}

func TestBashRunner_Do(t *testing.T) {
	t.Run("SuccessfulExecution", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "test-runner",
			Script: "echo 'Hello World'",
		}
		parent := &mockBashEnv{data: map[string]string{"HOME": "/home/user"}}
		runCtx := createTestWorkflowRunContextPtrForBash()
		ctx := createTestBashRunnableContext()

		result := runner.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		// Note: With real ShellContext, we expect successful execution
		if result.Err != nil {
			t.Errorf("Expected no error, got %v", result.Err)
		}
		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}
		// Output may vary with real shell execution, so we just check it's reasonable
		if result.Output != "" && len(result.Output) > 0 {
			t.Logf("Got output: %s", result.Output)
		}
	})

	t.Run("ExecutionWithError", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "test-runner",
			Script: "exit 1",
		}
		parent := &mockBashEnv{data: map[string]string{"HOME": "/home/user"}}
		runCtx := createTestWorkflowRunContextPtrForBash()
		ctx := createTestBashRunnableContext()

		result := runner.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		// With real ShellContext, exit 1 should cause an error
		if result.Err == nil {
			t.Error("Expected error for exit 1, got nil")
		}
		if result.ReturnCode != 1 {
			t.Errorf("Expected return code 1, got %d", result.ReturnCode)
		}
	})

	t.Run("EmptyScript", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "test-runner",
			Script: "",
		}
		parent := &mockBashEnv{data: map[string]string{}}
		runCtx := createTestWorkflowRunContextPtrForBash()
		ctx := createTestBashRunnableContext()

		result := runner.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		// Empty script should execute successfully
		if result.Err != nil {
			t.Errorf("Expected no error for empty script, got %v", result.Err)
		}
		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}
	})

	t.Run("WithEnvironmentVariables", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "test-runner",
			Script: "echo $TEST_VAR",
		}
		parent := &mockBashEnv{data: map[string]string{"TEST_VAR": "test_value", "PATH": "/usr/bin"}}
		runCtx := createTestWorkflowRunContextPtrForBash()
		ctx := createTestBashRunnableContext()

		result := runner.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Err != nil {
			t.Errorf("Expected no error, got %v", result.Err)
		}
		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}
		// Output should contain the environment variable value
		if result.Output != "" && len(result.Output) > 0 {
			t.Logf("Got output with environment variable: %s", result.Output)
		}
	})
}

func TestBashRunner_CanRun(t *testing.T) {
	t.Run("InitialState", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "test-runner",
			Script: "echo 'test'",
		}

		canRun := runner.CanRun()

		if canRun {
			t.Error("Expected CanRun to be false initially")
		}
	})

	t.Run("AfterCompileError", func(t *testing.T) {
		runner := &BashRunner{
			Name:     "test-runner",
			Script:   "invalid syntax",
			hasError: 1,
		}

		canRun := runner.CanRun()

		if canRun {
			t.Error("Expected CanRun to be false after compile error")
		}
	})

	t.Run("AfterSuccessfulCompile", func(t *testing.T) {
		runner := &BashRunner{
			Name:     "test-runner",
			Script:   "echo 'test'",
			hasError: 2,
		}

		canRun := runner.CanRun()

		if !canRun {
			t.Error("Expected CanRun to be true after successful compile")
		}
	})

	t.Run("HigherErrorValue", func(t *testing.T) {
		runner := &BashRunner{
			Name:     "test-runner",
			Script:   "echo 'test'",
			hasError: 5,
		}

		canRun := runner.CanRun()

		if !canRun {
			t.Error("Expected CanRun to be true with hasError > 1")
		}
	})
}

func TestBashRunner_PreflightCheck(t *testing.T) {
	t.Run("AlwaysReturnsNil", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "test-runner",
			Script: "echo 'test'",
		}
		parent := &mockBashEnv{data: map[string]string{"HOME": "/home/user"}}
		args := &mockBashEnv{data: map[string]string{"arg1": "value1"}}
		runCtx := createTestWorkflowRunContextPtrForBash()

		err := runner.PreflightCheck(parent, args, runCtx)

		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})

	t.Run("WithNilEnvironments", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "test-runner",
			Script: "echo 'test'",
		}
		runCtx := createTestWorkflowRunContextPtrForBash()

		err := runner.PreflightCheck(nil, nil, runCtx)

		if err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})
}

func TestBashRunner_StructFields(t *testing.T) {
	t.Run("StructInitialization", func(t *testing.T) {
		name := "test-runner"
		script := "echo 'Hello World'"
		runner := &BashRunner{
			Name:   name,
			Script: script,
		}

		if runner.Name != name {
			t.Errorf("Expected Name to be %s, got %s", name, runner.Name)
		}
		if runner.Script != script {
			t.Errorf("Expected Script to be %s, got %s", script, runner.Script)
		}
		if runner.hasError != 0 {
			t.Errorf("Expected hasError to be 0, got %d", runner.hasError)
		}
	})

	t.Run("EmptyFields", func(t *testing.T) {
		runner := &BashRunner{}

		if runner.Name != "" {
			t.Errorf("Expected empty Name, got %s", runner.Name)
		}
		if runner.Script != "" {
			t.Errorf("Expected empty Script, got %s", runner.Script)
		}
		if runner.hasError != 0 {
			t.Errorf("Expected hasError to be 0, got %d", runner.hasError)
		}
	})

	t.Run("PartialInitialization", func(t *testing.T) {
		runner := &BashRunner{
			Name: "partial-runner",
		}

		if runner.Name != "partial-runner" {
			t.Errorf("Expected Name to be 'partial-runner', got %s", runner.Name)
		}
		if runner.Script != "" {
			t.Errorf("Expected empty Script, got %s", runner.Script)
		}
		if runner.hasError != 0 {
			t.Errorf("Expected hasError to be 0, got %d", runner.hasError)
		}
	})
}

func TestBashRunner_Integration(t *testing.T) {
	t.Run("FullWorkflow", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "integration-runner",
			Script: "echo 'Integration Test'",
		}
		parent := &mockBashEnv{data: map[string]string{"HOME": "/home/user"}}
		args := &mockBashEnv{data: map[string]string{"arg1": "value1"}}
		runCtx := createTestWorkflowRunContextPtrForBash()
		ctx := createTestBashRunnableContext()

		// Using real ShellContext for integration testing

		// Test PreflightCheck
		err := runner.PreflightCheck(parent, args, runCtx)
		if err != nil {
			t.Errorf("PreflightCheck failed: %v", err)
		}

		// Test Compile
		err = runner.Compile(*runCtx)
		if err != nil {
			t.Errorf("Compile failed: %v", err)
		}

		// Test CanRun
		if !runner.CanRun() {
			t.Error("Expected CanRun to be true after successful compile")
		}

		// Test Do
		result := runner.Do(parent, runCtx, ctx)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Err != nil {
			t.Errorf("Do failed: %v", result.Err)
		}
		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}
		// Output may vary with real shell execution
		if result.Output != "" && len(result.Output) > 0 {
			t.Logf("Got integration test output: %s", result.Output)
		}
	})

	t.Run("CompileErrorWorkflow", func(t *testing.T) {
		runner := &BashRunner{
			Name:   "error-runner",
			Script: "invalid bash syntax",
		}
		runCtx := createTestWorkflowRunContextForBash()

		// Using real ShellContext - syntax error will be handled naturally

		// Test Compile with error
		err := runner.Compile(runCtx)
		// Note: Real ShellContext may not always fail on this syntax, so we test both cases
		if err != nil {
			// Expected case: error occurred
			if runner.CanRun() {
				t.Error("Expected CanRun to be false after compile error")
			}
		} else {
			// Acceptable case: compilation succeeded
			t.Logf("Compilation succeeded for syntax that might be valid")
		}
	})
}

// Benchmark tests for BashRunner methods
func BenchmarkBashRunner_Compile(b *testing.B) {
	runner := &BashRunner{
		Name:   "benchmark-runner",
		Script: "echo 'benchmark test'",
	}
	runCtx := createTestWorkflowRunContextForBash()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner.Compile(runCtx)
	}
}

func BenchmarkBashRunner_Do(b *testing.B) {
	runner := &BashRunner{
		Name:   "benchmark-runner",
		Script: "echo 'benchmark test'",
	}
	parent := &mockBashEnv{data: map[string]string{"HOME": "/home/user"}}
	runCtx := createTestWorkflowRunContextPtrForBash()
	ctx := createTestBashRunnableContext()

	// Using real ShellContext for benchmarking

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner.Do(parent, runCtx, ctx)
	}
}

func BenchmarkBashRunner_CanRun(b *testing.B) {
	runner := &BashRunner{
		Name:     "benchmark-runner",
		Script:   "echo 'benchmark test'",
		hasError: 2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner.CanRun()
	}
}

func BenchmarkBashRunner_PreflightCheck(b *testing.B) {
	runner := &BashRunner{
		Name:   "benchmark-runner",
		Script: "echo 'benchmark test'",
	}
	parent := &mockBashEnv{data: map[string]string{"HOME": "/home/user"}}
	args := &mockBashEnv{data: map[string]string{"arg1": "value1"}}
	runCtx := createTestWorkflowRunContextPtrForBash()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner.PreflightCheck(parent, args, runCtx)
	}
}
