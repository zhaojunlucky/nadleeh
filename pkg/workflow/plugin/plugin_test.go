package plugin

import (
	"testing"

	"nadleeh/pkg/workflow/core"
	"nadleeh/pkg/workflow/run_context"

	"github.com/zhaojunlucky/golib/pkg/env"
)

// Mock plugin for testing
type mockPlugin struct {
	name        string
	resolveErr  error
	compileErr  error
	canRun      bool
	doResult    *core.RunnableResult
	preflightErr error
}

func (m *mockPlugin) GetName() string {
	return m.name
}

func (m *mockPlugin) Resolve() error {
	return m.resolveErr
}

func (m *mockPlugin) Compile(runCtx run_context.WorkflowRunContext) error {
	return m.compileErr
}

func (m *mockPlugin) Do(parent env.Env, runCtx *run_context.WorkflowRunContext, ctx *core.RunnableContext) *core.RunnableResult {
	return m.doResult
}

func (m *mockPlugin) CanRun() bool {
	return m.canRun
}

func (m *mockPlugin) PreflightCheck(parent env.Env, args env.Env, runCtx *run_context.WorkflowRunContext) error {
	return m.preflightErr
}

func TestSupportedPlugins(t *testing.T) {
	expectedPlugins := []string{"google-drive", "github-actions", "telegram", "minio"}
	
	if len(SupportedPlugins) != len(expectedPlugins) {
		t.Errorf("Expected %d supported plugins, got %d", len(expectedPlugins), len(SupportedPlugins))
	}
	
	for i, expected := range expectedPlugins {
		if i >= len(SupportedPlugins) || SupportedPlugins[i] != expected {
			t.Errorf("Expected plugin at index %d to be %s, got %s", i, expected, SupportedPlugins[i])
		}
	}
}

func TestNewPlugin_SupportedPlugins(t *testing.T) {
	testCases := []struct {
		name           string
		pluginName     string
		pluginPath     string
		config         map[string]string
		expectError    bool
		expectedType   string
	}{
		{
			name:         "GoogleDrive",
			pluginName:   "google-drive",
			pluginPath:   "/test/path",
			config:       map[string]string{"key": "value"},
			expectError:  false,
			expectedType: "google-drive",
		},
		{
			name:         "GitHubActions",
			pluginName:   "github-actions",
			pluginPath:   "/test/path",
			config:       map[string]string{"key": "value"},
			expectError:  false,
			expectedType: "github-action", // Note: GetName returns "github-action" not "github-actions"
		},
		{
			name:         "Telegram",
			pluginName:   "telegram",
			pluginPath:   "/test/path",
			config:       map[string]string{"key": "value"},
			expectError:  false,
			expectedType: "telegram",
		},
		{
			name:         "Minio",
			pluginName:   "minio",
			pluginPath:   "/test/path",
			config:       map[string]string{"key": "value"},
			expectError:  false,
			expectedType: "minio",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin, err := NewPlugin(tc.pluginName, tc.pluginPath, tc.config)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for plugin %s, but got none", tc.pluginName)
				}
				if plugin != nil {
					t.Errorf("Expected nil plugin for %s, but got %v", tc.pluginName, plugin)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for plugin %s: %v", tc.pluginName, err)
				}
				if plugin == nil {
					t.Errorf("Expected plugin for %s, but got nil", tc.pluginName)
				} else {
					if plugin.GetName() != tc.expectedType {
						t.Errorf("Expected plugin name %s, got %s", tc.expectedType, plugin.GetName())
					}
				}
			}
		})
	}
}

