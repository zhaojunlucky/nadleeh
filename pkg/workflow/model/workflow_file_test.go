package workflow

import (
	"encoding/base64"
	"fmt"
	"io"
	"nadleeh/pkg/workflow/core"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akamensky/argparse"
)

// Mock argparse.Arg for testing
type mockArg struct {
	value   interface{}
	parsed  bool
	happened bool
}

func (m *mockArg) GetParsed() bool {
	return m.parsed
}

func (m *mockArg) GetResult() interface{} {
	return m.value
}

func (m *mockArg) GetLname() string {
	return "test"
}

func (m *mockArg) GetSname() string {
	return "t"
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

func (m *mockArg) GetHappened() bool {
	return m.happened
}

func (m *mockArg) GetRemainder() []string {
	return nil
}

func (m *mockArg) GetPositional() bool {
	return false
}


func TestWorkflowProvider_Download(t *testing.T) {
	t.Run("GitHubProvider", func(t *testing.T) {
		// Test with @ prefix (default repository)
		// Note: This will make actual GitHub API calls, so we skip it in unit tests
		t.Skip("Skipping GitHub API test to avoid external dependencies")
	})

	t.Run("HTTPSProvider", func(t *testing.T) {
		// This test would require mocking HTTP client, so we test the logic separately
		t.Skip("Skipping HTTPS download test to avoid external dependencies")
	})

	t.Run("UnsupportedProvider", func(t *testing.T) {
		provider := &workflowProvider{
			Type: "unsupported",
		}

		_, err := provider.Download("test.yml")
		if err == nil {
			t.Error("Expected error for unsupported provider type")
		}
		if !strings.Contains(err.Error(), "unsupported provider type") {
			t.Errorf("Expected unsupported provider error, got: %v", err)
		}
	})
}

func TestWorkflowProvider_DownloadHTTP(t *testing.T) {
	t.Run("ValidHTTPSServer", func(t *testing.T) {
		// Test URL construction logic
		// We can't test actual download without mocking HTTP client
		t.Skip("Skipping actual HTTP download test")
	})

	t.Run("InvalidServer_Empty", func(t *testing.T) {
		provider := &workflowProvider{
			Type:   httpsProvider,
			Server: "",
		}

		_, err := provider.downloadHTTP("test.yml")
		if err == nil {
			t.Error("Expected error for empty server")
		}
		if !strings.Contains(err.Error(), "server is required") {
			t.Errorf("Expected server required error, got: %v", err)
		}
	})

	t.Run("InvalidServer_NotHTTPS", func(t *testing.T) {
		provider := &workflowProvider{
			Type:   httpsProvider,
			Server: "http://example.com",
		}

		_, err := provider.downloadHTTP("test.yml")
		if err == nil {
			t.Error("Expected error for non-HTTPS server")
		}
		if !strings.Contains(err.Error(), "server is required") {
			t.Errorf("Expected server required error, got: %v", err)
		}
	})
}

func TestWorkflowProvider_DownloadGitHub(t *testing.T) {
	t.Run("DefaultRepository_WithAtPrefix", func(t *testing.T) {
		// Test path parsing logic for @ prefix
		// We can't test actual GitHub API without mocking, but we can test error cases
		t.Skip("Skipping GitHub API test to avoid external dependencies")
	})

	t.Run("CustomRepository_ValidPath", func(t *testing.T) {
		// Test path parsing logic for custom repository
		// We can't test actual GitHub API without mocking
		t.Skip("Skipping GitHub API test to avoid external dependencies")
	})

	t.Run("InvalidPath_TooFewSegments", func(t *testing.T) {
		provider := &workflowProvider{
			Type: githubProvider,
			Cred: workflowCred{
				Type: "",
			},
		}

		_, err := provider.downloadGitHub("owner/repo")
		if err == nil {
			t.Error("Expected error for invalid GitHub path")
		}
		if !strings.Contains(err.Error(), "invalid github workflow file") {
			t.Errorf("Expected invalid path error, got: %v", err)
		}
	})

	t.Run("InvalidCredType", func(t *testing.T) {
		provider := &workflowProvider{
			Type:  githubProvider,
			Owner: "testowner",
			Name:  "testrepo",
			Cred: workflowCred{
				Type: "invalid",
			},
		}

		_, err := provider.downloadGitHub("@test.yml")
		if err == nil {
			t.Error("Expected error for invalid cred type")
		}
		if !strings.Contains(err.Error(), "only Bearer or empty cred type supported") {
			t.Errorf("Expected cred type error, got: %v", err)
		}
	})
}

func TestWorkflowProvider_DownloadURL(t *testing.T) {
	t.Run("InvalidURL", func(t *testing.T) {
		provider := &workflowProvider{}

		testCases := []string{
			"invalid-url",
			"ftp://example.com/file.yml",
			"not a url at all",
			"",
		}

		for _, url := range testCases {
			_, err := provider.downloadURL(url)
			if err == nil {
				t.Errorf("Expected error for invalid URL: %s", url)
			}
			if !strings.Contains(err.Error(), "invalid url") {
				t.Errorf("Expected invalid URL error for %s, got: %v", url, err)
			}
		}
	})

	t.Run("ValidURL_HTTPSFormat", func(t *testing.T) {
		// Test URL validation logic
		// We can't test actual HTTP requests without mocking
		// But we can verify the URL passes validation
		t.Skip("Skipping actual HTTP request test")
	})
}

func TestWorkflowProvider_AddHTTPHeader(t *testing.T) {
	t.Run("BearerAuth", func(t *testing.T) {
		provider := &workflowProvider{
			Cred: workflowCred{
				Type:     bearer,
				Password: "test-token",
			},
		}

		req, err := http.NewRequest("GET", "https://example.com", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		err = provider.addHTTPHeader(req)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		expected := "Bearer test-token"
		if authHeader != expected {
			t.Errorf("Expected Authorization header '%s', got '%s'", expected, authHeader)
		}
	})

	t.Run("BasicAuth", func(t *testing.T) {
		provider := &workflowProvider{
			Cred: workflowCred{
				Type:     basic,
				Username: "testuser",
				Password: "testpass",
			},
		}

		req, err := http.NewRequest("GET", "https://example.com", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		err = provider.addHTTPHeader(req)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		expectedAuth := base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
		expected := fmt.Sprintf("Basic %s", expectedAuth)
		if authHeader != expected {
			t.Errorf("Expected Authorization header '%s', got '%s'", expected, authHeader)
		}
	})

	t.Run("NoAuth", func(t *testing.T) {
		provider := &workflowProvider{
			Cred: workflowCred{
				Type: "",
			},
		}

		req, err := http.NewRequest("GET", "https://example.com", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		err = provider.addHTTPHeader(req)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		if authHeader != "" {
			t.Errorf("Expected no Authorization header, got '%s'", authHeader)
		}
	})
}

func TestLoadWorkflowFile(t *testing.T) {
	t.Run("LocalFile_ValidYAML", func(t *testing.T) {
		// Create a temporary YAML file
		tempDir := t.TempDir()
		yamlFile := filepath.Join(tempDir, "test.yml")
		yamlContent := `
name: "test-workflow"
jobs:
  test-job:
    steps:
      - name: "test-step"
        run: "echo test"
`
		if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to create test YAML file: %v", err)
		}

		args := map[string]argparse.Arg{
			"provider": &mockArg{value: nil, parsed: false, happened: false},
		}
		workflowArgs := core.NewWorkflowArgs(args)

		reader, err := LoadWorkflowFile(yamlFile, workflowArgs)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if reader == nil {
			t.Error("Expected non-nil reader")
		}

		// Verify content
		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read content: %v", err)
		}
		if !strings.Contains(string(content), "test-workflow") {
			t.Error("Expected workflow content not found")
		}
	})

	t.Run("LocalFile_InvalidExtension", func(t *testing.T) {
		// This should trigger log.Fatal, so we skip it
		t.Skip("Skipping test that would trigger log.Fatal for invalid extension")
	})

	t.Run("LocalFile_NonexistentFile", func(t *testing.T) {
		// This should trigger log.Fatal, so we skip it
		t.Skip("Skipping test that would trigger log.Fatal for nonexistent file")
	})

	t.Run("LocalFile_Directory", func(t *testing.T) {
		// This should trigger log.Fatal, so we skip it
		t.Skip("Skipping test that would trigger log.Fatal for directory")
	})

	t.Run("WithProvider_EmptyProvider", func(t *testing.T) {
		emptyProvider := ""
		args := map[string]argparse.Arg{
			"provider": &mockArg{value: &emptyProvider, parsed: true, happened: true},
		}
		workflowArgs := core.NewWorkflowArgs(args)

		_, err := LoadWorkflowFile("test.yml", workflowArgs)
		if err == nil {
			t.Error("Expected error for empty provider")
		}
		if !strings.Contains(err.Error(), "provider is empty") {
			t.Errorf("Expected empty provider error, got: %v", err)
		}
	})

	t.Run("WithProvider_GitHubDefault", func(t *testing.T) {
		// This would use default GitHub provider and make API calls
		t.Skip("Skipping GitHub provider test to avoid external dependencies")
	})

	t.Run("WithProvider_CustomProvider", func(t *testing.T) {
		// Create a temporary provider file
		tempDir := t.TempDir()
		
		// Mock user home directory by creating the provider file structure
		providerDir := filepath.Join(tempDir, ".nadleeh", "providers")
		if err := os.MkdirAll(providerDir, 0755); err != nil {
			t.Fatalf("Failed to create provider directory: %v", err)
		}

		providerFile := filepath.Join(providerDir, "custom")
		providerContent := `
type: "https"
server: "https://example.com"
cred:
  type: "Bearer"
  password: "test-token"
`
		if err := os.WriteFile(providerFile, []byte(providerContent), 0644); err != nil {
			t.Fatalf("Failed to create provider file: %v", err)
		}

		// This would require mocking user.Current() and file operations
		t.Skip("Skipping custom provider test due to user.Current() dependency")
	})

	t.Run("WithProvider_InvalidProviderFile", func(t *testing.T) {
		// This would require mocking user.Current() and file operations
		t.Skip("Skipping invalid provider test due to user.Current() dependency")
	})

	t.Run("ArgumentParsingError", func(t *testing.T) {
		// This test would require mocking argument.GetStringFromArg to return an error
		t.Skip("Skipping argument parsing error test due to external dependency")
	})
}

func TestWorkflowCredStruct(t *testing.T) {
	t.Run("StructFields", func(t *testing.T) {
		cred := workflowCred{
			Type:     "Bearer",
			Username: "testuser",
			Password: "testpass",
		}

		if cred.Type != "Bearer" {
			t.Errorf("Expected Type 'Bearer', got '%s'", cred.Type)
		}
		if cred.Username != "testuser" {
			t.Errorf("Expected Username 'testuser', got '%s'", cred.Username)
		}
		if cred.Password != "testpass" {
			t.Errorf("Expected Password 'testpass', got '%s'", cred.Password)
		}
	})
}

func TestWorkflowProviderStruct(t *testing.T) {
	t.Run("StructFields", func(t *testing.T) {
		provider := workflowProvider{
			Type:   "github",
			Server: "https://github.com",
			Cred: workflowCred{
				Type:     "Bearer",
				Username: "user",
				Password: "token",
			},
		}

		if provider.Type != "github" {
			t.Errorf("Expected Type 'github', got '%s'", provider.Type)
		}
		if provider.Server != "https://github.com" {
			t.Errorf("Expected Server 'https://github.com', got '%s'", provider.Server)
		}
		if provider.Cred.Type != "Bearer" {
			t.Errorf("Expected Cred.Type 'Bearer', got '%s'", provider.Cred.Type)
		}
	})
}

func TestDefaultGitHubProvider(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		if defaultGitHubProvider.Type != githubProvider {
			t.Errorf("Expected Type '%s', got '%s'", githubProvider, defaultGitHubProvider.Type)
		}
		if defaultGitHubProvider.Server != "https://github.com" {
			t.Errorf("Expected Server 'https://github.com', got '%s'", defaultGitHubProvider.Server)
		}
		if defaultGitHubProvider.Cred.Type != "" {
			t.Errorf("Expected empty Cred.Type, got '%s'", defaultGitHubProvider.Cred.Type)
		}
	})
}

