package script

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type NJSFile struct {
}

// ReadFileAsLines reads a file and returns its content as a slice of strings (lines).
// It provides enhanced error handling, input validation, and performance optimizations.
//
// Parameters:
//   - filePath: path to the file to read
//
// Returns:
//   - []string: slice of lines from the file (without line endings)
//   - error: detailed error information if operation fails
//
// Features:
//   - Input validation (empty path, file existence)
//   - Memory-efficient reading with pre-allocated capacity estimation
//   - Enhanced error messages with context
//   - Handles large files efficiently
//   - Preserves empty lines
func (js *NJSFile) ReadFileAsLines(filePath string) ([]string, error) {
	// Input validation
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// Clean and validate the file path
	cleanPath := filepath.Clean(filePath)

	// Check if file exists and get file info for optimization
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %s", cleanPath)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied accessing file: %s", cleanPath)
		}
		return nil, fmt.Errorf("error accessing file %s: %w", cleanPath, err)
	}

	// Ensure it's actually a file, not a directory
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", cleanPath)
	}

	// Check for empty file
	if fileInfo.Size() == 0 {
		return []string{}, nil
	}

	// Open the file
	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", cleanPath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log the close error, but don't override the main error
			fmt.Fprintf(os.Stderr, "Warning: failed to close file %s: %v\n", cleanPath, closeErr)
		}
	}()

	// Estimate initial capacity based on file size
	// Assume average line length of 80 characters for better memory allocation
	estimatedLines := int(fileInfo.Size()/80) + 1
	if estimatedLines > 10000 {
		// Cap initial allocation for very large files
		estimatedLines = 10000
	}

	lines := make([]string, 0, estimatedLines)
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large files to improve performance
	if fileInfo.Size() > 1024*1024 { // 1MB
		buf := make([]byte, 0, 64*1024) // 64KB buffer
		scanner.Buffer(buf, 1024*1024)  // 1MB max token size
	}

	lineCount := 0
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		lineCount++

		// Safety check for extremely large files
		if lineCount > 1000000 { // 1 million lines
			return nil, fmt.Errorf("file too large: exceeds maximum line limit of 1,000,000 lines")
		}
	}

	// Check for scanner errors
	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s at line %d: %w", cleanPath, lineCount+1, err)
	}

	return lines, nil
}

// ReadFileAsString reads a file and returns its entire content as a string.
// It provides enhanced error handling, input validation, and performance optimizations.
//
// Parameters:
//   - filePath: path to the file to read
//
// Returns:
//   - *string: pointer to string containing the entire file content
//   - error: detailed error information if operation fails
//
// Features:
//   - Input validation (empty path, file existence)
//   - Memory-efficient reading with size-based optimization
//   - Enhanced error messages with context
//   - Handles large files efficiently with size limits
//   - Preserves original file encoding
func (js *NJSFile) ReadFileAsString(filePath string) (*string, error) {
	// Input validation
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// Clean and validate the file path
	cleanPath := filepath.Clean(filePath)
	
	// Check if file exists and get file info for optimization
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %s", cleanPath)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied accessing file: %s", cleanPath)
		}
		return nil, fmt.Errorf("error accessing file %s: %w", cleanPath, err)
	}

	// Ensure it's actually a file, not a directory
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", cleanPath)
	}

	// Check file size limits for safety
	const maxFileSize = 100 * 1024 * 1024 // 100MB limit
	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes exceeds maximum size of %d bytes", fileInfo.Size(), maxFileSize)
	}

	// Handle empty file
	if fileInfo.Size() == 0 {
		emptyString := ""
		return &emptyString, nil
	}

	// Open the file
	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", cleanPath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log the close error, but don't override the main error
			fmt.Fprintf(os.Stderr, "Warning: failed to close file %s: %v\n", cleanPath, closeErr)
		}
	}()

	// For large files, use a more efficient reading approach
	var bytes []byte
	if fileInfo.Size() > 1024*1024 { // 1MB
		// Pre-allocate buffer for large files
		bytes = make([]byte, 0, fileInfo.Size())
		
		// Read in chunks for better memory management
		buffer := make([]byte, 64*1024) // 64KB chunks
		for {
			n, err := file.Read(buffer)
			if n > 0 {
				bytes = append(bytes, buffer[:n]...)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("error reading file %s: %w", cleanPath, err)
			}
		}
	} else {
		// For smaller files, use io.ReadAll for simplicity
		bytes, err = io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %w", cleanPath, err)
		}
	}

	// Convert to string
	text := string(bytes)
	return &text, nil
}

func (js *NJSFile) IsFile(filePath string) (bool, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	return !fi.IsDir(), err
}

func (js *NJSFile) IsDir(filePath string) (bool, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), err
}

func (js *NJSFile) DeleteFile(filePath string) error {
	return os.RemoveAll(filePath)
}

func (js *NJSFile) WriteFile(filePath string, content string) error {
	return os.WriteFile(filePath, []byte(content), fs.ModePerm)
}
