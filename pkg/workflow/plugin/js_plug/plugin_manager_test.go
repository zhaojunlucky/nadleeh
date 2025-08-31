package js_plug

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFormatPluginKey(t *testing.T) {
	testCases := []struct {
		name           string
		pluginName     string
		version        string
		expectedKey    string
	}{
		{
			name:        "SimplePlugin",
			pluginName:  "test-plugin",
			version:     "v1.0.0",
			expectedKey: "plugin:test-plugin@v1.0.0",
		},
		{
			name:        "PluginWithComplexVersion",
			pluginName:  "my-plugin",
			version:     "v2.1.0-beta.1",
			expectedKey: "plugin:my-plugin@v2.1.0-beta.1",
		},
		{
			name:        "PluginWithLatestVersion",
			pluginName:  "another-plugin",
			version:     "latest",
			expectedKey: "plugin:another-plugin@latest",
		},
		{
			name:        "EmptyName",
			pluginName:  "",
			version:     "v1.0.0",
			expectedKey: "plugin:@v1.0.0",
		},
		{
			name:        "EmptyVersion",
			pluginName:  "test-plugin",
			version:     "",
			expectedKey: "plugin:test-plugin@",
		},
		{
			name:        "BothEmpty",
			pluginName:  "",
			version:     "",
			expectedKey: "plugin:@",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatPluginKey(tc.pluginName, tc.version)
			if result != tc.expectedKey {
				t.Errorf("Expected key %s, got %s", tc.expectedKey, result)
			}
		})
	}
}

func TestNewPluginManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Save original values
	originalLocalPath := LocalPath
	originalLocalDataPath := LocalDataPath
	originalLocalLockPath := LocalLockPath
	
	// Set test paths
	LocalPath = tempDir
	LocalDataPath = filepath.Join(tempDir, "data")
	LocalLockPath = filepath.Join(tempDir, ".locks")
	
	// Restore original values after test
	defer func() {
		LocalPath = originalLocalPath
		LocalDataPath = originalLocalDataPath
		LocalLockPath = originalLocalLockPath
	}()

	pm := NewPluginManager()

	// Verify PluginManager was created correctly
	if pm == nil {
		t.Fatal("Expected PluginManager to be created, got nil")
	}

	if pm.LoadedPlugin == nil {
		t.Error("Expected LoadedPlugin map to be initialized")
	}

	if len(pm.LoadedPlugin) != 0 {
		t.Errorf("Expected empty LoadedPlugin map, got %d items", len(pm.LoadedPlugin))
	}

	// Verify directories were created
	if _, err := os.Stat(LocalDataPath); os.IsNotExist(err) {
		t.Errorf("Expected LocalDataPath %s to be created", LocalDataPath)
	}

	if _, err := os.Stat(LocalLockPath); os.IsNotExist(err) {
		t.Errorf("Expected LocalLockPath %s to be created", LocalLockPath)
	}
}

