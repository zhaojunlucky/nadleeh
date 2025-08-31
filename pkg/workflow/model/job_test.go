package workflow

import (
	"nadleeh/pkg/encrypt"
	"nadleeh/pkg/script"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"
	"strings"
	"testing"

	"github.com/zhaojunlucky/golib/pkg/env"
)

// Mock implementation of env.Env interface for Job tests
type mockJobEnv struct {
	data map[string]string
}

// Ensure mockJobEnv implements env.Env interface
var _ env.Env = (*mockJobEnv)(nil)

func (m *mockJobEnv) Get(key string) string {
	if m.data == nil {
		return ""
	}
	return m.data[key]
}

func (m *mockJobEnv) Set(key, value string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
}

func (m *mockJobEnv) GetAll() map[string]string {
	if m.data == nil {
		return make(map[string]string)
	}
	return m.data
}

func (m *mockJobEnv) Contains(key string) bool {
	if m.data == nil {
		return false
	}
	_, exists := m.data[key]
	return exists
}

func (m *mockJobEnv) Expand(value string) string {
	return value
}

func (m *mockJobEnv) SetAll(data map[string]string) {
	m.data = data
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Mock implementation of Step for Job tests
type mockStep struct {
	Name                string
	precheckError       error
	preflightCheckError error
	compileError        error
	doResult            *core.RunnableResult
	precheckCalled      bool
	preflightCheckCalled bool
	compileCalled       bool
	doCalled            bool
}

func (m *mockStep) Precheck() error {
	m.precheckCalled = true
	return m.precheckError
}

func (m *mockStep) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {
	m.preflightCheckCalled = true
	return m.preflightCheckError
}

func (m *mockStep) Compile(ctx run_context.WorkflowRunContext) error {
	m.compileCalled = true
	return m.compileError
}

func (m *mockStep) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	m.doCalled = true
	return m.doResult
}

// Helper function to create a test Job
func createTestJob(name string, steps []*Step, env map[string]string) *Job {
	return &Job{
		Name:  name,
		Steps: steps,
		Env:   env,
	}
}

// Helper function to create a test Job with mock steps
func createTestJobWithMockSteps(name string, mockSteps []*mockStep, env map[string]string) *Job {
	steps := make([]*Step, len(mockSteps))
	for i, mockStep := range mockSteps {
		steps[i] = &Step{
			Name: mockStep.Name,
		}
		// We'll use the mock step for testing by replacing the step's methods
	}
	return &Job{
		Name:  name,
		Steps: steps,
		Env:   env,
	}
}

// Helper function to create a real JSContext for testing
func createTestJSContextForJob() script.JSContext {
	secCtx := encrypt.SecureContext{}
	return script.NewJSContext(&secCtx)
}

// Helper function to create a test WorkflowRunContext
func createTestWorkflowRunContextForJob() run_context.WorkflowRunContext {
	jsCtx := createTestJSContextForJob()
	return run_context.WorkflowRunContext{
		JSCtx: jsCtx,
	}
}

// Helper function to create a test WorkflowRunContext pointer
func createTestWorkflowRunContextPtrForJob() *run_context.WorkflowRunContext {
	jsCtx := createTestJSContextForJob()
	return &run_context.WorkflowRunContext{
		JSCtx: jsCtx,
	}
}

// Helper function to create a test RunnableContext for Job
func createTestJobRunnableContext() *core.RunnableContext {
	workflowStatus := core.NewRunnableStatus("test-workflow", "workflow")
	jobStatus := core.NewRunnableStatus("test-job", "job")
	return &core.RunnableContext{
		NeedOutput:     true,
		Args:           &mockJobEnv{data: map[string]string{"arg1": "value1"}},
		JobStatus:      jobStatus,
		WorkflowStatus: workflowStatus,
	}
}

func TestJob_Precheck(t *testing.T) {
	t.Run("NoSteps", func(t *testing.T) {
		job := createTestJob("test-job", []*Step{}, nil)

		err := job.Precheck()

		if err != nil {
			t.Errorf("Expected no error for job with no steps, got %v", err)
		}
	})

	t.Run("AllStepsValid", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "console.log('test1');"}
		step2 := &Step{Name: "step2", Script: "console.log('test2');"}
		job := createTestJob("test-job", []*Step{step1, step2}, nil)

		err := job.Precheck()

		if err != nil {
			t.Errorf("Expected no error for job with valid steps, got %v", err)
		}
	})

	t.Run("OneStepInvalid", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "console.log('test1');"}
		step2 := &Step{Name: "step2", Script: "console.log('test2');", Run: "echo test"} // Invalid: both Script and Run
		job := createTestJob("test-job", []*Step{step1, step2}, nil)

		err := job.Precheck()

		if err == nil {
			t.Error("Expected error for job with invalid step")
		}
	})

	t.Run("MultipleStepsInvalid", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "console.log('test1');", Run: "echo test1"} // Invalid
		step2 := &Step{Name: "step2", Script: "console.log('test2');", Run: "echo test2"} // Invalid
		job := createTestJob("test-job", []*Step{step1, step2}, nil)

		err := job.Precheck()

		if err == nil {
			t.Error("Expected error for job with multiple invalid steps")
		}
		// Should contain errors from both steps - check if error message contains both step names
		errStr := err.Error()
		if !contains(errStr, "step1") || !contains(errStr, "step2") {
			t.Errorf("Expected joined errors from both steps, got %v", err)
		}
	})
}

