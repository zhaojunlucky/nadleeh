package workflow

import (
	"errors"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"
	"testing"

	"github.com/zhaojunlucky/golib/pkg/env"
)

// Mock implementation of plugin.Plugin interface
type mockPlugin struct {
	compileError      error
	doResult          *core.RunnableResult
	canRunResult      bool
	preflightError    error
	resolveError      error
	name              string
	compileCalled     bool
	doCalled          bool
	canRunCalled      bool
	preflightCalled   bool
	resolveCalled     bool
}

func (m *mockPlugin) Compile(runCtx run_context.WorkflowRunContext) error {
	m.compileCalled = true
	return m.compileError
}

func (m *mockPlugin) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	m.doCalled = true
	return m.doResult
}

func (m *mockPlugin) CanRun() bool {
	m.canRunCalled = true
	return m.canRunResult
}

func (m *mockPlugin) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {
	m.preflightCalled = true
	return m.preflightError
}

func (m *mockPlugin) Resolve() error {
	m.resolveCalled = true
	return m.resolveError
}

func (m *mockPlugin) GetName() string {
	return m.name
}

// Mock implementation of env.Env interface for PluginRunner tests
type mockPluginEnv struct {
	data map[string]string
}

func (m *mockPluginEnv) Get(key string) string {
	if m.data == nil {
		return ""
	}
	return m.data[key]
}

func (m *mockPluginEnv) Set(key, value string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
}

func (m *mockPluginEnv) GetAll() map[string]string {
	if m.data == nil {
		return make(map[string]string)
	}
	return m.data
}

func (m *mockPluginEnv) Contains(key string) bool {
	if m.data == nil {
		return false
	}
	_, exists := m.data[key]
	return exists
}

func (m *mockPluginEnv) Expand(value string) string {
	return value
}

func (m *mockPluginEnv) SetAll(data map[string]string) {
	m.data = data
}

// Helper function to create a test PluginRunner
func createTestPluginRunner(plug *mockPlugin) *PluginRunner {
	return &PluginRunner{
		Config:   map[string]string{"test": "config"},
		StepName: "test-step",
		plug:     plug,
	}
}

// Helper function to create a test WorkflowRunContext
func createTestWorkflowRunContext() run_context.WorkflowRunContext {
	// Return a minimal implementation for testing
	return run_context.WorkflowRunContext{}
}

// Helper function to create a test RunnableContext
func createTestRunnableContext() *core.RunnableContext {
	return &core.RunnableContext{
		NeedOutput:     true,
		Args:           &mockPluginEnv{data: map[string]string{"arg1": "value1"}},
		JobStatus:      &core.RunnableStatus{},
		WorkflowStatus: &core.RunnableStatus{},
	}
}