func TestGlobalVariables(t *testing.T) {
	t.Run("ConstantValues", func(t *testing.T) {
		if bearer != "Bearer" {
			t.Errorf("Expected bearer 'Bearer', got '%s'", bearer)
		}
		if basic != "Basic" {
			t.Errorf("Expected basic 'Basic', got '%s'", basic)
		}
		if githubProvider != "github" {
			t.Errorf("Expected githubProvider 'github', got '%s'", githubProvider)
		}
		if httpsProvider != "https" {
			t.Errorf("Expected httpsProvider 'https', got '%s'", httpsProvider)
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkWorkflowProvider_AddHTTPHeader_Bearer(b *testing.B) {
	provider := &workflowProvider{
		Cred: workflowCred{
			Type:     bearer,
			Password: "test-token",
		},
	}

	req, _ := http.NewRequest("GET", "https://example.com", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.addHTTPHeader(req)
	}
}

func BenchmarkWorkflowProvider_AddHTTPHeader_Basic(b *testing.B) {
	provider := &workflowProvider{
		Cred: workflowCred{
			Type:     basic,
			Username: "testuser",
			Password: "testpass",
		},
	}

	req, _ := http.NewRequest("GET", "https://example.com", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.addHTTPHeader(req)
	}
}

func BenchmarkLoadWorkflowFile_LocalFile(b *testing.B) {
	// Create a temporary YAML file
	tempDir := b.TempDir()
	yamlFile := filepath.Join(tempDir, "benchmark.yml")
	yamlContent := `
name: "benchmark-workflow"
jobs:
  test-job:
    steps:
      - name: "test-step"
        run: "echo test"
`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		b.Fatalf("Failed to create test YAML file: %v", err)
	}

	args := map[string]argparse.Arg{
		"provider": &mockArg{value: nil},
	}
	workflowArgs := core.NewWorkflowArgs(args)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader, err := LoadWorkflowFile(yamlFile, workflowArgs)
		if err != nil {
			b.Fatalf("LoadWorkflowFile failed: %v", err)
		}
		if reader != nil {
			// Close the reader if it's a file
			if file, ok := reader.(*os.File); ok {
				file.Close()
			}
		}
	}
}
