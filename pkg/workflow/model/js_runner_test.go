package workflow

import (
	"nadleeh/pkg/encrypt"
	"nadleeh/pkg/script"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"
	"testing"

	"github.com/zhaojunlucky/golib/pkg/env"
)

// Mock implementation of env.Env interface for JSRunner tests
type mockJSEnv struct {
	data map[string]string
}

// Ensure mockJSEnv implements env.Env interface
var _ env.Env = (*mockJSEnv)(nil)

func (m *mockJSEnv) Get(key string) string {
	if m.data == nil {
		return ""
	}
	return m.data[key]
}

func (m *mockJSEnv) Set(key, value string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
}

func (m *mockJSEnv) GetAll() map[string]string {
	if m.data == nil {
		return make(map[string]string)
	}
	return m.data
}

func (m *mockJSEnv) Contains(key string) bool {
	if m.data == nil {
		return false
	}
	_, exists := m.data[key]
	return exists
}

func (m *mockJSEnv) Expand(value string) string {
	return value
}

func (m *mockJSEnv) SetAll(data map[string]string) {
	m.data = data
}

// Helper function to create a test JSRunner
func createTestJSRunner(name, script string) *JSRunner {
	return &JSRunner{
		Name:     name,
		Script:   script,
		hasError: 0,
	}
}

// Helper function to create a real JSContext for testing
func createTestJSContextForJSRunner() script.JSContext {
	secCtx := encrypt.SecureContext{}
	return script.NewJSContext(&secCtx)
}

// Helper function to create a test WorkflowRunContext with real JSContext
func createTestWorkflowRunContextForJSRunner() run_context.WorkflowRunContext {
	jsCtx := createTestJSContextForJSRunner()
	return run_context.WorkflowRunContext{
		JSCtx: jsCtx,
	}
}

// Helper function to create a test WorkflowRunContext pointer with real JSContext
func createTestWorkflowRunContextPtr() *run_context.WorkflowRunContext {
	jsCtx := createTestJSContextForJSRunner()
	return &run_context.WorkflowRunContext{
		JSCtx: jsCtx,
	}
}

// Helper function to create a test RunnableContext
func createTestJSRunnableContext() *core.RunnableContext {
	return &core.RunnableContext{
		NeedOutput:     true,
		Args:           &mockJSEnv{data: map[string]string{"arg1": "value1"}},
		JobStatus:      &core.RunnableStatus{},
		WorkflowStatus: &core.RunnableStatus{},
	}
}

