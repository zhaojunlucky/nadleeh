package core

import (
	"errors"
	"strings"
	"testing"
)

// Test RunnableStatus struct fields and initialization
func TestRunnableStatusStruct(t *testing.T) {
	t.Run("StructFields", func(t *testing.T) {
		status := NewRunnableStatus("test-runnable", "test-type")
		
		// Verify initial state
		if status.name != "test-runnable" {
			t.Errorf("Expected name 'test-runnable', got '%s'", status.name)
		}
		if status.rType != "test-type" {
			t.Errorf("Expected rType 'test-type', got '%s'", status.rType)
		}
		if status.status != NotStart {
			t.Errorf("Expected status '%s', got '%s'", NotStart, status.status)
		}
		if status.errs != nil {
			t.Errorf("Expected errs to be nil, got %v", status.errs)
		}
		if status.childs != nil {
			t.Errorf("Expected childs to be nil, got %v", status.childs)
		}
		if status.childMap == nil {
			t.Error("Expected childMap to be initialized")
		}
		if status.ContinueOnErr {
			t.Error("Expected ContinueOnErr to be false")
		}
	})
}

// Test NewRunnableStatus constructor
func TestNewRunnableStatus(t *testing.T) {
	tests := []struct {
		name     string
		rType    string
		testName string
	}{
		{"test-job", "job", "ValidJobStatus"},
		{"test-step", "step", "ValidStepStatus"},
		{"test-workflow", "workflow", "ValidWorkflowStatus"},
		{"", "", "EmptyNameAndType"},
		{"long-name-with-dashes-and-numbers-123", "complex-type", "LongNamesAndTypes"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			status := NewRunnableStatus(tt.name, tt.rType)
			
			if status == nil {
				t.Fatal("Expected non-nil RunnableStatus")
			}
			if status.name != tt.name {
				t.Errorf("Expected name '%s', got '%s'", tt.name, status.name)
			}
			if status.rType != tt.rType {
				t.Errorf("Expected rType '%s', got '%s'", tt.rType, status.rType)
			}
			if status.status != NotStart {
				t.Errorf("Expected initial status '%s', got '%s'", NotStart, status.status)
			}
			if status.childMap == nil {
				t.Error("Expected childMap to be initialized")
			}
		})
	}
}

// Test Status method
func TestRunnableStatus_Status(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		expectedStatus string
	}{
		{"NotStartStatus", NotStart, NotStart},
		{"RunningStatus", Running, Running},
		{"PassStatus", Pass, Pass},
		{"FailStatus", Fail, Fail},
		{"SkippedStatus", Skipped, Skipped},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := NewRunnableStatus("test", "test")
			status.status = tt.initialStatus
			
			result := status.Status()
			if result != tt.expectedStatus {
				t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, result)
			}
		})
	}
}

// Test Start method
func TestRunnableStatus_Start(t *testing.T) {
	t.Run("StartFromNotStart", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		
		// Verify initial state
		if status.Status() != NotStart {
			t.Errorf("Expected initial status '%s', got '%s'", NotStart, status.Status())
		}
		
		status.Start()
		
		if status.Status() != Running {
			t.Errorf("Expected status after Start() to be '%s', got '%s'", Running, status.Status())
		}
	})

	t.Run("StartFromOtherStates", func(t *testing.T) {
		states := []string{Pass, Fail, Skipped}
		
		for _, state := range states {
			status := NewRunnableStatus("test", "test")
			status.status = state
			
			status.Start()
			
			if status.Status() != Running {
				t.Errorf("Expected status after Start() to be '%s', got '%s'", Running, status.Status())
			}
		}
	})
}

// Test Skipped method
func TestRunnableStatus_Skipped(t *testing.T) {
	t.Run("SkippedFromNotStart", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		
		status.Skipped()
		
		if status.Status() != Skipped {
			t.Errorf("Expected status after Skipped() to be '%s', got '%s'", Skipped, status.Status())
		}
	})

	t.Run("SkippedFromOtherStates", func(t *testing.T) {
		states := []string{Running, Pass, Fail}
		
		for _, state := range states {
			status := NewRunnableStatus("test", "test")
			status.status = state
			
			status.Skipped()
			
			if status.Status() != Skipped {
				t.Errorf("Expected status after Skipped() to be '%s', got '%s'", Skipped, status.Status())
			}
		}
	})
}

