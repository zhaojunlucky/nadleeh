package util

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCalculateFileSHA256(t *testing.T) {
	t.Run("EmptyFile", func(t *testing.T) {
		// Create a temporary empty file
		tempDir := t.TempDir()
		emptyFile := filepath.Join(tempDir, "empty.txt")
		
		file, err := os.Create(emptyFile)
		if err != nil {
			t.Fatalf("Failed to create empty test file: %v", err)
		}
		file.Close()
		
		hash, err := CalculateFileSHA256(emptyFile)
		if err != nil {
			t.Errorf("Expected no error for empty file, got: %v", err)
		}
		
		// SHA-256 of empty file is known
		expectedHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})
	
	t.Run("SimpleTextFile", func(t *testing.T) {
		// Create a temporary file with known content
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		content := "Hello, World!"
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		hash, err := CalculateFileSHA256(testFile)
		if err != nil {
			t.Errorf("Expected no error for simple text file, got: %v", err)
		}
		
		// Calculate expected hash manually
		expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})
	
	t.Run("BinaryFile", func(t *testing.T) {
		// Create a temporary binary file
		tempDir := t.TempDir()
		binaryFile := filepath.Join(tempDir, "binary.bin")
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
		
		err := os.WriteFile(binaryFile, binaryData, 0644)
		if err != nil {
			t.Fatalf("Failed to create binary test file: %v", err)
		}
		
		hash, err := CalculateFileSHA256(binaryFile)
		if err != nil {
			t.Errorf("Expected no error for binary file, got: %v", err)
		}
		
		// Calculate expected hash manually
		expectedHash := fmt.Sprintf("%x", sha256.Sum256(binaryData))
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})
	
	t.Run("LargeFile", func(t *testing.T) {
		// Create a large temporary file (1MB)
		tempDir := t.TempDir()
		largeFile := filepath.Join(tempDir, "large.txt")
		
		// Create content that's 1MB
		content := strings.Repeat("A", 1024*1024)
		err := os.WriteFile(largeFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create large test file: %v", err)
		}
		
		hash, err := CalculateFileSHA256(largeFile)
		if err != nil {
			t.Errorf("Expected no error for large file, got: %v", err)
		}
		
		// Verify hash is not empty and has correct length (64 hex characters)
		if len(hash) != 64 {
			t.Errorf("Expected hash length 64, got %d", len(hash))
		}
		
		// Verify hash contains only hex characters
		for _, char := range hash {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				t.Errorf("Hash contains non-hex character: %c", char)
				break
			}
		}
	})
	
	t.Run("FileWithNewlines", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "newlines.txt")
		content := "Line 1\nLine 2\nLine 3\n"
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file with newlines: %v", err)
		}
		
		hash, err := CalculateFileSHA256(testFile)
		if err != nil {
			t.Errorf("Expected no error for file with newlines, got: %v", err)
		}
		
		// Calculate expected hash manually
		expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})
	
	t.Run("FileWithUnicodeContent", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "unicode.txt")
		content := "Hello 世界 🌍 Привет мир"
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create unicode test file: %v", err)
		}
		
		hash, err := CalculateFileSHA256(testFile)
		if err != nil {
			t.Errorf("Expected no error for unicode file, got: %v", err)
		}
		
		// Calculate expected hash manually
		expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})
	
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentFile := "/path/that/does/not/exist/file.txt"
		
		hash, err := CalculateFileSHA256(nonExistentFile)
		
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
		
		if hash != "" {
			t.Errorf("Expected empty hash for error case, got: %s", hash)
		}
		
		// Check error message contains expected text
		if !strings.Contains(err.Error(), "failed to open file") {
			t.Errorf("Expected error message to contain 'failed to open file', got: %s", err.Error())
		}
	})
	
	t.Run("DirectoryInsteadOfFile", func(t *testing.T) {
		tempDir := t.TempDir()
		
		hash, err := CalculateFileSHA256(tempDir)
		
		if err == nil {
			t.Error("Expected error when trying to hash a directory, got nil")
		}
		
		if hash != "" {
			t.Errorf("Expected empty hash for error case, got: %s", hash)
		}
	})
	
	t.Run("FileWithNoReadPermissions", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "noperm.txt")
		
		// Create file with content
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		// Remove read permissions
		err = os.Chmod(testFile, 0000)
		if err != nil {
			t.Fatalf("Failed to change file permissions: %v", err)
		}
		
		// Restore permissions after test for cleanup
		defer os.Chmod(testFile, 0644)
		
		hash, err := CalculateFileSHA256(testFile)
		
		if err == nil {
			t.Error("Expected error for file without read permissions, got nil")
		}
		
		if hash != "" {
			t.Errorf("Expected empty hash for error case, got: %s", hash)
		}
	})
	
	t.Run("SameContentSameHash", func(t *testing.T) {
		tempDir := t.TempDir()
		content := "Identical content for hash comparison"
		
		// Create two files with identical content
		file1 := filepath.Join(tempDir, "file1.txt")
		file2 := filepath.Join(tempDir, "file2.txt")
		
		err := os.WriteFile(file1, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create first test file: %v", err)
		}
		
		err = os.WriteFile(file2, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create second test file: %v", err)
		}
		
		hash1, err1 := CalculateFileSHA256(file1)
		if err1 != nil {
			t.Errorf("Error calculating hash for file1: %v", err1)
		}
		
		hash2, err2 := CalculateFileSHA256(file2)
		if err2 != nil {
			t.Errorf("Error calculating hash for file2: %v", err2)
		}
		
		if hash1 != hash2 {
			t.Errorf("Expected identical hashes for identical content. Hash1: %s, Hash2: %s", hash1, hash2)
		}
	})
	
	t.Run("DifferentContentDifferentHash", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Create two files with different content
		file1 := filepath.Join(tempDir, "file1.txt")
		file2 := filepath.Join(tempDir, "file2.txt")
		
		err := os.WriteFile(file1, []byte("Content A"), 0644)
		if err != nil {
			t.Fatalf("Failed to create first test file: %v", err)
		}
		
		err = os.WriteFile(file2, []byte("Content B"), 0644)
		if err != nil {
			t.Fatalf("Failed to create second test file: %v", err)
		}
		
		hash1, err1 := CalculateFileSHA256(file1)
		if err1 != nil {
			t.Errorf("Error calculating hash for file1: %v", err1)
		}
		
		hash2, err2 := CalculateFileSHA256(file2)
		if err2 != nil {
			t.Errorf("Error calculating hash for file2: %v", err2)
		}
		
		if hash1 == hash2 {
			t.Errorf("Expected different hashes for different content. Both hashes: %s", hash1)
		}
	})
}

