package script

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNJSFile_ReadFileAsLines(t *testing.T) {
	njsFile := &NJSFile{}

	t.Run("ValidFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		content := "line 1\nline 2\nline 3\n"
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		lines, err := njsFile.ReadFileAsLines(testFile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		expected := []string{"line 1", "line 2", "line 3"}
		if len(lines) != len(expected) {
			t.Fatalf("Expected %d lines, got %d", len(expected), len(lines))
		}

		for i, line := range lines {
			if line != expected[i] {
				t.Errorf("Line %d: expected %q, got %q", i, expected[i], line)
			}
		}
	})

	t.Run("EmptyFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "empty.txt")
		
		err := os.WriteFile(testFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		lines, err := njsFile.ReadFileAsLines(testFile)
		if err != nil {
			t.Fatalf("Expected no error for empty file, got: %v", err)
		}

		if len(lines) != 0 {
			t.Errorf("Expected empty slice for empty file, got %d lines", len(lines))
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		lines, err := njsFile.ReadFileAsLines("/path/that/does/not/exist.txt")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
		if lines != nil {
			t.Error("Expected nil lines for non-existent file")
		}
		if !strings.Contains(err.Error(), "file does not exist") {
			t.Errorf("Expected 'file does not exist' error, got: %v", err)
		}
	})

	t.Run("EmptyPath", func(t *testing.T) {
		lines, err := njsFile.ReadFileAsLines("")
		if err == nil {
			t.Error("Expected error for empty path")
		}
		if lines != nil {
			t.Error("Expected nil lines for empty path")
		}
		if !strings.Contains(err.Error(), "file path cannot be empty") {
			t.Errorf("Expected 'file path cannot be empty' error, got: %v", err)
		}
	})

	t.Run("DirectoryPath", func(t *testing.T) {
		tempDir := t.TempDir()
		
		lines, err := njsFile.ReadFileAsLines(tempDir)
		if err == nil {
			t.Error("Expected error when path is a directory")
		}
		if lines != nil {
			t.Error("Expected nil lines when path is a directory")
		}
		if !strings.Contains(err.Error(), "path is a directory") {
			t.Errorf("Expected 'path is a directory' error, got: %v", err)
		}
	})

	t.Run("FileWithEmptyLines", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "with_empty_lines.txt")
		content := "line 1\n\nline 3\n\nline 5"
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		lines, err := njsFile.ReadFileAsLines(testFile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		expected := []string{"line 1", "", "line 3", "", "line 5"}
		if len(lines) != len(expected) {
			t.Fatalf("Expected %d lines, got %d", len(expected), len(lines))
		}

		for i, line := range lines {
			if line != expected[i] {
				t.Errorf("Line %d: expected %q, got %q", i, expected[i], line)
			}
		}
	})

	t.Run("LargeFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "large.txt")
		
		// Create a file with 1000 lines
		var content strings.Builder
		for i := 0; i < 1000; i++ {
			content.WriteString("This is line ")
			content.WriteString(strings.Repeat("x", 100)) // Make lines longer
			content.WriteString("\n")
		}
		
		err := os.WriteFile(testFile, []byte(content.String()), 0644)
		if err != nil {
			t.Fatalf("Failed to create large test file: %v", err)
		}

		lines, err := njsFile.ReadFileAsLines(testFile)
		if err != nil {
			t.Fatalf("Expected no error for large file, got: %v", err)
		}

		if len(lines) != 1000 {
			t.Errorf("Expected 1000 lines, got %d", len(lines))
		}

		// Check first and last lines
		if !strings.HasPrefix(lines[0], "This is line ") {
			t.Errorf("Unexpected first line: %s", lines[0])
		}
		if !strings.HasPrefix(lines[999], "This is line ") {
			t.Errorf("Unexpected last line: %s", lines[999])
		}
	})
}