func TestPluginRunner_Compile(t *testing.T) {
	t.Run("SuccessfulCompile", func(t *testing.T) {
		mockPlug := &mockPlugin{
			compileError: nil,
		}
		runner := createTestPluginRunner(mockPlug)
		runCtx := createTestWorkflowRunContext()

		err := runner.Compile(runCtx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !mockPlug.compileCalled {
			t.Error("Expected Compile to be called on plugin")
		}
	})

	t.Run("CompileError", func(t *testing.T) {
		expectedError := errors.New("compile failed")
		mockPlug := &mockPlugin{
			compileError: expectedError,
		}
		runner := createTestPluginRunner(mockPlug)
		runCtx := createTestWorkflowRunContext()

		err := runner.Compile(runCtx)

		if err != expectedError {
			t.Errorf("Expected error %v, got %v", expectedError, err)
		}
		if !mockPlug.compileCalled {
			t.Error("Expected Compile to be called on plugin")
		}
	})
}

func TestPluginRunner_Do(t *testing.T) {
	t.Run("SuccessfulDo", func(t *testing.T) {
		expectedResult := &core.RunnableResult{
			Err:        nil,
			ReturnCode: 0,
			Output:     "test output",
		}
		mockPlug := &mockPlugin{
			doResult: expectedResult,
		}
		runner := createTestPluginRunner(mockPlug)
		parent := &mockPluginEnv{data: map[string]string{"parent": "value"}}
		runCtx := &run_context.WorkflowRunContext{}
		ctx := createTestRunnableContext()

		result := runner.Do(parent, runCtx, ctx)

		if result != expectedResult {
			t.Errorf("Expected result %v, got %v", expectedResult, result)
		}
		if !mockPlug.doCalled {
			t.Error("Expected Do to be called on plugin")
		}
	})

	t.Run("DoWithError", func(t *testing.T) {
		expectedResult := &core.RunnableResult{
			Err:        errors.New("execution failed"),
			ReturnCode: 1,
			Output:     "",
		}
		mockPlug := &mockPlugin{
			doResult: expectedResult,
		}
		runner := createTestPluginRunner(mockPlug)
		parent := &mockPluginEnv{data: map[string]string{"parent": "value"}}
		runCtx := &run_context.WorkflowRunContext{}
		ctx := createTestRunnableContext()

		result := runner.Do(parent, runCtx, ctx)

		if result != expectedResult {
			t.Errorf("Expected result %v, got %v", expectedResult, result)
		}
		if !mockPlug.doCalled {
			t.Error("Expected Do to be called on plugin")
		}
	})

	t.Run("DoWithNilResult", func(t *testing.T) {
		mockPlug := &mockPlugin{
			doResult: nil,
		}
		runner := createTestPluginRunner(mockPlug)
		parent := &mockPluginEnv{data: map[string]string{"parent": "value"}}
		runCtx := &run_context.WorkflowRunContext{}
		ctx := createTestRunnableContext()

		result := runner.Do(parent, runCtx, ctx)

		if result != nil {
			t.Errorf("Expected nil result, got %v", result)
		}
		if !mockPlug.doCalled {
			t.Error("Expected Do to be called on plugin")
		}
	})
}

func TestPluginRunner_CanRun(t *testing.T) {
	t.Run("PluginCanRun_True", func(t *testing.T) {
		mockPlug := &mockPlugin{
			canRunResult: true,
		}
		runner := createTestPluginRunner(mockPlug)

		result := runner.CanRun()

		// Note: PluginRunner.CanRun() returns !p.plug.CanRun()
		if result != false {
			t.Errorf("Expected false (negated), got %v", result)
		}
		if !mockPlug.canRunCalled {
			t.Error("Expected CanRun to be called on plugin")
		}
	})

	t.Run("PluginCanRun_False", func(t *testing.T) {
		mockPlug := &mockPlugin{
			canRunResult: false,
		}
		runner := createTestPluginRunner(mockPlug)

		result := runner.CanRun()

		// Note: PluginRunner.CanRun() returns !p.plug.CanRun()
		if result != true {
			t.Errorf("Expected true (negated), got %v", result)
		}
		if !mockPlug.canRunCalled {
			t.Error("Expected CanRun to be called on plugin")
		}
	})
}

func TestPluginRunner_PreflightCheck(t *testing.T) {
	t.Run("SuccessfulPreflightCheck", func(t *testing.T) {
		mockPlug := &mockPlugin{
			preflightError: nil,
		}
		runner := createTestPluginRunner(mockPlug)
		parent := &mockPluginEnv{data: map[string]string{"parent": "value"}}
		args := &mockPluginEnv{data: map[string]string{"arg": "value"}}
		runCtx := &run_context.WorkflowRunContext{}

		err := runner.PreflightCheck(parent, args, runCtx)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !mockPlug.preflightCalled {
			t.Error("Expected PreflightCheck to be called on plugin")
		}
	})

	t.Run("PreflightCheckError", func(t *testing.T) {
		expectedError := errors.New("preflight check failed")
		mockPlug := &mockPlugin{
			preflightError: expectedError,
		}
		runner := createTestPluginRunner(mockPlug)
		parent := &mockPluginEnv{data: map[string]string{"parent": "value"}}
		args := &mockPluginEnv{data: map[string]string{"arg": "value"}}
		runCtx := &run_context.WorkflowRunContext{}

		err := runner.PreflightCheck(parent, args, runCtx)

		if err != expectedError {
			t.Errorf("Expected error %v, got %v", expectedError, err)
		}
		if !mockPlug.preflightCalled {
			t.Error("Expected PreflightCheck to be called on plugin")
		}
	})
}

func TestPluginRunner_StructFields(t *testing.T) {
	t.Run("StructInitialization", func(t *testing.T) {
		config := map[string]string{"key1": "value1", "key2": "value2"}
		stepName := "test-step-name"
		mockPlug := &mockPlugin{name: "test-plugin"}

		runner := &PluginRunner{
			Config:   config,
			StepName: stepName,
			plug:     mockPlug,
		}

		if runner.Config == nil {
			t.Error("Expected Config to be set")
		}
		if len(runner.Config) != 2 {
			t.Errorf("Expected Config to have 2 items, got %d", len(runner.Config))
		}
		if runner.Config["key1"] != "value1" {
			t.Errorf("Expected Config['key1'] to be 'value1', got %s", runner.Config["key1"])
		}
		if runner.StepName != stepName {
			t.Errorf("Expected StepName to be %s, got %s", stepName, runner.StepName)
		}
		if runner.plug != mockPlug {
			t.Error("Expected plug to be set to mockPlug")
		}
	})

	t.Run("EmptyConfig", func(t *testing.T) {
		runner := &PluginRunner{
			Config:   map[string]string{},
			StepName: "empty-config-step",
			plug:     &mockPlugin{},
		}

		if runner.Config == nil {
			t.Error("Expected Config to be initialized")
		}
		if len(runner.Config) != 0 {
			t.Errorf("Expected empty Config, got %d items", len(runner.Config))
		}
	})

	t.Run("NilConfig", func(t *testing.T) {
		runner := &PluginRunner{
			Config:   nil,
			StepName: "nil-config-step",
			plug:     &mockPlugin{},
		}

		if runner.Config != nil {
			t.Error("Expected Config to be nil")
		}
	})
}

// Integration tests
func TestPluginRunner_Integration(t *testing.T) {
	t.Run("FullWorkflow", func(t *testing.T) {
		mockPlug := &mockPlugin{
			compileError:   nil,
			doResult:       &core.RunnableResult{Err: nil, ReturnCode: 0, Output: "success"},
			canRunResult:   true,
			preflightError: nil,
			name:           "integration-plugin",
		}
		runner := createTestPluginRunner(mockPlug)
		runCtx := createTestWorkflowRunContext()
		parent := &mockPluginEnv{data: map[string]string{"parent": "value"}}
		args := &mockPluginEnv{data: map[string]string{"arg": "value"}}
		ctx := createTestRunnableContext()

		// Test Compile
		err := runner.Compile(runCtx)
		if err != nil {
			t.Errorf("Compile failed: %v", err)
		}

		// Test PreflightCheck
		err = runner.PreflightCheck(parent, args, &run_context.WorkflowRunContext{})
		if err != nil {
			t.Errorf("PreflightCheck failed: %v", err)
		}

		// Test CanRun
		canRun := runner.CanRun()
		if canRun != false { // Remember: PluginRunner negates the result
			t.Errorf("Expected CanRun to be false (negated), got %v", canRun)
		}

		// Test Do
		result := runner.Do(parent, &run_context.WorkflowRunContext{}, ctx)
		if result == nil {
			t.Error("Expected non-nil result")
			return
		}
		if result.Err != nil {
			t.Errorf("Expected no error in result, got %v", result.Err)
		}
		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}

		// Verify all methods were called
		if !mockPlug.compileCalled {
			t.Error("Expected Compile to be called")
		}
		if !mockPlug.preflightCalled {
			t.Error("Expected PreflightCheck to be called")
		}
		if !mockPlug.canRunCalled {
			t.Error("Expected CanRun to be called")
		}
		if !mockPlug.doCalled {
			t.Error("Expected Do to be called")
		}
	})
}