func TestCalculateFileSHA256_EdgeCases(t *testing.T) {
	t.Run("EmptyFilePath", func(t *testing.T) {
		hash, err := CalculateFileSHA256("")
		
		if err == nil {
			t.Error("Expected error for empty file path, got nil")
		}
		
		if hash != "" {
			t.Errorf("Expected empty hash for error case, got: %s", hash)
		}
	})
	
	t.Run("FileWithOnlyWhitespace", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "whitespace.txt")
		content := "   \t\n\r   "
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create whitespace test file: %v", err)
		}
		
		hash, err := CalculateFileSHA256(testFile)
		if err != nil {
			t.Errorf("Expected no error for whitespace file, got: %v", err)
		}
		
		// Calculate expected hash manually
		expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})
	
	t.Run("FileWithNullBytes", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "nullbytes.bin")
		content := []byte{'H', 'e', 'l', 'l', 'o', 0, 'W', 'o', 'r', 'l', 'd', 0}
		
		err := os.WriteFile(testFile, content, 0644)
		if err != nil {
			t.Fatalf("Failed to create null bytes test file: %v", err)
		}
		
		hash, err := CalculateFileSHA256(testFile)
		if err != nil {
			t.Errorf("Expected no error for file with null bytes, got: %v", err)
		}
		
		// Calculate expected hash manually
		expectedHash := fmt.Sprintf("%x", sha256.Sum256(content))
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})
}

func TestCalculateFileSHA256_Consistency(t *testing.T) {
	t.Run("MultipleCalculationsConsistent", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "consistent.txt")
		content := "Content for consistency testing"
		
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		// Calculate hash multiple times
		var hashes []string
		for i := 0; i < 5; i++ {
			hash, err := CalculateFileSHA256(testFile)
			if err != nil {
				t.Errorf("Error in calculation %d: %v", i+1, err)
			}
			hashes = append(hashes, hash)
		}
		
		// All hashes should be identical
		firstHash := hashes[0]
		for i, hash := range hashes {
			if hash != firstHash {
				t.Errorf("Hash %d differs from first hash. Expected: %s, Got: %s", i+1, firstHash, hash)
			}
		}
	})
}

// Benchmark tests
func BenchmarkCalculateFileSHA256(b *testing.B) {
	// Create test files of different sizes
	tempDir := b.TempDir()
	
	b.Run("SmallFile", func(b *testing.B) {
		smallFile := filepath.Join(tempDir, "small.txt")
		content := strings.Repeat("A", 1024) // 1KB
		os.WriteFile(smallFile, []byte(content), 0644)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CalculateFileSHA256(smallFile)
		}
	})
	
	b.Run("MediumFile", func(b *testing.B) {
		mediumFile := filepath.Join(tempDir, "medium.txt")
		content := strings.Repeat("B", 1024*100) // 100KB
		os.WriteFile(mediumFile, []byte(content), 0644)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CalculateFileSHA256(mediumFile)
		}
	})
	
	b.Run("LargeFile", func(b *testing.B) {
		largeFile := filepath.Join(tempDir, "large.txt")
		content := strings.Repeat("C", 1024*1024) // 1MB
		os.WriteFile(largeFile, []byte(content), 0644)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CalculateFileSHA256(largeFile)
		}
	})
	
	b.Run("EmptyFile", func(b *testing.B) {
		emptyFile := filepath.Join(tempDir, "empty.txt")
		os.WriteFile(emptyFile, []byte(""), 0644)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CalculateFileSHA256(emptyFile)
		}
	})
}