// Test Finish method
func TestRunnableStatus_Finish(t *testing.T) {
	t.Run("FinishWithoutErrors", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		status.Start()
		
		status.Finish()
		
		if status.Status() != Pass {
			t.Errorf("Expected status after Finish() with no errors to be '%s', got '%s'", Pass, status.Status())
		}
		if len(status.errs) != 0 {
			t.Errorf("Expected no errors, got %v", status.errs)
		}
	})

	t.Run("FinishWithSingleError", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		status.Start()
		
		err := errors.New("test error")
		status.Finish(err)
		
		if status.Status() != Fail {
			t.Errorf("Expected status after Finish() with error to be '%s', got '%s'", Fail, status.Status())
		}
		if len(status.errs) != 1 {
			t.Errorf("Expected 1 error, got %d", len(status.errs))
		}
		if status.errs[0] != "test error" {
			t.Errorf("Expected error message 'test error', got '%s'", status.errs[0])
		}
	})

	t.Run("FinishWithMultipleErrors", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		status.Start()
		
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		err3 := errors.New("error 3")
		status.Finish(err1, err2, err3)
		
		if status.Status() != Fail {
			t.Errorf("Expected status after Finish() with errors to be '%s', got '%s'", Fail, status.Status())
		}
		if len(status.errs) != 3 {
			t.Errorf("Expected 3 errors, got %d", len(status.errs))
		}
		
		expectedErrors := []string{"error 1", "error 2", "error 3"}
		for i, expectedErr := range expectedErrors {
			if status.errs[i] != expectedErr {
				t.Errorf("Expected error[%d] to be '%s', got '%s'", i, expectedErr, status.errs[i])
			}
		}
	})

	t.Run("FinishWithEmptyErrorSlice", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		status.Start()
		
		var errs []error
		status.Finish(errs...)
		
		if status.Status() != Pass {
			t.Errorf("Expected status after Finish() with empty error slice to be '%s', got '%s'", Pass, status.Status())
		}
	})
}

// Test AddChild method
func TestRunnableStatus_AddChild(t *testing.T) {
	t.Run("AddSingleChild", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		child := NewRunnableStatus("child", "step")
		
		parent.AddChild(child)
		
		if len(parent.childs) != 1 {
			t.Errorf("Expected 1 child, got %d", len(parent.childs))
		}
		if parent.childs[0] != child {
			t.Error("Expected child to be added to childs slice")
		}
		if parent.childMap["child"] != child {
			t.Error("Expected child to be added to childMap")
		}
	})

	t.Run("AddMultipleChildren", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		child1 := NewRunnableStatus("child1", "step")
		child2 := NewRunnableStatus("child2", "step")
		child3 := NewRunnableStatus("child3", "step")
		
		parent.AddChild(child1)
		parent.AddChild(child2)
		parent.AddChild(child3)
		
		if len(parent.childs) != 3 {
			t.Errorf("Expected 3 children, got %d", len(parent.childs))
		}
		if len(parent.childMap) != 3 {
			t.Errorf("Expected 3 children in map, got %d", len(parent.childMap))
		}
		
		// Verify order is preserved
		if parent.childs[0] != child1 || parent.childs[1] != child2 || parent.childs[2] != child3 {
			t.Error("Expected children order to be preserved")
		}
	})

	t.Run("AddChildWithEmptyName", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		child := NewRunnableStatus("", "step")
		
		parent.AddChild(child)
		
		if len(parent.childs) != 1 {
			t.Errorf("Expected 1 child, got %d", len(parent.childs))
		}
		if len(parent.childMap) != 0 {
			t.Errorf("Expected 0 children in map for empty name, got %d", len(parent.childMap))
		}
	})
}

// Test GetChild method
func TestRunnableStatus_GetChild(t *testing.T) {
	t.Run("GetExistingChild", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		child := NewRunnableStatus("target-child", "step")
		
		parent.AddChild(child)
		
		result := parent.GetChild("target-child")
		if result != child {
			t.Error("Expected to get the correct child")
		}
	})

	t.Run("GetNonExistentChild", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		
		result := parent.GetChild("non-existent")
		if result != nil {
			t.Error("Expected nil for non-existent child")
		}
	})

	t.Run("GetChildFromMultipleChildren", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		child1 := NewRunnableStatus("child1", "step")
		child2 := NewRunnableStatus("child2", "step")
		child3 := NewRunnableStatus("child3", "step")
		
		parent.AddChild(child1)
		parent.AddChild(child2)
		parent.AddChild(child3)
		
		result := parent.GetChild("child2")
		if result != child2 {
			t.Error("Expected to get child2")
		}
	})
}