func TestNJSFile_ReadFileAsString(t *testing.T) {
	njsFile := &NJSFile{}

	t.Run("ValidFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		content := "Hello, World!\nThis is a test file.\nWith multiple lines."
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := njsFile.ReadFileAsString(testFile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		if *result != content {
			t.Errorf("Expected %q, got %q", content, *result)
		}
	})

	t.Run("EmptyFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "empty.txt")
		
		err := os.WriteFile(testFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		result, err := njsFile.ReadFileAsString(testFile)
		if err != nil {
			t.Fatalf("Expected no error for empty file, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result for empty file")
		}

		if *result != "" {
			t.Errorf("Expected empty string for empty file, got %q", *result)
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		result, err := njsFile.ReadFileAsString("/path/that/does/not/exist.txt")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
		if result != nil {
			t.Error("Expected nil result for non-existent file")
		}
		if !strings.Contains(err.Error(), "file does not exist") {
			t.Errorf("Expected 'file does not exist' error, got: %v", err)
		}
	})

	t.Run("EmptyPath", func(t *testing.T) {
		result, err := njsFile.ReadFileAsString("")
		if err == nil {
			t.Error("Expected error for empty path")
		}
		if result != nil {
			t.Error("Expected nil result for empty path")
		}
		if !strings.Contains(err.Error(), "file path cannot be empty") {
			t.Errorf("Expected 'file path cannot be empty' error, got: %v", err)
		}
	})

	t.Run("DirectoryPath", func(t *testing.T) {
		tempDir := t.TempDir()
		
		result, err := njsFile.ReadFileAsString(tempDir)
		if err == nil {
			t.Error("Expected error when path is a directory")
		}
		if result != nil {
			t.Error("Expected nil result when path is a directory")
		}
		if !strings.Contains(err.Error(), "path is a directory") {
			t.Errorf("Expected 'path is a directory' error, got: %v", err)
		}
	})

	t.Run("FileWithSpecialCharacters", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "special.txt")
		content := "Special chars: áéíóú ñ 中文 🚀 \n\t\r"
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := njsFile.ReadFileAsString(testFile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		if *result != content {
			t.Errorf("Expected %q, got %q", content, *result)
		}
	})

	t.Run("LargeFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "large.txt")
		
		// Create a file larger than 1MB to test chunked reading
		var content strings.Builder
		line := strings.Repeat("This is a long line with lots of content to make the file large. ", 100) + "\n"
		for i := 0; i < 2000; i++ { // This should create a file > 1MB
			content.WriteString(line)
		}
		
		err := os.WriteFile(testFile, []byte(content.String()), 0644)
		if err != nil {
			t.Fatalf("Failed to create large test file: %v", err)
		}

		result, err := njsFile.ReadFileAsString(testFile)
		if err != nil {
			t.Fatalf("Expected no error for large file, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result for large file")
		}

		if *result != content.String() {
			t.Error("Large file content mismatch")
		}

		// Verify the file was actually large enough to trigger chunked reading
		fileInfo, _ := os.Stat(testFile)
		if fileInfo.Size() <= 1024*1024 {
			t.Logf("Warning: test file was only %d bytes, may not have tested chunked reading", fileInfo.Size())
		}
	})

	t.Run("BinaryFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "binary.bin")
		
		// Create binary content with null bytes and other binary data
		binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD, 0x7F, 0x80, 0x81}
		
		err := os.WriteFile(testFile, binaryContent, 0644)
		if err != nil {
			t.Fatalf("Failed to create binary test file: %v", err)
		}

		result, err := njsFile.ReadFileAsString(testFile)
		if err != nil {
			t.Fatalf("Expected no error for binary file, got: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result for binary file")
		}

		// Convert back to bytes to verify
		resultBytes := []byte(*result)
		if len(resultBytes) != len(binaryContent) {
			t.Errorf("Expected %d bytes, got %d", len(binaryContent), len(resultBytes))
		}

		for i, b := range binaryContent {
			if i < len(resultBytes) && resultBytes[i] != b {
				t.Errorf("Byte mismatch at position %d: expected %02x, got %02x", i, b, resultBytes[i])
			}
		}
	})
}

func TestNJSFile_IsFile(t *testing.T) {
	njsFile := &NJSFile{}

	t.Run("ValidFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		isFile, err := njsFile.IsFile(testFile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if !isFile {
			t.Error("Expected true for regular file")
		}
	})

	t.Run("Directory", func(t *testing.T) {
		tempDir := t.TempDir()

		isFile, err := njsFile.IsFile(tempDir)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if isFile {
			t.Error("Expected false for directory")
		}
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		isFile, err := njsFile.IsFile("/path/that/does/not/exist")
		if err == nil {
			t.Error("Expected error for non-existent path")
		}

		if isFile {
			t.Error("Expected false for non-existent path")
		}
	})

	t.Run("EmptyPath", func(t *testing.T) {
		isFile, err := njsFile.IsFile("")
		if err == nil {
			t.Error("Expected error for empty path")
		}

		if isFile {
			t.Error("Expected false for empty path")
		}
	})

	t.Run("SymbolicLinkToFile", func(t *testing.T) {
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

		isFile, err := njsFile.IsFile(symLink)
		if err != nil {
			t.Fatalf("Expected no error for symbolic link to file, got: %v", err)
		}

		if !isFile {
			t.Error("Expected true for symbolic link to file")
		}
	})
}