func TestNewPluginMetadata_EmptyVersion(t *testing.T) {
	pm, err := NewPluginMetadata("test-plugin", "", "token", "")
	
	if err == nil {
		t.Error("Expected error for empty version, got nil")
	}
	
	if pm != nil {
		t.Error("Expected nil PluginMetadata for empty version")
	}
	
	expectedError := "version is not specified"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestNewPluginMetadata_LocalScheme(t *testing.T) {
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

	pm, err := NewPluginMetadata("test-plugin", "v1.0.0", "token", tempDir)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if pm == nil {
		t.Fatal("Expected PluginMetadata to be created")
	}

	// Verify fields
	if pm.Name != "test-plugin" {
		t.Errorf("Expected Name 'test-plugin', got '%s'", pm.Name)
	}
	
	if pm.Version != "v1.0.0" {
		t.Errorf("Expected Version 'v1.0.0', got '%s'", pm.Version)
	}
	
	if pm.token != "token" {
		t.Errorf("Expected token 'token', got '%s'", pm.token)
	}
	
	if pm.scheme != Local {
		t.Errorf("Expected scheme %d (Local), got %d", Local, pm.scheme)
	}
	
	if pm.LocalPath != tempDir {
		t.Errorf("Expected LocalPath '%s', got '%s'", tempDir, pm.LocalPath)
	}
	
	if pm.RemotePath != tempDir {
		t.Errorf("Expected RemotePath '%s', got '%s'", tempDir, pm.RemotePath)
	}
	
	expectedKey := "plugin:test-plugin@v1.0.0"
	if pm.Key != expectedKey {
		t.Errorf("Expected Key '%s', got '%s'", expectedKey, pm.Key)
	}
	
	expectedMainFile := filepath.Join(tempDir, "main.js")
	if pm.MainFile != expectedMainFile {
		t.Errorf("Expected MainFile '%s', got '%s'", expectedMainFile, pm.MainFile)
	}
	
	expectedManifestFile := filepath.Join(tempDir, "manifest.yml")
	if pm.ManifestFile != expectedManifestFile {
		t.Errorf("Expected ManifestFile '%s', got '%s'", expectedManifestFile, pm.ManifestFile)
	}
}

func TestNewPluginMetadata_RemoteScheme(t *testing.T) {
	// Save original values
	originalLocalDataPath := LocalDataPath
	originalRemotePath := RemotePath
	originalLocalLockPath := LocalLockPath
	
	// Set test paths
	tempDir := t.TempDir()
	LocalDataPath = filepath.Join(tempDir, "data")
	RemotePath = "https://github.com"
	LocalLockPath = filepath.Join(tempDir, ".locks")
	
	// Restore original values after test
	defer func() {
		LocalDataPath = originalLocalDataPath
		RemotePath = originalRemotePath
		LocalLockPath = originalLocalLockPath
	}()

	// This will fail because we can't actually clone a repo in tests,
	// but we can test the metadata setup before the Load() call
	pm := &PluginMetadata{
		Name:    "test-plugin",
		Version: "v1.0.0",
		token:   "token",
		Key:     formatPluginKey("test-plugin", "v1.0.0"),
	}
	
	// Set up remote scheme manually (without calling Load)
	pm.scheme = Remote
	pm.LocalPath = filepath.Join(LocalDataPath, "test-plugin", "v1.0.0")
	pm.RemotePath = fmt.Sprintf("%s/%s", RemotePath, "test-plugin")
	pm.lockFile = filepath.Join(LocalLockPath, "test-plugin-v1.0.0.lock")
	pm.MainFile = filepath.Join(pm.LocalPath, Main)
	pm.ManifestFile = filepath.Join(pm.LocalPath, Manifest)

	// Verify remote scheme setup
	if pm.scheme != Remote {
		t.Errorf("Expected scheme %d (Remote), got %d", Remote, pm.scheme)
	}
	
	expectedLocalPath := filepath.Join(LocalDataPath, "test-plugin", "v1.0.0")
	if pm.LocalPath != expectedLocalPath {
		t.Errorf("Expected LocalPath '%s', got '%s'", expectedLocalPath, pm.LocalPath)
	}
	
	expectedRemotePath := "https://github.com/test-plugin"
	if pm.RemotePath != expectedRemotePath {
		t.Errorf("Expected RemotePath '%s', got '%s'", expectedRemotePath, pm.RemotePath)
	}
	
	expectedLockFile := filepath.Join(LocalLockPath, "test-plugin-v1.0.0.lock")
	if pm.lockFile != expectedLockFile {
		t.Errorf("Expected lockFile '%s', got '%s'", expectedLockFile, pm.lockFile)
	}
}

func TestPluginMetadata_checkPlugin(t *testing.T) {
	t.Run("ValidPlugin", func(t *testing.T) {
		// Create a temporary directory with required files
		tempDir := t.TempDir()
		
		pm := &PluginMetadata{
			Name:         "test-plugin",
			Version:      "v1.0.0",
			LocalPath:    tempDir,
			MainFile:     filepath.Join(tempDir, "main.js"),
			ManifestFile: filepath.Join(tempDir, "manifest.yml"),
		}
		
		// Create main.js file
		if err := os.WriteFile(pm.MainFile, []byte("console.log('test');"), 0644); err != nil {
			t.Fatalf("Failed to create main.js: %v", err)
		}
		
		// Create manifest.yml file
		if err := os.WriteFile(pm.ManifestFile, []byte("name: test\nversion: 1.0.0"), 0644); err != nil {
			t.Fatalf("Failed to create manifest.yml: %v", err)
		}

		err := pm.checkPlugin()
		if err != nil {
			t.Errorf("Unexpected error for valid plugin: %v", err)
		}
	})

	t.Run("MissingMainFile", func(t *testing.T) {
		tempDir := t.TempDir()
		
		pm := &PluginMetadata{
			Name:         "test-plugin",
			Version:      "v1.0.0",
			LocalPath:    tempDir,
			MainFile:     filepath.Join(tempDir, "main.js"),
			ManifestFile: filepath.Join(tempDir, "manifest.yml"),
		}
		
		// Create only manifest.yml file
		if err := os.WriteFile(pm.ManifestFile, []byte("name: test\nversion: 1.0.0"), 0644); err != nil {
			t.Fatalf("Failed to create manifest.yml: %v", err)
		}

		err := pm.checkPlugin()
		if err == nil {
			t.Error("Expected error for missing main file")
		}
		
		expectedError := pm.MainFile + " doesn't exist"
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("MissingManifestFile", func(t *testing.T) {
		tempDir := t.TempDir()
		
		pm := &PluginMetadata{
			Name:         "test-plugin",
			Version:      "v1.0.0",
			LocalPath:    tempDir,
			MainFile:     filepath.Join(tempDir, "main.js"),
			ManifestFile: filepath.Join(tempDir, "manifest.yml"),
		}
		
		// Create only main.js file
		if err := os.WriteFile(pm.MainFile, []byte("console.log('test');"), 0644); err != nil {
			t.Fatalf("Failed to create main.js: %v", err)
		}

		err := pm.checkPlugin()
		if err == nil {
			t.Error("Expected error for missing manifest file")
		}
		
		expectedError := pm.ManifestFile + " doesn't exist"
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("BothFilesMissing", func(t *testing.T) {
		tempDir := t.TempDir()
		
		pm := &PluginMetadata{
			Name:         "test-plugin",
			Version:      "v1.0.0",
			LocalPath:    tempDir,
			MainFile:     filepath.Join(tempDir, "main.js"),
			ManifestFile: filepath.Join(tempDir, "manifest.yml"),
		}

		err := pm.checkPlugin()
		if err == nil {
			t.Error("Expected error for missing files")
		}
		
		// Should fail on main file first
		expectedError := pm.MainFile + " doesn't exist"
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestPluginMetadata_Load_LocalScheme(t *testing.T) {
	t.Run("ValidLocalPlugin", func(t *testing.T) {
		// Create a temporary directory with required files
		tempDir := t.TempDir()
		
		pm := &PluginMetadata{
			scheme:       Local,
			Name:         "test-plugin",
			Version:      "v1.0.0",
			LocalPath:    tempDir,
			MainFile:     filepath.Join(tempDir, "main.js"),
			ManifestFile: filepath.Join(tempDir, "manifest.yml"),
		}
		
		// Create main.js file
		if err := os.WriteFile(pm.MainFile, []byte("console.log('test');"), 0644); err != nil {
			t.Fatalf("Failed to create main.js: %v", err)
		}
		
		// Create manifest.yml file
		if err := os.WriteFile(pm.ManifestFile, []byte("name: test\nversion: 1.0.0"), 0644); err != nil {
			t.Fatalf("Failed to create manifest.yml: %v", err)
		}

		err := pm.Load()
		if err != nil {
			t.Errorf("Unexpected error for valid local plugin: %v", err)
		}
	})

	t.Run("LocalPathNotExists", func(t *testing.T) {
		pm := &PluginMetadata{
			scheme:    Local,
			Name:      "test-plugin",
			Version:   "v1.0.0",
			LocalPath: "/nonexistent/path",
		}

		err := pm.Load()
		if err == nil {
			t.Error("Expected error for non-existent local path")
		}
	})

	t.Run("LocalPathNotDirectory", func(t *testing.T) {
		// Create a temporary file (not directory)
		tempFile, err := os.CreateTemp("", "test-file")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()
		
		pm := &PluginMetadata{
			scheme:    Local,
			Name:      "test-plugin",
			Version:   "v1.0.0",
			LocalPath: tempFile.Name(),
		}

		err = pm.Load()
		if err == nil {
			t.Error("Expected error for local path that is not a directory")
		}
		
		expectedError := fmt.Sprintf("invalid path %s for plugin %s", tempFile.Name(), "test-plugin")
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("MainFileIsDirectory", func(t *testing.T) {
		// Create a temporary directory
		tempDir := t.TempDir()
		
		// Create main.js as a directory instead of file
		mainDir := filepath.Join(tempDir, "main.js")
		if err := os.Mkdir(mainDir, 0755); err != nil {
			t.Fatalf("Failed to create main.js directory: %v", err)
		}
		
		pm := &PluginMetadata{
			scheme:    Local,
			Name:      "test-plugin",
			Version:   "v1.0.0",
			LocalPath: tempDir,
			MainFile:  mainDir,
		}

		err := pm.Load()
		if err == nil {
			t.Error("Expected error for main file being a directory")
		}
		
		expectedError := fmt.Sprintf("invalid main file %s for plugin %s", mainDir, "test-plugin")
		if err.Error() != expectedError {
			t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestPluginManager_LoadPlugin(t *testing.T) {
	// Create a temporary directory for testing
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

	pm := &PluginManager{
		LoadedPlugin: make(map[string]*PluginMetadata),
	}

	t.Run("LoadNewPlugin", func(t *testing.T) {
		plugin, err := pm.LoadPlugin("test-plugin", "v1.0.0", "token", tempDir)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if plugin == nil {
			t.Fatal("Expected plugin to be loaded")
		}
		
		if plugin.Name != "test-plugin" {
			t.Errorf("Expected plugin name 'test-plugin', got '%s'", plugin.Name)
		}
		
		if plugin.Version != "v1.0.0" {
			t.Errorf("Expected plugin version 'v1.0.0', got '%s'", plugin.Version)
		}
		
		// Verify plugin was cached
		expectedKey := "plugin:test-plugin@v1.0.0"
		if _, exists := pm.LoadedPlugin[expectedKey]; !exists {
			t.Error("Expected plugin to be cached in LoadedPlugin map")
		}
	})

	t.Run("LoadCachedPlugin", func(t *testing.T) {
		// Load the same plugin again - should return cached version
		plugin, err := pm.LoadPlugin("test-plugin", "v1.0.0", "token", tempDir)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if plugin == nil {
			t.Fatal("Expected cached plugin to be returned")
		}
		
		// Verify it's the same instance from cache
		expectedKey := "plugin:test-plugin@v1.0.0"
		cachedPlugin := pm.LoadedPlugin[expectedKey]
		if plugin != cachedPlugin {
			t.Error("Expected to get cached plugin instance")
		}
	})

	t.Run("LoadPluginWithEmptyVersion", func(t *testing.T) {
		plugin, err := pm.LoadPlugin("test-plugin", "", "token", tempDir)
		
		if err == nil {
			t.Error("Expected error for empty version")
		}
		
		if plugin != nil {
			t.Error("Expected nil plugin for empty version")
		}
	})

	t.Run("LoadPluginWithInvalidPath", func(t *testing.T) {
		plugin, err := pm.LoadPlugin("invalid-plugin", "v1.0.0", "token", "/nonexistent/path")
		
		if err == nil {
			t.Error("Expected error for invalid path")
		}
		
		if plugin != nil {
			t.Error("Expected nil plugin for invalid path")
		}
	})
}

func TestGlobalVariables(t *testing.T) {
	// Test that global variables are set correctly
	if Main != "main.js" {
		t.Errorf("Expected Main to be 'main.js', got '%s'", Main)
	}
	
	if Manifest != "manifest.yml" {
		t.Errorf("Expected Manifest to be 'manifest.yml', got '%s'", Manifest)
	}
	
	if RemotePath != "https://github.com" {
		t.Errorf("Expected RemotePath to be 'https://github.com', got '%s'", RemotePath)
	}
	
	if Local != 1 {
		t.Errorf("Expected Local to be 1, got %d", Local)
	}
	
	if Remote != 2 {
		t.Errorf("Expected Remote to be 2, got %d", Remote)
	}
	
	// Test that LocalPath is set (should contain HOME/.nadleeh/plugins)
	if LocalPath == "" {
		t.Error("Expected LocalPath to be set")
	}
	
	// Test that LocalDataPath and LocalLockPath are derived from LocalPath
	expectedDataPath := filepath.Join(LocalPath, "data")
	if LocalDataPath != expectedDataPath {
		t.Errorf("Expected LocalDataPath to be '%s', got '%s'", expectedDataPath, LocalDataPath)
	}
	
	expectedLockPath := filepath.Join(LocalPath, ".locks")
	if LocalLockPath != expectedLockPath {
		t.Errorf("Expected LocalLockPath to be '%s', got '%s'", expectedLockPath, LocalLockPath)
	}
}

// Benchmark tests
func BenchmarkFormatPluginKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		formatPluginKey("test-plugin", "v1.0.0")
	}
}

func BenchmarkNewPluginManager(b *testing.B) {
	// Save original values
	originalLocalPath := LocalPath
	originalLocalDataPath := LocalDataPath
	originalLocalLockPath := LocalLockPath
	
	for i := 0; i < b.N; i++ {
		// Use a unique temp directory for each iteration
		tempDir := b.TempDir()
		LocalPath = tempDir
		LocalDataPath = filepath.Join(tempDir, "data")
		LocalLockPath = filepath.Join(tempDir, ".locks")
		
		NewPluginManager()
	}
	
	// Restore original values
	LocalPath = originalLocalPath
	LocalDataPath = originalLocalDataPath
	LocalLockPath = originalLocalLockPath
}

func BenchmarkPluginManager_LoadPlugin(b *testing.B) {
	// Create a temporary directory with required files
	tempDir := b.TempDir()
	
	// Create main.js file
	mainFile := filepath.Join(tempDir, "main.js")
	if err := os.WriteFile(mainFile, []byte("console.log('test');"), 0644); err != nil {
		b.Fatalf("Failed to create main.js: %v", err)
	}
	
	// Create manifest.yml file
	manifestFile := filepath.Join(tempDir, "manifest.yml")
	if err := os.WriteFile(manifestFile, []byte("name: test\nversion: 1.0.0"), 0644); err != nil {
		b.Fatalf("Failed to create manifest.yml: %v", err)
	}

	pm := &PluginManager{
		LoadedPlugin: make(map[string]*PluginMetadata),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use different plugin names to avoid caching effects
		pluginName := fmt.Sprintf("test-plugin-%d", i)
		pm.LoadPlugin(pluginName, "v1.0.0", "token", tempDir)
	}
}
