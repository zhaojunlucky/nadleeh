package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirExists(t *testing.T) {
	t.Run("ExistingDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		
		exists, err := DirExists(tempDir)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !exists {
			t.Error("Expected directory to exist")
		}
	})
	
	t.Run("NonExistentDirectory", func(t *testing.T) {
		nonExistentPath := "/path/that/does/not/exist"
		
		exists, err := DirExists(nonExistentPath)
		if err != nil {
			t.Fatalf("Expected no error for non-existent path, got: %v", err)
		}
		if exists {
			t.Error("Expected directory to not exist")
		}
	})
	
	t.Run("PathIsFile", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "testfile.txt")
		
		// Create a test file
		err := os.WriteFile(tempFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		exists, err := DirExists(tempFile)
		if err == nil {
			t.Error("Expected error when path is a file")
		}
		if exists {
			t.Error("Expected exists to be false when path is a file")
		}
		if !strings.Contains(err.Error(), "is a file") {
			t.Errorf("Expected error message to contain 'is a file', got: %v", err)
		}
	})
	
	t.Run("PermissionDenied", func(t *testing.T) {
		// This test might not work on all systems, so we'll skip it if we can't create the scenario
		tempDir := t.TempDir()
		restrictedDir := filepath.Join(tempDir, "restricted")
		
		err := os.Mkdir(restrictedDir, 0000) // No permissions
		if err != nil {
			t.Skip("Cannot create restricted directory for test")
		}
		defer os.Chmod(restrictedDir, 0755) // Restore permissions for cleanup
		
		// Try to check a subdirectory of the restricted directory
		testPath := filepath.Join(restrictedDir, "subdir")
		_, err = DirExists(testPath)
		if err == nil {
			t.Skip("Expected permission error, but got none (system may not enforce permissions)")
		}
	})
}

func TestFileExists(t *testing.T) {
	t.Run("ExistingFile", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "testfile.txt")
		
		// Create a test file
		err := os.WriteFile(tempFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		exists, err := FileExists(tempFile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !exists {
			t.Error("Expected file to exist")
		}
	})
	
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentPath := "/path/that/does/not/exist.txt"
		
		exists, err := FileExists(nonExistentPath)
		if err != nil {
			t.Fatalf("Expected no error for non-existent file, got: %v", err)
		}
		if exists {
			t.Error("Expected file to not exist")
		}
	})
	
	t.Run("PathIsDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		
		exists, err := FileExists(tempDir)
		if err == nil {
			t.Error("Expected error when path is a directory")
		}
		if exists {
			t.Error("Expected exists to be false when path is a directory")
		}
		if !strings.Contains(err.Error(), "is a file") {
			t.Errorf("Expected error message to contain 'is a file', got: %v", err)
		}
	})
	
	t.Run("EmptyFile", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "empty.txt")
		
		// Create an empty file
		err := os.WriteFile(tempFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}
		
		exists, err := FileExists(tempFile)
		if err != nil {
			t.Fatalf("Expected no error for empty file, got: %v", err)
		}
		if !exists {
			t.Error("Expected empty file to exist")
		}
	})
}

func TestLogFileWithLineNo(t *testing.T) {
	t.Run("ValidFile", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "testfile.txt")
		content := "line 1\nline 2\nline 3"
		
		err := os.WriteFile(tempFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		err = LogFileWithLineNo("test", tempFile)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		// Note: This function prints to stdout, so we can't easily capture and test the output
		// In a real-world scenario, you might want to refactor this to accept an io.Writer
	})
	
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentFile := "/path/that/does/not/exist.txt"
		
		err := LogFileWithLineNo("test", nonExistentFile)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
	
	t.Run("EmptyFile", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "empty.txt")
		
		err := os.WriteFile(tempFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}
		
		err = LogFileWithLineNo("empty", tempFile)
		if err != nil {
			t.Errorf("Expected no error for empty file, got: %v", err)
		}
	})
}

func TestLogStrWithLineNo(t *testing.T) {
	t.Run("MultiLineString", func(t *testing.T) {
		content := "line 1\nline 2\nline 3"
		
		err := LogStrWithLineNo("test", content)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		// Note: This function prints to stdout, so we can't easily capture and test the output
	})
	
	t.Run("EmptyString", func(t *testing.T) {
		err := LogStrWithLineNo("empty", "")
		if err != nil {
			t.Errorf("Expected no error for empty string, got: %v", err)
		}
	})
	
	t.Run("SingleLineString", func(t *testing.T) {
		err := LogStrWithLineNo("single", "single line")
		if err != nil {
			t.Errorf("Expected no error for single line, got: %v", err)
		}
	})
	
	t.Run("StringWithEmptyLines", func(t *testing.T) {
		content := "line 1\n\nline 3\n\nline 5"
		
		err := LogStrWithLineNo("test", content)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})
	
	t.Run("StringEndingWithNewline", func(t *testing.T) {
		content := "line 1\nline 2\n"
		
		err := LogStrWithLineNo("test", content)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})
}

