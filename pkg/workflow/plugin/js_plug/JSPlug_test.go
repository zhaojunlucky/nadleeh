package js_plug

import (
	"os"
	"path/filepath"
	"testing"

	"nadleeh/pkg/encrypt"
	"nadleeh/pkg/script"
	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"

	"gopkg.in/yaml.v3"
)

// Mock environment for testing
type mockEnv struct {
	data map[string]string
}

func newMockEnv() *mockEnv {
	return &mockEnv{
		data: make(map[string]string),
	}
}

func (m *mockEnv) Get(key string) string {
	return m.data[key]
}

func (m *mockEnv) Set(key, value string) {
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
	return value // Simple implementation for testing
}

func (m *mockEnv) SetAll(values map[string]string) {
	for k, v := range values {
		m.data[k] = v
	}
}

// Helper function to create a real JSContext for testing
func createTestJSContext() script.JSContext {
	secCtx := encrypt.SecureContext{}
	return script.NewJSContext(&secCtx)
}

func TestJSPlug_GetName(t *testing.T) {
	testCases := []struct {
		name         string
		pluginName   string
		version      string
		expectedName string
	}{
		{
			name:         "SimplePlugin",
			pluginName:   "test-plugin",
			version:      "v1.0.0",
			expectedName: "test-plugin-v1.0.0",
		},
		{
			name:         "PluginWithComplexVersion",
			pluginName:   "my-plugin",
			version:      "v2.1.0-beta.1",
			expectedName: "my-plugin-v2.1.0-beta.1",
		},
		{
			name:         "EmptyName",
			pluginName:   "",
			version:      "v1.0.0",
			expectedName: "-v1.0.0",
		},
		{
			name:         "EmptyVersion",
			pluginName:   "test-plugin",
			version:      "",
			expectedName: "test-plugin-",
		},
		{
			name:         "BothEmpty",
			pluginName:   "",
			version:      "",
			expectedName: "-",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsPlug := &JSPlug{
				PluginName: tc.pluginName,
				Version:    tc.version,
			}

			result := jsPlug.GetName()
			if result != tc.expectedName {
				t.Errorf("Expected name %s, got %s", tc.expectedName, result)
			}
		})
	}
}

func TestJSPlug_CanRun(t *testing.T) {
	testCases := []struct {
		name        string
		hasError    int
		expectedRun bool
	}{
		{
			name:        "NoError",
			hasError:    0,
			expectedRun: false,
		},
		{
			name:        "CompilationError",
			hasError:    1,
			expectedRun: false,
		},
		{
			name:        "CompilationSuccess",
			hasError:    2,
			expectedRun: true,
		},
		{
			name:        "HigherValue",
			hasError:    3,
			expectedRun: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsPlug := &JSPlug{
				hasError: tc.hasError,
			}

			result := jsPlug.CanRun()
			if result != tc.expectedRun {
				t.Errorf("Expected CanRun %v, got %v", tc.expectedRun, result)
			}
		})
	}
}

func TestJSPlug_Compile(t *testing.T) {
	// Create a temporary directory with main.js file
	tempDir := t.TempDir()
	mainFile := filepath.Join(tempDir, "main.js")
	if err := os.WriteFile(mainFile, []byte("console.log('test');"), 0644); err != nil {
		t.Fatalf("Failed to create main.js: %v", err)
	}

	t.Run("CompilationSuccess", func(t *testing.T) {
		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			pm: &PluginMetadata{
				MainFile: mainFile,
			},
			hasError: 0,
		}

		jsCtx := createTestJSContext()
		runCtx := run_context.WorkflowRunContext{
			JSCtx: jsCtx,
		}

		err := jsPlug.Compile(runCtx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if jsPlug.hasError != 2 {
			t.Errorf("Expected hasError to be 2, got %d", jsPlug.hasError)
		}
	})

	t.Run("CompilationError", func(t *testing.T) {
		// Create an invalid JS file to trigger compilation error
		invalidMainFile := filepath.Join(tempDir, "invalid.js")
		if err := os.WriteFile(invalidMainFile, []byte("invalid javascript syntax {{{"), 0644); err != nil {
			t.Fatalf("Failed to create invalid.js: %v", err)
		}

		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			pm: &PluginMetadata{
				MainFile: invalidMainFile,
			},
			hasError: 0,
		}

		jsCtx := createTestJSContext()
		runCtx := run_context.WorkflowRunContext{
			JSCtx: jsCtx,
		}

		err := jsPlug.Compile(runCtx)
		if err == nil {
			t.Error("Expected compilation error")
		}

		if jsPlug.hasError != 1 {
			t.Errorf("Expected hasError to be 1, got %d", jsPlug.hasError)
		}
	})
}