// Test GetChildByIndex method
func TestRunnableStatus_GetChildByIndex(t *testing.T) {
	t.Run("GetChildByValidIndex", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		child1 := NewRunnableStatus("child1", "step")
		child2 := NewRunnableStatus("child2", "step")
		
		parent.AddChild(child1)
		parent.AddChild(child2)
		
		result0 := parent.GetChildByIndex(0)
		if result0 != child1 {
			t.Error("Expected to get child1 at index 0")
		}
		
		result1 := parent.GetChildByIndex(1)
		if result1 != child2 {
			t.Error("Expected to get child2 at index 1")
		}
	})

	// Note: We don't test invalid indices as that would cause a panic
	// which is the expected behavior for out-of-bounds access
}

// Test errors method (private method tested through Reason)
func TestRunnableStatus_errors(t *testing.T) {
	t.Run("NoErrors", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		
		reason := status.Reason()
		if reason != "" {
			t.Errorf("Expected empty reason for no errors, got '%s'", reason)
		}
	})

	t.Run("SingleError", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		status.Finish(errors.New("test error"))
		
		reason := status.Reason()
		if reason != "test error" {
			t.Errorf("Expected reason 'test error', got '%s'", reason)
		}
	})

	t.Run("MultipleErrors", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		status.Finish(errors.New("error 1"), errors.New("error 2"))
		
		reason := status.Reason()
		expected := "error 1\nerror 2"
		if reason != expected {
			t.Errorf("Expected reason '%s', got '%s'", expected, reason)
		}
	})

	t.Run("ErrorsFromChildren", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		child1 := NewRunnableStatus("child1", "step")
		child2 := NewRunnableStatus("child2", "step")
		
		parent.AddChild(child1)
		parent.AddChild(child2)
		
		parent.Finish(errors.New("parent error"))
		child1.Finish(errors.New("child1 error"))
		child2.Finish(errors.New("child2 error"))
		
		reason := parent.Reason()
		expectedErrors := []string{"parent error", "child1 error", "child2 error"}
		
		for _, expectedErr := range expectedErrors {
			if !strings.Contains(reason, expectedErr) {
				t.Errorf("Expected reason to contain '%s', got '%s'", expectedErr, reason)
			}
		}
	})
}

// Test Reason method
func TestRunnableStatus_Reason(t *testing.T) {
	t.Run("EmptyReason", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		
		reason := status.Reason()
		if reason != "" {
			t.Errorf("Expected empty reason, got '%s'", reason)
		}
	})

	t.Run("ReasonWithNewlines", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		status.Finish(errors.New("line 1"), errors.New("line 2"), errors.New("line 3"))
		
		reason := status.Reason()
		lines := strings.Split(reason, "\n")
		if len(lines) != 3 {
			t.Errorf("Expected 3 lines in reason, got %d", len(lines))
		}
		if lines[0] != "line 1" || lines[1] != "line 2" || lines[2] != "line 3" {
			t.Errorf("Expected lines to match errors, got %v", lines)
		}
	})
}

// Test FutureStatus method
func TestRunnableStatus_FutureStatus(t *testing.T) {
	t.Run("PassStatus", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		status.Finish() // Pass status
		
		future := status.FutureStatus()
		if future != Pass {
			t.Errorf("Expected future status '%s', got '%s'", Pass, future)
		}
	})

	t.Run("FailStatusWithContinueOnErr", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		status.Finish(errors.New("test error")) // Fail status
		status.ContinueOnErr = true
		
		future := status.FutureStatus()
		if future != Pass {
			t.Errorf("Expected future status '%s' with ContinueOnErr=true, got '%s'", Pass, future)
		}
	})

	t.Run("FailStatusWithoutContinueOnErr", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		status.Finish(errors.New("test error")) // Fail status
		status.ContinueOnErr = false
		
		future := status.FutureStatus()
		if future != Fail {
			t.Errorf("Expected future status '%s' with ContinueOnErr=false, got '%s'", Fail, future)
		}
	})

	t.Run("ChildFailureAffectsFuture", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		child := NewRunnableStatus("child", "step")
		
		parent.AddChild(child)
		parent.Finish() // Parent passes
		child.Finish(errors.New("child error")) // Child fails
		
		future := parent.FutureStatus()
		if future != Fail {
			t.Errorf("Expected future status '%s' due to child failure, got '%s'", Fail, future)
		}
	})

	t.Run("ChildFailureWithContinueOnErr", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		child := NewRunnableStatus("child", "step")
		
		parent.AddChild(child)
		parent.Finish() // Parent passes
		child.Finish(errors.New("child error")) // Child fails
		child.ContinueOnErr = true
		
		future := parent.FutureStatus()
		if future != Pass {
			t.Errorf("Expected future status '%s' with child ContinueOnErr=true, got '%s'", Pass, future)
		}
	})

	t.Run("MultipleChildrenMixedStatus", func(t *testing.T) {
		parent := NewRunnableStatus("parent", "job")
		child1 := NewRunnableStatus("child1", "step")
		child2 := NewRunnableStatus("child2", "step")
		child3 := NewRunnableStatus("child3", "step")
		
		parent.AddChild(child1)
		parent.AddChild(child2)
		parent.AddChild(child3)
		
		parent.Finish() // Parent passes
		child1.Finish() // Child1 passes
		child2.Finish(errors.New("child2 error")) // Child2 fails
		child2.ContinueOnErr = true
		child3.Finish() // Child3 passes
		
		future := parent.FutureStatus()
		if future != Pass {
			t.Errorf("Expected future status '%s' with mixed children (one with ContinueOnErr), got '%s'", Pass, future)
		}
	})

	t.Run("DeepNestedChildFailure", func(t *testing.T) {
		grandparent := NewRunnableStatus("grandparent", "workflow")
		parent := NewRunnableStatus("parent", "job")
		child := NewRunnableStatus("child", "step")
		
		grandparent.AddChild(parent)
		parent.AddChild(child)
		
		grandparent.Finish() // Grandparent passes
		parent.Finish()      // Parent passes
		child.Finish(errors.New("deep child error")) // Deep child fails
		
		future := grandparent.FutureStatus()
		if future != Fail {
			t.Errorf("Expected future status '%s' due to deep child failure, got '%s'", Fail, future)
		}
	})
}