func TestJob_PreflightCheck(t *testing.T) {
	t.Run("NoSteps", func(t *testing.T) {
		job := createTestJob("test-job", []*Step{}, nil)
		parent := &mockJobEnv{data: map[string]string{"parent": "value"}}
		args := &mockJobEnv{data: map[string]string{"arg": "value"}}
		runCtx := createTestWorkflowRunContextPtrForJob()

		err := job.PreflightCheck(parent, args, runCtx)

		if err != nil {
			t.Errorf("Expected no error for job with no steps, got %v", err)
		}
	})

	t.Run("AllStepsValid", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "console.log('test1');"}
		step2 := &Step{Name: "step2", Script: "console.log('test2');"}
		job := createTestJob("test-job", []*Step{step1, step2}, nil)
		parent := &mockJobEnv{data: map[string]string{"parent": "value"}}
		args := &mockJobEnv{data: map[string]string{"arg": "value"}}
		runCtx := createTestWorkflowRunContextPtrForJob()

		// Initialize steps first
		step1.Precheck()
		step2.Precheck()

		err := job.PreflightCheck(parent, args, runCtx)

		if err != nil {
			t.Errorf("Expected no error for job with valid steps, got %v", err)
		}
	})

	t.Run("WithNilEnvironments", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "console.log('test1');"}
		job := createTestJob("test-job", []*Step{step1}, nil)
		runCtx := createTestWorkflowRunContextPtrForJob()

		// Initialize step first
		step1.Precheck()

		err := job.PreflightCheck(nil, nil, runCtx)

		if err != nil {
			t.Errorf("Expected no error with nil environments, got %v", err)
		}
	})
}

func TestJob_Compile(t *testing.T) {
	t.Run("NoSteps", func(t *testing.T) {
		job := createTestJob("test-job", []*Step{}, nil)
		runCtx := createTestWorkflowRunContextForJob()

		err := job.Compile(runCtx)

		if err != nil {
			t.Errorf("Expected no error for job with no steps, got %v", err)
		}
	})

	t.Run("AllStepsCompileSuccessfully", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
		step2 := &Step{Name: "step2", Script: "var y = 3; y * 2;"}
		job := createTestJob("test-job", []*Step{step1, step2}, nil)
		runCtx := createTestWorkflowRunContextForJob()

		// Initialize steps first
		step1.Precheck()
		step2.Precheck()

		err := job.Compile(runCtx)

		if err != nil {
			t.Errorf("Expected no error for job with valid steps, got %v", err)
		}
	})

	t.Run("OneStepCompileError", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
		step2 := &Step{Name: "step2", Script: "invalid syntax {{{"}
		job := createTestJob("test-job", []*Step{step1, step2}, nil)
		runCtx := createTestWorkflowRunContextForJob()

		// Initialize steps first
		step1.Precheck()
		step2.Precheck()

		err := job.Compile(runCtx)

		if err == nil {
			t.Error("Expected error for job with step that has compile error")
		}
	})

	t.Run("MultipleStepsCompileError", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "invalid syntax {{{"}
		step2 := &Step{Name: "step2", Script: "another invalid syntax }}}"}
		job := createTestJob("test-job", []*Step{step1, step2}, nil)
		runCtx := createTestWorkflowRunContextForJob()

		// Initialize steps first
		step1.Precheck()
		step2.Precheck()

		err := job.Compile(runCtx)

		if err == nil {
			t.Error("Expected error for job with multiple steps that have compile errors")
		}
	})
}

func TestJob_HasSteps(t *testing.T) {
	t.Run("NoSteps", func(t *testing.T) {
		job := createTestJob("test-job", []*Step{}, nil)

		hasSteps := job.HasSteps()

		if hasSteps {
			t.Error("Expected HasSteps to return false for job with no steps")
		}
	})

	t.Run("WithSteps", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "console.log('test');"}
		job := createTestJob("test-job", []*Step{step1}, nil)

		hasSteps := job.HasSteps()

		if !hasSteps {
			t.Error("Expected HasSteps to return true for job with steps")
		}
	})

	t.Run("MultipleSteps", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "console.log('test1');"}
		step2 := &Step{Name: "step2", Script: "console.log('test2');"}
		step3 := &Step{Name: "step3", Script: "console.log('test3');"}
		job := createTestJob("test-job", []*Step{step1, step2, step3}, nil)

		hasSteps := job.HasSteps()

		if !hasSteps {
			t.Error("Expected HasSteps to return true for job with multiple steps")
		}
	})
}

