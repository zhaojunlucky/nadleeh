package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func CalculateFileSHA256(filePath string) (string, error) {
	// 1. Open the file.
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 2. Create a new SHA-256 hash object.
	hash := sha256.New()

	// 3. Copy the file's content to the hash object.
	// io.Copy handles reading from the file and writing to the hash object efficiently.
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to copy file content to hash: %w", err)
	}

	// 4. Get the final hash sum and return it as a hexadecimal string.
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