func TestJSPlug_PreflightCheck(t *testing.T) {
	t.Run("ValidManifest", func(t *testing.T) {
		// Create a temporary directory with manifest file
		tempDir := t.TempDir()
		manifestFile := filepath.Join(tempDir, "manifest.yml")
		
		manifestContent := `metadata:
  workflow_version: "1.0"
  version: "1.0.0"
  name: "test-plugin"
  description: "Test plugin"
runtime:
  args:
    - name: "input"
      pattern: ".*"
      required: true
    - name: "output"
      pattern: ""
      required: false
`
		if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
			t.Fatalf("Failed to create manifest.yml: %v", err)
		}

		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			pm: &PluginMetadata{
				ManifestFile: manifestFile,
			},
		}

		mockEnv := newMockEnv()
		mockEnv.Set("input", "test-value")
		mockEnv.Set("output", "test-output")

		jsCtx := createTestJSContext()
		runCtx := &run_context.WorkflowRunContext{
			JSCtx: jsCtx,
		}

		err := jsPlug.PreflightCheck(mockEnv, mockEnv, runCtx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("MissingRequiredArgument", func(t *testing.T) {
		// Create a temporary directory with manifest file
		tempDir := t.TempDir()
		manifestFile := filepath.Join(tempDir, "manifest.yml")
		
		manifestContent := `metadata:
  workflow_version: "1.0"
  version: "1.0.0"
  name: "test-plugin"
runtime:
  args:
    - name: "required_arg"
      required: true
`
		if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
			t.Fatalf("Failed to create manifest.yml: %v", err)
		}

		// Test manifest parsing first
		var testManifest struct {
			Metadata struct {
				WorkflowVersion string `yaml:"workflow_version"`
				Version         string `yaml:"version"`
				Name            string `yaml:"name"`
			} `yaml:"metadata"`
			Runtime struct {
				Args []struct {
					Name     string `yaml:"name"`
					Pattern  string `yaml:"pattern"`
					Required bool   `yaml:"required"`
				} `yaml:"args"`
			} `yaml:"runtime"`
		}
		
		file, err := os.Open(manifestFile)
		if err != nil {
			t.Fatalf("Failed to open manifest file: %v", err)
		}
		defer file.Close()
		
		err = yaml.NewDecoder(file).Decode(&testManifest)
		if err != nil {
			t.Fatalf("Failed to parse manifest: %v", err)
		}
		
		if len(testManifest.Runtime.Args) == 0 {
			t.Fatal("No args found in manifest")
		}
		
		if testManifest.Runtime.Args[0].Name != "required_arg" {
			t.Fatalf("Expected arg name 'required_arg', got '%s'", testManifest.Runtime.Args[0].Name)
		}
		
		if !testManifest.Runtime.Args[0].Required {
			t.Fatal("Expected arg to be required")
		}

		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			pm: &PluginMetadata{
				ManifestFile: manifestFile,
			},
		}

		mockEnv := newMockEnv()
		// Don't set the required argument
		
		jsCtx := createTestJSContext()
		runCtx := &run_context.WorkflowRunContext{
			JSCtx: jsCtx,
		}

		err = jsPlug.PreflightCheck(mockEnv, mockEnv, runCtx)
		if err == nil {
			t.Error("Expected error for missing required argument")
			return
		}

		t.Logf("PreflightCheck returned error: %v", err)
		t.Logf("Error type: %T", err)
		
		if err != nil {
			expectedError := "required argument 'required_arg' is not provided for plugin test-plugin"
			if err.Error() != expectedError {
				t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
			}
		}
	})

	t.Run("PatternMismatch", func(t *testing.T) {
		// Create a temporary directory with manifest file
		tempDir := t.TempDir()
		manifestFile := filepath.Join(tempDir, "manifest.yml")
		
		manifestContent := `metadata:
  workflow_version: "1.0"
  version: "1.0.0"
  name: "test-plugin"
runtime:
  args:
    - name: "email"
      pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
      required: true
`
		if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
			t.Fatalf("Failed to create manifest.yml: %v", err)
		}

		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			pm: &PluginMetadata{
				ManifestFile: manifestFile,
			},
		}

		mockEnv := newMockEnv()
		mockEnv.Set("email", "invalid-email")

		jsCtx := createTestJSContext()
		runCtx := &run_context.WorkflowRunContext{
			JSCtx: jsCtx,
		}

		err := jsPlug.PreflightCheck(mockEnv, mockEnv, runCtx)
		if err == nil {
			t.Error("Expected error for pattern mismatch")
		}

		expectedError := "argument 'email' value doesn't match pattern '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$' for plugin test-plugin"
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("InvalidManifestFile", func(t *testing.T) {
		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			pm: &PluginMetadata{
				ManifestFile: "/nonexistent/manifest.yml",
			},
		}

		mockEnv := newMockEnv()

		jsCtx := createTestJSContext()
		runCtx := &run_context.WorkflowRunContext{
			JSCtx: jsCtx,
		}

		err := jsPlug.PreflightCheck(mockEnv, mockEnv, runCtx)
		if err == nil {
			t.Error("Expected error for nonexistent manifest file")
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		// Create a temporary directory with invalid manifest file
		tempDir := t.TempDir()
		manifestFile := filepath.Join(tempDir, "manifest.yml")
		
		invalidYAML := `invalid: yaml: content: [unclosed`
		if err := os.WriteFile(manifestFile, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("Failed to create invalid manifest.yml: %v", err)
		}

		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			pm: &PluginMetadata{
				ManifestFile: manifestFile,
			},
		}

		mockEnv := newMockEnv()

		jsCtx := createTestJSContext()
		runCtx := &run_context.WorkflowRunContext{
			JSCtx: jsCtx,
		}

		err := jsPlug.PreflightCheck(mockEnv, mockEnv, runCtx)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestJSPlug_Resolve(t *testing.T) {
	// Save original PM
	originalPM := PM
	defer func() { PM = originalPM }()

	t.Run("ResolveSuccess", func(t *testing.T) {
		// Create a temporary directory with required files
		tempDir := t.TempDir()
		
		// Create main.js file
		mainFile := filepath.Join(tempDir, "main.js")
		if err := os.WriteFile(mainFile, []byte("console.log('test');"), 0644); err != nil {
			t.Fatalf("Failed to create main.js: %v", err)
		}
		
		// Create manifest.yml file
		manifestFile := filepath.Join(tempDir, "manifest.yml")
		if err := os.WriteFile(manifestFile, []byte("name: test\nversion: 1.0.0"), 0644); err != nil {
			t.Fatalf("Failed to create manifest.yml: %v", err)
		}

		// Create a mock plugin manager
		mockPM := &PluginManager{
			LoadedPlugin: make(map[string]*PluginMetadata),
		}
		PM = mockPM

		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			Version:    "v1.0.0",
			PluginPath: tempDir,
		}

		err := jsPlug.Resolve()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if jsPlug.pm == nil {
			t.Error("Expected pm to be set after resolve")
		}
	})

	t.Run("ResolveError", func(t *testing.T) {
		// Create a mock plugin manager that will fail
		mockPM := &PluginManager{
			LoadedPlugin: make(map[string]*PluginMetadata),
		}
		PM = mockPM

		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			Version:    "v1.0.0",
			PluginPath: "/nonexistent/path",
		}

		err := jsPlug.Resolve()
		if err == nil {
			t.Error("Expected error for nonexistent plugin path")
		}

		if jsPlug.pm != nil {
			t.Error("Expected pm to be nil after failed resolve")
		}
	})
}