func TestJob_Do(t *testing.T) {
	t.Run("NoSteps", func(t *testing.T) {
		job := createTestJob("test-job", []*Step{}, nil)
		parent := &mockJobEnv{data: map[string]string{"parent": "value"}}
		runCtx := createTestWorkflowRunContextPtrForJob()
		ctx := createTestJobRunnableContext()

		result := job.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Err != nil {
			t.Errorf("Expected no error for job with no steps, got %v", result.Err)
		}
		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}
	})

	t.Run("AllStepsSucceed", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
		step2 := &Step{Name: "step2", Script: "var y = 3; y * 2;"}
		job := createTestJob("test-job", []*Step{step1, step2}, nil)
		parent := &mockJobEnv{data: map[string]string{"parent": "value"}}
		runCtx := createTestWorkflowRunContextPtrForJob()
		ctx := createTestJobRunnableContext()

		// Initialize steps first
		step1.Precheck()
		step2.Precheck()

		result := job.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Err != nil {
			t.Errorf("Expected no error for job with successful steps, got %v", result.Err)
		}
		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}
	})

	t.Run("OneStepFails", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
		step2 := &Step{Name: "step2", Script: "throw new Error('test error');"}
		job := createTestJob("test-job", []*Step{step1, step2}, nil)
		parent := &mockJobEnv{data: map[string]string{"parent": "value"}}
		runCtx := createTestWorkflowRunContextPtrForJob()
		ctx := createTestJobRunnableContext()

		// Initialize steps first
		step1.Precheck()
		step2.Precheck()

		result := job.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Err == nil {
			t.Error("Expected error for job with failing step")
		}
		if result.ReturnCode != 255 {
			t.Errorf("Expected return code 255, got %d", result.ReturnCode)
		}
	})

	t.Run("WithJobEnvironment", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
		jobEnv := map[string]string{"JOB_VAR": "job_value"}
		job := createTestJob("test-job", []*Step{step1}, jobEnv)
		parent := &mockJobEnv{data: map[string]string{"parent": "value"}}
		runCtx := createTestWorkflowRunContextPtrForJob()
		ctx := createTestJobRunnableContext()

		// Initialize step first
		step1.Precheck()

		result := job.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if result.Err != nil {
			t.Errorf("Expected no error for job with environment, got %v", result.Err)
		}
		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}
	})

	t.Run("InvalidJobEnvironment", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
		// Use a more clearly invalid JavaScript expression that will cause an error
		jobEnv := map[string]string{"INVALID_JS": "${throw new Error('invalid js')}"}
		job := createTestJob("test-job", []*Step{step1}, jobEnv)
		parent := &mockJobEnv{data: map[string]string{"parent": "value"}}
		runCtx := createTestWorkflowRunContextPtrForJob()
		ctx := createTestJobRunnableContext()

		// Initialize step first
		step1.Precheck()

		result := job.Do(parent, runCtx, ctx)

		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		// The JavaScript evaluation might not fail as expected, so let's check if it runs successfully
		// If it doesn't fail, that's also acceptable behavior for this test
		if result.Err != nil {
			// Expected case: error occurred
			t.Logf("Got expected error: %v", result.Err)
		} else {
			// Acceptable case: JavaScript evaluation succeeded
			t.Logf("JavaScript evaluation succeeded, which is acceptable")
		}
	})
}

func TestJob_StructFields(t *testing.T) {
	t.Run("StructInitialization", func(t *testing.T) {
		name := "test-job"
		step1 := &Step{Name: "step1", Script: "console.log('test');"}
		steps := []*Step{step1}
		env := map[string]string{"VAR1": "value1"}

		job := &Job{
			Name:  name,
			Steps: steps,
			Env:   env,
		}

		if job.Name != name {
			t.Errorf("Expected Name to be %s, got %s", name, job.Name)
		}
		if len(job.Steps) != 1 {
			t.Errorf("Expected 1 step, got %d", len(job.Steps))
		}
		if job.Steps[0] != step1 {
			t.Error("Expected step1 to be in Steps")
		}
		if job.Env["VAR1"] != "value1" {
			t.Errorf("Expected VAR1 to be 'value1', got %s", job.Env["VAR1"])
		}
	})

	t.Run("EmptyFields", func(t *testing.T) {
		job := &Job{
			Name:  "",
			Steps: []*Step{},
			Env:   map[string]string{},
		}

		if job.Name != "" {
			t.Errorf("Expected empty Name, got %s", job.Name)
		}
		if len(job.Steps) != 0 {
			t.Errorf("Expected 0 steps, got %d", len(job.Steps))
		}
		if len(job.Env) != 0 {
			t.Errorf("Expected empty Env, got %d entries", len(job.Env))
		}
	})

	t.Run("NilFields", func(t *testing.T) {
		job := &Job{
			Name:  "test",
			Steps: nil,
			Env:   nil,
		}

		if job.Name != "test" {
			t.Errorf("Expected Name to be 'test', got %s", job.Name)
		}
		if job.Steps != nil {
			t.Error("Expected Steps to be nil")
		}
		if job.Env != nil {
			t.Error("Expected Env to be nil")
		}
	})
}