func TestJSRunner_Compile(t *testing.T) {
	t.Run("SuccessfulCompile", func(t *testing.T) {
		runner := createTestJSRunner("test-runner", "var x = 5; x + 10;")
		runCtx := createTestWorkflowRunContextForJSRunner()

		err := runner.Compile(runCtx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if runner.hasError != 2 {
			t.Errorf("Expected hasError to be 2 (success), got %d", runner.hasError)
		}
	})

	t.Run("CompileError", func(t *testing.T) {
		runner := createTestJSRunner("test-runner", "invalid js syntax {{{")
		runCtx := createTestWorkflowRunContextForJSRunner()

		err := runner.Compile(runCtx)

		if err == nil {
			t.Error("Expected compile error for invalid syntax")
		}
		if runner.hasError != 1 {
			t.Errorf("Expected hasError to be 1 (error), got %d", runner.hasError)
		}
	})

	t.Run("EmptyScript", func(t *testing.T) {
		runner := createTestJSRunner("empty-runner", "")
		runCtx := createTestWorkflowRunContextForJSRunner()

		err := runner.Compile(runCtx)

		if err != nil {
			t.Errorf("Expected no error for empty script, got %v", err)
		}
		if runner.hasError != 2 {
			t.Errorf("Expected hasError to be 2 (success), got %d", runner.hasError)
		}
	})

	t.Run("SimpleExpression", func(t *testing.T) {
		runner := createTestJSRunner("expr-runner", "2 + 3")
		runCtx := createTestWorkflowRunContextForJSRunner()

		err := runner.Compile(runCtx)

		if err != nil {
			t.Errorf("Expected no error for simple expression, got %v", err)
		}
		if runner.hasError != 2 {
			t.Errorf("Expected hasError to be 2 (success), got %d", runner.hasError)
		}
	})
}

func TestJSRunner_Do(t *testing.T) {
	t.Run("SuccessfulExecution", func(t *testing.T) {
		runner := createTestJSRunner("test-runner", "5 + 3")
		parent := &mockJSEnv{data: map[string]string{"parent": "value"}}
		runCtx := createTestWorkflowRunContextPtr()
		ctx := createTestJSRunnableContext()

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
		// The output should contain the result of the JavaScript execution
		if result.Output == "" {
			t.Error("Expected non-empty output from JavaScript execution")
		}
	})

	t.Run("ExecutionWithError", func(t *testing.T) {
		runner := createTestJSRunner("error-runner", "throw new Error('test error');")
		parent := &mockJSEnv{data: map[string]string{"parent": "value"}}
		runCtx := createTestWorkflowRunContextPtr()
		ctx := createTestJSRunnableContext()

		result := runner.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Err == nil {
			t.Error("Expected error from JavaScript execution")
		}
		if result.ReturnCode == 0 {
			t.Error("Expected non-zero return code for error")
		}
	})

	t.Run("SimpleCalculation", func(t *testing.T) {
		runner := createTestJSRunner("calc-runner", "var x = 10; var y = 20; x + y;")
		parent := &mockJSEnv{data: map[string]string{"parent": "value"}}
		runCtx := createTestWorkflowRunContextPtr()
		ctx := createTestJSRunnableContext()

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
	})

	t.Run("EmptyScript", func(t *testing.T) {
		runner := createTestJSRunner("empty-runner", "")
		parent := &mockJSEnv{data: map[string]string{"parent": "value"}}
		runCtx := createTestWorkflowRunContextPtr()
		ctx := createTestJSRunnableContext()

		result := runner.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Err != nil {
			t.Errorf("Expected no error for empty script, got %v", result.Err)
		}
		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}
	})
}

func TestJSRunner_CanRun(t *testing.T) {
	t.Run("InitialState", func(t *testing.T) {
		runner := createTestJSRunner("test-runner", "var x = 5;")

		canRun := runner.CanRun()

		if canRun != false {
			t.Errorf("Expected CanRun to be false initially (hasError=0), got %v", canRun)
		}
	})

	t.Run("AfterCompileError", func(t *testing.T) {
		runner := createTestJSRunner("test-runner", "var x = 5;")
		runner.hasError = 1 // Simulate compile error

		canRun := runner.CanRun()

		if canRun != false {
			t.Errorf("Expected CanRun to be false after compile error (hasError=1), got %v", canRun)
		}
	})

	t.Run("AfterSuccessfulCompile", func(t *testing.T) {
		runner := createTestJSRunner("test-runner", "var x = 5;")
		runner.hasError = 2 // Simulate successful compile

		canRun := runner.CanRun()

		if canRun != true {
			t.Errorf("Expected CanRun to be true after successful compile (hasError=2), got %v", canRun)
		}
	})

	t.Run("HigherErrorValue", func(t *testing.T) {
		runner := createTestJSRunner("test-runner", "var x = 5;")
		runner.hasError = 5 // Test edge case

		canRun := runner.CanRun()

		if canRun != true {
			t.Errorf("Expected CanRun to be true for hasError > 1 (hasError=5), got %v", canRun)
		}
	})
}

func TestJSRunner_PreflightCheck(t *testing.T) {
	t.Run("AlwaysReturnsNil", func(t *testing.T) {
		runner := createTestJSRunner("test-runner", "var x = 5;")
		parent := &mockJSEnv{data: map[string]string{"parent": "value"}}
		args := &mockJSEnv{data: map[string]string{"arg": "value"}}
		runCtx := createTestWorkflowRunContextPtr()

		err := runner.PreflightCheck(parent, args, runCtx)

		if err != nil {
			t.Errorf("Expected PreflightCheck to always return nil, got %v", err)
		}
	})

	t.Run("WithNilEnvironments", func(t *testing.T) {
		runner := createTestJSRunner("test-runner", "var x = 5;")
		runCtx := createTestWorkflowRunContextPtr()

		err := runner.PreflightCheck(nil, nil, runCtx)

		if err != nil {
			t.Errorf("Expected PreflightCheck to return nil even with nil envs, got %v", err)
		}
	})
}