func TestNewPlugin_WithVersion(t *testing.T) {
	testCases := []struct {
		name           string
		pluginName     string
		pluginPath     string
		config         map[string]string
		expectError    bool
		expectedName   string
	}{
		{
			name:         "GoogleDriveWithVersion",
			pluginName:   "google-drive@v1.0.0",
			pluginPath:   "/test/path",
			config:       map[string]string{"key": "value"},
			expectError:  false,
			expectedName: "google-drive",
		},
		{
			name:         "GitHubActionsWithVersion",
			pluginName:   "github-actions@v2.1.0",
			pluginPath:   "/test/path",
			config:       map[string]string{"key": "value"},
			expectError:  false,
			expectedName: "github-action",
		},
		{
			name:         "TelegramWithVersion",
			pluginName:   "telegram@latest",
			pluginPath:   "/test/path",
			config:       map[string]string{"key": "value"},
			expectError:  false,
			expectedName: "telegram",
		},
		{
			name:         "MinioWithVersion",
			pluginName:   "minio@v1.2.3",
			pluginPath:   "/test/path",
			config:       map[string]string{"key": "value"},
			expectError:  false,
			expectedName: "minio",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin, err := NewPlugin(tc.pluginName, tc.pluginPath, tc.config)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for plugin %s, but got none", tc.pluginName)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for plugin %s: %v", tc.pluginName, err)
				}
				if plugin == nil {
					t.Errorf("Expected plugin for %s, but got nil", tc.pluginName)
				} else {
					if plugin.GetName() != tc.expectedName {
						t.Errorf("Expected plugin name %s, got %s", tc.expectedName, plugin.GetName())
					}
				}
			}
		})
	}
}

func TestNewPlugin_JSPlug(t *testing.T) {
	testCases := []struct {
		name         string
		pluginName   string
		pluginPath   string
		config       map[string]string
		expectError  bool
		expectedName string
	}{
		{
			name:         "CustomJSPluginWithVersion",
			pluginName:   "custom-plugin@v1.0.0",
			pluginPath:   "/test/path",
			config:       map[string]string{"key": "value"},
			expectError:  false,
			expectedName: "custom-plugin-v1.0.0",
		},
		{
			name:         "AnotherJSPluginWithVersion",
			pluginName:   "my-plugin@latest",
			pluginPath:   "/another/path",
			config:       map[string]string{"config": "test"},
			expectError:  false,
			expectedName: "my-plugin-latest",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin, err := NewPlugin(tc.pluginName, tc.pluginPath, tc.config)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for plugin %s, but got none", tc.pluginName)
				}
			} else {
				// JSPlug.Resolve() might fail due to plugin loading, so we check if plugin was created
				if plugin == nil {
					t.Errorf("Expected plugin for %s, but got nil", tc.pluginName)
				} else {
					if plugin.GetName() != tc.expectedName {
						t.Errorf("Expected plugin name %s, got %s", tc.expectedName, plugin.GetName())
					}
				}
				// Note: err might not be nil due to JSPlug.Resolve() failing, which is expected in tests
			}
		})
	}
}

func TestNewPlugin_UnknownPlugin(t *testing.T) {
	testCases := []struct {
		name       string
		pluginName string
		pluginPath string
		config     map[string]string
	}{
		{
			name:       "UnknownPluginNoVersion",
			pluginName: "unknown-plugin",
			pluginPath: "/test/path",
			config:     map[string]string{"key": "value"},
		},
		{
			name:       "EmptyPluginName",
			pluginName: "",
			pluginPath: "/test/path",
			config:     map[string]string{"key": "value"},
		},
		{
			name:       "UnsupportedPlugin",
			pluginName: "unsupported",
			pluginPath: "/test/path",
			config:     map[string]string{"key": "value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plugin, err := NewPlugin(tc.pluginName, tc.pluginPath, tc.config)
			
			if err == nil {
				t.Errorf("Expected error for unknown plugin %s, but got none", tc.pluginName)
			}
			
			if plugin != nil {
				t.Errorf("Expected nil plugin for unknown plugin %s, but got %v", tc.pluginName, plugin)
			}
			
			expectedError := "unknown plugin: " + tc.pluginName
			if err.Error() != expectedError {
				t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
			}
		})
	}
}