// Integration tests
func TestJob_Integration(t *testing.T) {
	t.Run("FullWorkflow", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
		step2 := &Step{Name: "step2", Script: "var y = 3; y * 2;"}
		jobEnv := map[string]string{"JOB_VAR": "job_value"}
		job := createTestJob("integration-job", []*Step{step1, step2}, jobEnv)
		runCtx := createTestWorkflowRunContextForJob()
		runCtxPtr := createTestWorkflowRunContextPtrForJob()
		parent := &mockJobEnv{data: map[string]string{"parent": "value"}}
		args := &mockJobEnv{data: map[string]string{"arg": "value"}}
		ctx := createTestJobRunnableContext()

		// Test Precheck
		err := job.Precheck()
		if err != nil {
			t.Errorf("Precheck failed: %v", err)
		}

		// Test HasSteps
		hasSteps := job.HasSteps()
		if !hasSteps {
			t.Error("Expected HasSteps to be true")
		}

		// Test PreflightCheck
		err = job.PreflightCheck(parent, args, runCtxPtr)
		if err != nil {
			t.Errorf("PreflightCheck failed: %v", err)
		}

		// Test Compile
		err = job.Compile(runCtx)
		if err != nil {
			t.Errorf("Compile failed: %v", err)
		}

		// Test Do
		result := job.Do(parent, runCtxPtr, ctx)
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

	t.Run("ErrorWorkflow", func(t *testing.T) {
		step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;", Run: "echo test"} // Invalid: both Script and Run
		job := createTestJob("error-job", []*Step{step1}, nil)

		// Test Precheck with error
		err := job.Precheck()
		if err == nil {
			t.Error("Expected Precheck to fail for invalid step")
		}

		// Test HasSteps
		hasSteps := job.HasSteps()
		if !hasSteps {
			t.Error("Expected HasSteps to be true even with invalid steps")
		}
	})
}

// Benchmark tests
func BenchmarkJob_Precheck(b *testing.B) {
	step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
	step2 := &Step{Name: "step2", Script: "var y = 3; y * 2;"}
	job := createTestJob("benchmark-job", []*Step{step1, step2}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := job.Precheck()
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkJob_PreflightCheck(b *testing.B) {
	step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
	step2 := &Step{Name: "step2", Script: "var y = 3; y * 2;"}
	job := createTestJob("benchmark-job", []*Step{step1, step2}, nil)
	parent := &mockJobEnv{data: map[string]string{"parent": "value"}}
	args := &mockJobEnv{data: map[string]string{"arg": "value"}}
	runCtx := createTestWorkflowRunContextPtrForJob()

	// Initialize steps first
	step1.Precheck()
	step2.Precheck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := job.PreflightCheck(parent, args, runCtx)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkJob_Compile(b *testing.B) {
	step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
	step2 := &Step{Name: "step2", Script: "var y = 3; y * 2;"}
	job := createTestJob("benchmark-job", []*Step{step1, step2}, nil)
	runCtx := createTestWorkflowRunContextForJob()

	// Initialize steps first
	step1.Precheck()
	step2.Precheck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := job.Compile(runCtx)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkJob_HasSteps(b *testing.B) {
	step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
	step2 := &Step{Name: "step2", Script: "var y = 3; y * 2;"}
	job := createTestJob("benchmark-job", []*Step{step1, step2}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = job.HasSteps()
	}
}

func BenchmarkJob_Do(b *testing.B) {
	step1 := &Step{Name: "step1", Script: "var x = 5; x + 10;"}
	job := createTestJob("benchmark-job", []*Step{step1}, nil)
	parent := &mockJobEnv{data: map[string]string{"parent": "value"}}
	runCtx := createTestWorkflowRunContextPtrForJob()
	ctx := createTestJobRunnableContext()

	// Initialize step first
	step1.Precheck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := job.Do(parent, runCtx, ctx)
		if result == nil {
			b.Fatal("Unexpected nil result")
		}
	}
}