func TestNJSFile_IsDir(t *testing.T) {
	njsFile := &NJSFile{}

	t.Run("ValidDirectory", func(t *testing.T) {
		tempDir := t.TempDir()

		isDir, err := njsFile.IsDir(tempDir)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if !isDir {
			t.Error("Expected true for directory")
		}
	})

	t.Run("RegularFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		isDir, err := njsFile.IsDir(testFile)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if isDir {
			t.Error("Expected false for regular file")
		}
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		isDir, err := njsFile.IsDir("/path/that/does/not/exist")
		if err == nil {
			t.Error("Expected error for non-existent path")
		}

		if isDir {
			t.Error("Expected false for non-existent path")
		}
	})

	t.Run("EmptyPath", func(t *testing.T) {
		isDir, err := njsFile.IsDir("")
		if err == nil {
			t.Error("Expected error for empty path")
		}

		if isDir {
			t.Error("Expected false for empty path")
		}
	})

	t.Run("SymbolicLinkToDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Create a subdirectory
		subDir := filepath.Join(tempDir, "subdir")
		err := os.Mkdir(subDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		// Create a symbolic link to the directory
		symLink := filepath.Join(tempDir, "symlink_dir")
		err = os.Symlink(subDir, symLink)
		if err != nil {
			t.Skip("Cannot create symbolic link on this system")
		}

		isDir, err := njsFile.IsDir(symLink)
		if err != nil {
			t.Fatalf("Expected no error for symbolic link to directory, got: %v", err)
		}

		if !isDir {
			t.Error("Expected true for symbolic link to directory")
		}
	})
}

func TestNJSFile_DeleteFile(t *testing.T) {
	njsFile := &NJSFile{}

	t.Run("DeleteRegularFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Fatal("Test file should exist before deletion")
		}

		err = njsFile.DeleteFile(testFile)
		if err != nil {
			t.Fatalf("Expected no error deleting file, got: %v", err)
		}

		// Verify file is deleted
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("File should not exist after deletion")
		}
	})

	t.Run("DeleteDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		testDir := filepath.Join(tempDir, "testdir")
		
		err := os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Create a file inside the directory
		testFile := filepath.Join(testDir, "file.txt")
		err = os.WriteFile(testFile, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file in test directory: %v", err)
		}

		// Verify directory exists
		if _, err := os.Stat(testDir); os.IsNotExist(err) {
			t.Fatal("Test directory should exist before deletion")
		}

		err = njsFile.DeleteFile(testDir)
		if err != nil {
			t.Fatalf("Expected no error deleting directory, got: %v", err)
		}

		// Verify directory is deleted
		if _, err := os.Stat(testDir); !os.IsNotExist(err) {
			t.Error("Directory should not exist after deletion")
		}
	})

	t.Run("DeleteNonExistentFile", func(t *testing.T) {
		err := njsFile.DeleteFile("/path/that/does/not/exist")
		// os.RemoveAll doesn't return an error for non-existent paths
		if err != nil {
			t.Errorf("Expected no error for non-existent path, got: %v", err)
		}
	})

	t.Run("DeleteEmptyPath", func(t *testing.T) {
		err := njsFile.DeleteFile("")
		// This should not cause any issues
		if err != nil {
			t.Errorf("Expected no error for empty path, got: %v", err)
		}
	})

	t.Run("DeleteNestedDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Create nested directory structure
		nestedDir := filepath.Join(tempDir, "level1", "level2", "level3")
		err := os.MkdirAll(nestedDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create nested directories: %v", err)
		}

		// Create files at different levels
		file1 := filepath.Join(tempDir, "level1", "file1.txt")
		file2 := filepath.Join(tempDir, "level1", "level2", "file2.txt")
		file3 := filepath.Join(nestedDir, "file3.txt")

		for _, file := range []string{file1, file2, file3} {
			err = os.WriteFile(file, []byte("content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", file, err)
			}
		}

		rootTestDir := filepath.Join(tempDir, "level1")
		err = njsFile.DeleteFile(rootTestDir)
		if err != nil {
			t.Fatalf("Expected no error deleting nested directory, got: %v", err)
		}

		// Verify entire structure is deleted
		if _, err := os.Stat(rootTestDir); !os.IsNotExist(err) {
			t.Error("Nested directory structure should not exist after deletion")
		}
	})
}