func TestGetProjectRootDir(t *testing.T) {
	t.Run("ValidGoModule", func(t *testing.T) {
		// This test assumes we're running in a valid Go module
		rootDir, err := GetProjectRootDir()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if rootDir == "" {
			t.Error("Expected non-empty root directory")
		}
		
		// Verify the returned path exists and is a directory
		exists, err := DirExists(rootDir)
		if err != nil {
			t.Fatalf("Error checking if root directory exists: %v", err)
		}
		if !exists {
			t.Error("Expected root directory to exist")
		}
		
		// Verify go.mod exists in the root directory
		goModPath := filepath.Join(rootDir, "go.mod")
		exists, err = FileExists(goModPath)
		if err != nil {
			t.Fatalf("Error checking if go.mod exists: %v", err)
		}
		if !exists {
			t.Error("Expected go.mod to exist in root directory")
		}
	})
	
	t.Run("AbsolutePath", func(t *testing.T) {
		rootDir, err := GetProjectRootDir()
		if err != nil {
			t.Skip("Cannot get project root for absolute path test")
		}
		
		if !filepath.IsAbs(rootDir) {
			t.Error("Expected absolute path for project root")
		}
	})
}

// Edge case tests
func TestEdgeCases(t *testing.T) {
	t.Run("DirExists_EmptyPath", func(t *testing.T) {
		exists, err := DirExists("")
		// Empty path behavior can vary by system, but exists should always be false
		if exists {
			t.Error("Expected exists to be false for empty path")
		}
		// On some systems, empty path might return an error, on others it might not
		// The important thing is that exists is false
		t.Logf("DirExists(\"\") returned exists=%v, err=%v", exists, err)
	})
	
	t.Run("FileExists_EmptyPath", func(t *testing.T) {
		exists, err := FileExists("")
		// Empty path behavior can vary by system, but exists should always be false
		if exists {
			t.Error("Expected exists to be false for empty path")
		}
		// On some systems, empty path might return an error, on others it might not
		// The important thing is that exists is false
		t.Logf("FileExists(\"\") returned exists=%v, err=%v", exists, err)
	})
	
	t.Run("SymbolicLinks", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Create a regular file
		regularFile := filepath.Join(tempDir, "regular.txt")
		err := os.WriteFile(regularFile, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create regular file: %v", err)
		}
		
		// Create a symbolic link to the file
		symLink := filepath.Join(tempDir, "symlink.txt")
		err = os.Symlink(regularFile, symLink)
		if err != nil {
			t.Skip("Cannot create symbolic link on this system")
		}
		
		// Test FileExists with symbolic link
		exists, err := FileExists(symLink)
		if err != nil {
			t.Errorf("Expected no error for symbolic link to file, got: %v", err)
		}
		if !exists {
			t.Error("Expected symbolic link to file to exist")
		}
		
		// Create a directory and a symbolic link to it
		regularDir := filepath.Join(tempDir, "regulardir")
		err = os.Mkdir(regularDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create regular directory: %v", err)
		}
		
		symLinkDir := filepath.Join(tempDir, "symlinkdir")
		err = os.Symlink(regularDir, symLinkDir)
		if err != nil {
			t.Skip("Cannot create symbolic link to directory on this system")
		}
		
		// Test DirExists with symbolic link to directory
		exists, err = DirExists(symLinkDir)
		if err != nil {
			t.Errorf("Expected no error for symbolic link to directory, got: %v", err)
		}
		if !exists {
			t.Error("Expected symbolic link to directory to exist")
		}
	})
}

// Benchmark tests
func BenchmarkDirExists(b *testing.B) {
	tempDir := b.TempDir()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DirExists(tempDir)
	}
}

func BenchmarkFileExists(b *testing.B) {
	tempDir := b.TempDir()
	tempFile := filepath.Join(tempDir, "benchmark.txt")
	os.WriteFile(tempFile, []byte("benchmark content"), 0644)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FileExists(tempFile)
	}
}

func BenchmarkLogStrWithLineNo(b *testing.B) {
	content := strings.Repeat("line content\n", 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogStrWithLineNo("benchmark", content)
	}
}

func BenchmarkGetProjectRootDir(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetProjectRootDir()
	}
}