// Test edge cases and integration scenarios
func TestRunnableStatus_EdgeCases(t *testing.T) {
	t.Run("EmptyNameHandling", func(t *testing.T) {
		status := NewRunnableStatus("", "")
		
		if status.name != "" {
			t.Errorf("Expected empty name to be preserved, got '%s'", status.name)
		}
		if status.rType != "" {
			t.Errorf("Expected empty rType to be preserved, got '%s'", status.rType)
		}
	})

	t.Run("ContinueOnErrFlag", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		
		// Initially false
		if status.ContinueOnErr {
			t.Error("Expected ContinueOnErr to be false initially")
		}
		
		// Can be set to true
		status.ContinueOnErr = true
		if !status.ContinueOnErr {
			t.Error("Expected ContinueOnErr to be true after setting")
		}
	})

	t.Run("StatusTransitions", func(t *testing.T) {
		status := NewRunnableStatus("test", "test")
		
		// NotStart -> Running -> Pass
		if status.Status() != NotStart {
			t.Errorf("Expected initial status '%s'", NotStart)
		}
		
		status.Start()
		if status.Status() != Running {
			t.Errorf("Expected status '%s' after Start()", Running)
		}
		
		status.Finish()
		if status.Status() != Pass {
			t.Errorf("Expected status '%s' after Finish()", Pass)
		}
		
		// Can transition to Skipped from any state
		status.Skipped()
		if status.Status() != Skipped {
			t.Errorf("Expected status '%s' after Skipped()", Skipped)
		}
	})
}

// Benchmark tests for RunnableStatus performance
func BenchmarkRunnableStatus_NewRunnableStatus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewRunnableStatus("test-name", "test-type")
	}
}

func BenchmarkRunnableStatus_Status(b *testing.B) {
	status := NewRunnableStatus("test", "test")
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = status.Status()
	}
}

func BenchmarkRunnableStatus_Start(b *testing.B) {
	status := NewRunnableStatus("test", "test")
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		status.Start()
	}
}

func BenchmarkRunnableStatus_Finish(b *testing.B) {
	status := NewRunnableStatus("test", "test")
	err := errors.New("benchmark error")
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		status.Finish(err)
	}
}

func BenchmarkRunnableStatus_AddChild(b *testing.B) {
	parent := NewRunnableStatus("parent", "job")
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		child := NewRunnableStatus("child", "step")
		parent.AddChild(child)
	}
}

func BenchmarkRunnableStatus_GetChild(b *testing.B) {
	parent := NewRunnableStatus("parent", "job")
	child := NewRunnableStatus("target", "step")
	parent.AddChild(child)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = parent.GetChild("target")
	}
}

func BenchmarkRunnableStatus_FutureStatus(b *testing.B) {
	status := NewRunnableStatus("test", "test")
	status.Finish()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = status.FutureStatus()
	}
}

func BenchmarkRunnableStatus_Reason(b *testing.B) {
	status := NewRunnableStatus("test", "test")
	status.Finish(errors.New("error 1"), errors.New("error 2"))
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = status.Reason()
	}
}