func TestNJSFile_WriteFile(t *testing.T) {
	njsFile := &NJSFile{}

	t.Run("WriteNewFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "new_file.txt")
		content := "Hello, World!\nThis is a test file."

		err := njsFile.WriteFile(testFile, content)
		if err != nil {
			t.Fatalf("Expected no error writing file, got: %v", err)
		}

		// Verify file was created and has correct content
		readContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}

		if string(readContent) != content {
			t.Errorf("Expected content %q, got %q", content, string(readContent))
		}
	})

	t.Run("OverwriteExistingFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "existing_file.txt")
		originalContent := "Original content"
		newContent := "New content that overwrites the original"

		// Create original file
		err := os.WriteFile(testFile, []byte(originalContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create original file: %v", err)
		}

		// Overwrite with new content
		err = njsFile.WriteFile(testFile, newContent)
		if err != nil {
			t.Fatalf("Expected no error overwriting file, got: %v", err)
		}

		// Verify file has new content
		readContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read overwritten file: %v", err)
		}

		if string(readContent) != newContent {
			t.Errorf("Expected content %q, got %q", newContent, string(readContent))
		}
	})

	t.Run("WriteEmptyFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "empty_file.txt")
		content := ""

		err := njsFile.WriteFile(testFile, content)
		if err != nil {
			t.Fatalf("Expected no error writing empty file, got: %v", err)
		}

		// Verify file was created and is empty
		readContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read empty file: %v", err)
		}

		if len(readContent) != 0 {
			t.Errorf("Expected empty file, got %d bytes", len(readContent))
		}
	})

	t.Run("WriteFileWithSpecialCharacters", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "special_chars.txt")
		content := "Special chars: áéíóú ñ 中文 🚀 \n\t\r"

		err := njsFile.WriteFile(testFile, content)
		if err != nil {
			t.Fatalf("Expected no error writing file with special chars, got: %v", err)
		}

		// Verify file has correct content
		readContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read file with special chars: %v", err)
		}

		if string(readContent) != content {
			t.Errorf("Expected content %q, got %q", content, string(readContent))
		}
	})

	t.Run("WriteToNonExistentDirectory", func(t *testing.T) {
		testFile := "/path/that/does/not/exist/file.txt"
		content := "This should fail"

		err := njsFile.WriteFile(testFile, content)
		if err == nil {
			t.Error("Expected error writing to non-existent directory")
		}
	})

	t.Run("WriteLargeFile", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "large_file.txt")
		
		// Create large content (1MB)
		var content strings.Builder
		line := strings.Repeat("This is a long line with lots of content. ", 100) + "\n"
		for i := 0; i < 1000; i++ {
			content.WriteString(line)
		}

		err := njsFile.WriteFile(testFile, content.String())
		if err != nil {
			t.Fatalf("Expected no error writing large file, got: %v", err)
		}

		// Verify file size
		fileInfo, err := os.Stat(testFile)
		if err != nil {
			t.Fatalf("Failed to stat large file: %v", err)
		}

		expectedSize := int64(len(content.String()))
		if fileInfo.Size() != expectedSize {
			t.Errorf("Expected file size %d, got %d", expectedSize, fileInfo.Size())
		}

		// Verify content (just check beginning and end)
		readContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read large file: %v", err)
		}

		if string(readContent) != content.String() {
			t.Error("Large file content mismatch")
		}
	})

	t.Run("WriteFilePermissions", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "perm_test.txt")
		content := "Testing file permissions"

		err := njsFile.WriteFile(testFile, content)
		if err != nil {
			t.Fatalf("Expected no error writing file, got: %v", err)
		}

		// Check file permissions (should be fs.ModePerm which is 0777)
		fileInfo, err := os.Stat(testFile)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}

		// Note: actual permissions may be modified by umask
		mode := fileInfo.Mode()
		if mode&0200 == 0 { // Check if owner write bit is set
			t.Error("Expected file to be writable by owner")
		}
		if mode&0400 == 0 { // Check if owner read bit is set
			t.Error("Expected file to be readable by owner")
		}
	})
}