func TestJSPlug_Do(t *testing.T) {
	t.Run("DoSuccess", func(t *testing.T) {
		// Create a temporary directory with main.js file
		tempDir := t.TempDir()
		mainFile := filepath.Join(tempDir, "main.js")
		if err := os.WriteFile(mainFile, []byte("'test output'"), 0644); err != nil {
			t.Fatalf("Failed to create main.js: %v", err)
		}

		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			Config:     map[string]string{"key": "value"},
			pm: &PluginMetadata{
				MainFile: mainFile,
			},
		}

		mockEnv := newMockEnv()
		jsCtx := createTestJSContext()
		
		runCtx := &run_context.WorkflowRunContext{
			JSCtx: jsCtx,
		}

		ctx := &core.RunnableContext{
			NeedOutput: true,
			Args:       mockEnv,
		}

		result := jsPlug.Do(mockEnv, runCtx, ctx)
		if result == nil {
			t.Fatal("Expected result to be returned")
		}

		if result.Err != nil {
			t.Errorf("Unexpected error: %v", result.Err)
		}

		if result.ReturnCode != 0 {
			t.Errorf("Expected return code 0, got %d", result.ReturnCode)
		}

		if result.Output != "test output" {
			t.Errorf("Expected output 'test output', got '%s'", result.Output)
		}
	})

	t.Run("DoWithRunError", func(t *testing.T) {
		// Create a temporary directory with invalid main.js file
		tempDir := t.TempDir()
		mainFile := filepath.Join(tempDir, "main.js")
		if err := os.WriteFile(mainFile, []byte("throw new Error('test error');"), 0644); err != nil {
			t.Fatalf("Failed to create main.js: %v", err)
		}

		jsPlug := &JSPlug{
			PluginName: "test-plugin",
			Config:     map[string]string{"key": "value"},
			pm: &PluginMetadata{
				MainFile: mainFile,
			},
		}

		mockEnv := newMockEnv()
		jsCtx := createTestJSContext()
		
		runCtx := &run_context.WorkflowRunContext{
			JSCtx: jsCtx,
		}

		ctx := &core.RunnableContext{
			NeedOutput: true,
			Args:       mockEnv,
		}

		result := jsPlug.Do(mockEnv, runCtx, ctx)
		if result == nil {
			t.Fatal("Expected result to be returned")
		}

		if result.Err == nil {
			t.Error("Expected error from JS execution")
		}

		if result.ReturnCode != 1 {
			t.Errorf("Expected return code 1, got %d", result.ReturnCode)
		}
	})
}

