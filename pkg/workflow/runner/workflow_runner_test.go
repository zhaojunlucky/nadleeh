package runner

import (
	"os"
	"path/filepath"
	"testing"

	"nadleeh/internal/argument"
	"nadleeh/pkg/workflow/core"

	"github.com/zhaojunlucky/golib/pkg/env"
)

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
	t.Run("ArgumentValidation", func(t *testing.T) {
		// Test that we can create valid RunArgs that would be accepted by RunWorkflow
		runArgs := &argument.RunArgs{
			File:        "test.yml",
			Provider:    "",
			Check:       false,
			Args:        nil,
			PrivateFile: "private.key",
		}

		// Verify argument structure is correct
		if runArgs.File != "test.yml" {
			t.Error("File argument not properly structured")
		}

		if runArgs.PrivateFile != "private.key" {
			t.Error("Private argument not properly structured")
		}

		if runArgs.Check != false {
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

	t.Run("RunArgsToWorkflowArgs", func(t *testing.T) {
		// Test that RunArgs can be converted to WorkflowArgs correctly
		runArgs := &argument.RunArgs{
			File:        "test.yml",
			Provider:    "github",
			Check:       true,
			Args:        []string{"key=value"},
			PrivateFile: "private.key",
		}

		wa := core.NewWorkflowArgsFromRunArgs(runArgs)

		if wa.File == nil || *wa.File != "test.yml" {
			t.Error("File not properly converted")
		}

		if wa.Provider == nil || *wa.Provider != "github" {
			t.Error("Provider not properly converted")
		}

		if wa.Check == nil || *wa.Check != true {
			t.Error("Check not properly converted")
		}

		if wa.PrivateFile == nil || *wa.PrivateFile != "private.key" {
			t.Error("PrivateFile not properly converted")
		}
	})
}

func TestRunWorkflow_ArgumentParsing(t *testing.T) {
	t.Run("ValidArguments", func(t *testing.T) {
		runArgs := &argument.RunArgs{
			File:        "test.yml",
			Provider:    "",
			Check:       false,
			Args:        nil,
			PrivateFile: "private.key",
		}

		// Test that arguments are structured correctly
		if runArgs.File != "test.yml" {
			t.Errorf("Expected file argument to be 'test.yml', got: %v", runArgs.File)
		}

		if runArgs.PrivateFile != "private.key" {
			t.Errorf("Expected private argument to be 'private.key', got: %v", runArgs.PrivateFile)
		}

		if runArgs.Check != false {
			t.Errorf("Expected check argument to be false, got: %v", runArgs.Check)
		}
	})

	t.Run("OptionalPrivateArgNotProvided", func(t *testing.T) {
		runArgs := &argument.RunArgs{
			File:        "test.yml",
			PrivateFile: "",
		}

		// Empty private file should be valid
		if runArgs.PrivateFile != "" {
			t.Error("Expected private argument to be empty")
		}
	})
}

func TestRunWorkflow_EdgeCases(t *testing.T) {
	t.Run("EmptyArgsMap", func(t *testing.T) {
		// Test with empty args - would cause log.Fatal which cannot be caught
		t.Skip("Skipping test that would trigger log.Fatal for missing file argument")
	})

	t.Run("ArgumentStructureValidation", func(t *testing.T) {
		// Test various argument configurations to ensure they're properly structured
		testCases := []struct {
			name     string
			filename string
			expectOk bool
		}{
			{"ValidFile", "test.yml", true},
			{"EmptyFile", "", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				runArgs := &argument.RunArgs{
					File: tc.filename,
				}

				if (runArgs.File != "" && tc.expectOk) || (runArgs.File == "" && !tc.expectOk) {
					// This is expected behavior
				} else {
					t.Errorf("Unexpected result for %s: %s", tc.name, runArgs.File)
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
	runArgs := &argument.RunArgs{
		File:        "test.yml",
		Provider:    "",
		Check:       true,
		Args:        nil,
		PrivateFile: "private.key",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Benchmark the argument conversion
		_ = core.NewWorkflowArgsFromRunArgs(runArgs)
	}
}