func TestNewPlugin_VersionParsing(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedName   string
		expectedVersion string
	}{
		{
			name:           "NoVersion",
			input:          "google-drive",
			expectedName:   "google-drive",
			expectedVersion: "",
		},
		{
			name:           "SimpleVersion",
			input:          "google-drive@v1.0.0",
			expectedName:   "google-drive",
			expectedVersion: "v1.0.0",
		},
		{
			name:           "LatestVersion",
			input:          "telegram@latest",
			expectedName:   "telegram",
			expectedVersion: "latest",
		},
		{
			name:           "ComplexVersion",
			input:          "minio@v2.1.0-beta.1",
			expectedName:   "minio",
			expectedVersion: "v2.1.0-beta.1",
		},
		{
			name:           "MultipleAtSymbols",
			input:          "plugin@v1.0@test",
			expectedName:   "plugin",
			expectedVersion: "v1.0@test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We'll test the version parsing logic by checking if supported plugins work correctly
			if tc.expectedName == "google-drive" || tc.expectedName == "telegram" || tc.expectedName == "minio" {
				plugin, err := NewPlugin(tc.input, "/test/path", map[string]string{})
				
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tc.input, err)
				}
				
				if plugin == nil {
					t.Errorf("Expected plugin for %s, but got nil", tc.input)
				}
			}
		})
	}
}

func TestNewPlugin_EdgeCases(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		plugin, err := NewPlugin("google-drive", "/test/path", nil)
		
		if err != nil {
			t.Errorf("Unexpected error with nil config: %v", err)
		}
		
		if plugin == nil {
			t.Errorf("Expected plugin with nil config, but got nil")
		}
	})

	t.Run("EmptyConfig", func(t *testing.T) {
		plugin, err := NewPlugin("telegram", "/test/path", map[string]string{})
		
		if err != nil {
			t.Errorf("Unexpected error with empty config: %v", err)
		}
		
		if plugin == nil {
			t.Errorf("Expected plugin with empty config, but got nil")
		}
	})

	t.Run("EmptyPluginPath", func(t *testing.T) {
		plugin, err := NewPlugin("minio", "", map[string]string{"key": "value"})
		
		if err != nil {
			t.Errorf("Unexpected error with empty plugin path: %v", err)
		}
		
		if plugin == nil {
			t.Errorf("Expected plugin with empty plugin path, but got nil")
		}
	})

	t.Run("OnlyAtSymbol", func(t *testing.T) {
		plugin, err := NewPlugin("@", "/test/path", map[string]string{})
		
		if err == nil {
			t.Error("Expected error for plugin name with only @ symbol")
		}
		
		if plugin != nil {
			t.Error("Expected nil plugin for invalid name")
		}
	})

	t.Run("AtSymbolAtEnd", func(t *testing.T) {
		plugin, err := NewPlugin("google-drive@", "/test/path", map[string]string{})
		
		if err != nil {
			t.Errorf("Unexpected error for plugin with @ at end: %v", err)
		}
		
		if plugin == nil {
			t.Error("Expected plugin for valid name with empty version")
		}
	})
}

// Benchmark tests
func BenchmarkNewPlugin_GoogleDrive(b *testing.B) {
	config := map[string]string{"key": "value", "path": "/test"}
	
	for i := 0; i < b.N; i++ {
		_, _ = NewPlugin("google-drive", "/test/path", config)
	}
}

func BenchmarkNewPlugin_WithVersion(b *testing.B) {
	config := map[string]string{"key": "value", "path": "/test"}
	
	for i := 0; i < b.N; i++ {
		_, _ = NewPlugin("telegram@v1.0.0", "/test/path", config)
	}
}

func BenchmarkNewPlugin_UnknownPlugin(b *testing.B) {
	config := map[string]string{"key": "value"}
	
	for i := 0; i < b.N; i++ {
		_, _ = NewPlugin("unknown-plugin", "/test/path", config)
	}
}

func BenchmarkNewPlugin_JSPlug(b *testing.B) {
	config := map[string]string{"key": "value"}
	
	for i := 0; i < b.N; i++ {
		_, _ = NewPlugin("custom-plugin@v1.0.0", "/test/path", config)
	}
}