func TestJSPlug_Structs(t *testing.T) {
	t.Run("MetadataStruct", func(t *testing.T) {
		m := metadata{
			WorkflowVersion: "1.0",
			Version:         "1.0.0",
			Name:            "test-plugin",
			Description:     "Test plugin description",
		}

		if m.WorkflowVersion != "1.0" {
			t.Errorf("Expected WorkflowVersion '1.0', got '%s'", m.WorkflowVersion)
		}

		if m.Version != "1.0.0" {
			t.Errorf("Expected Version '1.0.0', got '%s'", m.Version)
		}

		if m.Name != "test-plugin" {
			t.Errorf("Expected Name 'test-plugin', got '%s'", m.Name)
		}

		if m.Description != "Test plugin description" {
			t.Errorf("Expected Description 'Test plugin description', got '%s'", m.Description)
		}
	})

	t.Run("RuntimeStruct", func(t *testing.T) {
		r := runtime{
			Args: []struct {
				Name     string `yaml:"name"`
				Pattern  string `yaml:"pattern"`
				Required bool   `yaml:"required"`
			}{
				{Name: "input", Pattern: ".*", Required: true},
				{Name: "output", Pattern: "", Required: false},
			},
		}

		if len(r.Args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(r.Args))
		}

		if r.Args[0].Name != "input" {
			t.Errorf("Expected first arg name 'input', got '%s'", r.Args[0].Name)
		}

		if r.Args[0].Required != true {
			t.Errorf("Expected first arg to be required")
		}

		if r.Args[1].Name != "output" {
			t.Errorf("Expected second arg name 'output', got '%s'", r.Args[1].Name)
		}

		if r.Args[1].Required != false {
			t.Errorf("Expected second arg to be optional")
		}
	})

	t.Run("ManifestStruct", func(t *testing.T) {
		m := manifest{
			Metadata: metadata{
				Name:    "test-plugin",
				Version: "1.0.0",
			},
			Runtime: runtime{
				Args: []struct {
					Name     string `yaml:"name"`
					Pattern  string `yaml:"pattern"`
					Required bool   `yaml:"required"`
				}{
					{Name: "test-arg", Pattern: ".*", Required: true},
				},
			},
		}

		if m.Metadata.Name != "test-plugin" {
			t.Errorf("Expected metadata Name 'test-plugin', got '%s'", m.Metadata.Name)
		}

		if len(m.Runtime.Args) != 1 {
			t.Errorf("Expected 1 runtime arg, got %d", len(m.Runtime.Args))
		}

		if m.Runtime.Args[0].Name != "test-arg" {
			t.Errorf("Expected runtime arg Name 'test-arg', got '%s'", m.Runtime.Args[0].Name)
		}
	})
}

// Benchmark tests
func BenchmarkJSPlug_GetName(b *testing.B) {
	jsPlug := &JSPlug{
		PluginName: "test-plugin",
		Version:    "v1.0.0",
	}

	for i := 0; i < b.N; i++ {
		jsPlug.GetName()
	}
}

func BenchmarkJSPlug_CanRun(b *testing.B) {
	jsPlug := &JSPlug{
		hasError: 2,
	}

	for i := 0; i < b.N; i++ {
		jsPlug.CanRun()
	}
}

func BenchmarkJSPlug_Compile(b *testing.B) {
	// Create a temporary directory with main.js file
	tempDir := b.TempDir()
	mainFile := filepath.Join(tempDir, "main.js")
	if err := os.WriteFile(mainFile, []byte("console.log('test');"), 0644); err != nil {
		b.Fatalf("Failed to create main.js: %v", err)
	}

	jsPlug := &JSPlug{
		PluginName: "test-plugin",
		pm: &PluginMetadata{
			MainFile: mainFile,
		},
	}

	jsCtx := createTestJSContext()
	runCtx := run_context.WorkflowRunContext{
		JSCtx: jsCtx,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsPlug.hasError = 0 // Reset for each iteration
		jsPlug.Compile(runCtx)
	}
}