func TestJSRunner_StructFields(t *testing.T) {
	t.Run("StructInitialization", func(t *testing.T) {
		name := "test-js-runner"
		script := "var x = 5; x + 10;"

		runner := &JSRunner{
			Name:     name,
			Script:   script,
			hasError: 0,
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
		runner := &JSRunner{
			Name:     "",
			Script:   "",
			hasError: 0,
		}

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
}

// Integration tests
func TestJSRunner_Integration(t *testing.T) {
	t.Run("FullWorkflow", func(t *testing.T) {
		runner := createTestJSRunner("integration-runner", "var result = 5 + 3; result;")
		runCtx := createTestWorkflowRunContextForJSRunner()
		runCtxPtr := createTestWorkflowRunContextPtr()
		parent := &mockJSEnv{data: map[string]string{"parent": "value"}}
		args := &mockJSEnv{data: map[string]string{"arg": "value"}}
		ctx := createTestJSRunnableContext()

		// Test Compile
		err := runner.Compile(runCtx)
		if err != nil {
			t.Errorf("Compile failed: %v", err)
		}

		// Test CanRun after successful compile
		canRun := runner.CanRun()
		if !canRun {
			t.Error("Expected CanRun to be true after successful compile")
		}

		// Test PreflightCheck
		err = runner.PreflightCheck(parent, args, runCtxPtr)
		if err != nil {
			t.Errorf("PreflightCheck failed: %v", err)
		}

		// Test Do
		result := runner.Do(parent, runCtxPtr, ctx)
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Err != nil {
			t.Errorf("Expected no error in result, got %v", result.Err)
		}
		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}
	})

	t.Run("CompileErrorWorkflow", func(t *testing.T) {
		runner := createTestJSRunner("error-runner", "invalid syntax {{{")
		runCtx := createTestWorkflowRunContextForJSRunner()

		// Test Compile with error
		err := runner.Compile(runCtx)
		if err == nil {
			t.Error("Expected compile error for invalid syntax")
		}

		// Test CanRun after compile error
		canRun := runner.CanRun()
		if canRun {
			t.Error("Expected CanRun to be false after compile error")
		}

		// Verify hasError state
		if runner.hasError != 1 {
			t.Errorf("Expected hasError to be 1 after compile error, got %d", runner.hasError)
		}
	})
}

// Benchmark tests
func BenchmarkJSRunner_Compile(b *testing.B) {
	runner := createTestJSRunner("benchmark-runner", "var x = 5; x + 10;")
	runCtx := createTestWorkflowRunContextForJSRunner()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := runner.Compile(runCtx)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkJSRunner_Do(b *testing.B) {
	runner := createTestJSRunner("benchmark-runner", "var x = 5; x + 10;")
	parent := &mockJSEnv{data: map[string]string{"parent": "value"}}
	runCtx := createTestWorkflowRunContextPtr()
	ctx := createTestJSRunnableContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := runner.Do(parent, runCtx, ctx)
		if result == nil {
			b.Fatal("Unexpected nil result")
		}
	}
}

func BenchmarkJSRunner_CanRun(b *testing.B) {
	runner := createTestJSRunner("benchmark-runner", "var x = 5; x + 10;")
	runner.hasError = 2 // Set to successful compile state

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runner.CanRun()
	}
}

func BenchmarkJSRunner_PreflightCheck(b *testing.B) {
	runner := createTestJSRunner("benchmark-runner", "var x = 5; x + 10;")
	parent := &mockJSEnv{data: map[string]string{"parent": "value"}}
	args := &mockJSEnv{data: map[string]string{"arg": "value"}}
	runCtx := createTestWorkflowRunContextPtr()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := runner.PreflightCheck(parent, args, runCtx)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