// Benchmark tests
func BenchmarkPluginRunner_Compile(b *testing.B) {
	mockPlug := &mockPlugin{compileError: nil}
	runner := createTestPluginRunner(mockPlug)
	runCtx := createTestWorkflowRunContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := runner.Compile(runCtx)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkPluginRunner_Do(b *testing.B) {
	mockPlug := &mockPlugin{
		doResult: &core.RunnableResult{Err: nil, ReturnCode: 0, Output: "benchmark"},
	}
	runner := createTestPluginRunner(mockPlug)
	parent := &mockPluginEnv{data: map[string]string{"parent": "value"}}
	runCtx := &run_context.WorkflowRunContext{}
	ctx := createTestRunnableContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := runner.Do(parent, runCtx, ctx)
		if result == nil {
			b.Fatal("Unexpected nil result")
		}
	}
}

func BenchmarkPluginRunner_CanRun(b *testing.B) {
	mockPlug := &mockPlugin{canRunResult: true}
	runner := createTestPluginRunner(mockPlug)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runner.CanRun()
	}
}

func BenchmarkPluginRunner_PreflightCheck(b *testing.B) {
	mockPlug := &mockPlugin{preflightError: nil}
	runner := createTestPluginRunner(mockPlug)
	parent := &mockPluginEnv{data: map[string]string{"parent": "value"}}
	args := &mockPluginEnv{data: map[string]string{"arg": "value"}}
	runCtx := &run_context.WorkflowRunContext{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := runner.PreflightCheck(parent, args, runCtx)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
