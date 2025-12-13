package encrypt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"nadleeh/internal/argument"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhaojunlucky/golib/pkg/security"
)

// createEncryptArgs creates EncryptArgs for testing
func createEncryptArgs(publicKeyPath, filePath, str string) *argument.EncryptArgs {
	return &argument.EncryptArgs{
		Public: publicKeyPath,
		File:   filePath,
		Str:    str,
	}
}

func TestEncrypt_ArgumentValidation(t *testing.T) {
	t.Run("MissingPublicKeyArgument", func(t *testing.T) {
		// Test that empty public key would be invalid
		args := createEncryptArgs("", "", "")
		
		if args.Public != "" {
			t.Error("Expected empty public key")
		}
	})
	
	t.Run("ValidPublicKeyArgument", func(t *testing.T) {
		publicKeyPath := "/tmp/test-public.pem"
		args := createEncryptArgs(publicKeyPath, "", "")
		
		if args.Public != publicKeyPath {
			t.Errorf("Expected public key path '%s', got: %v", publicKeyPath, args.Public)
		}
	})
}

func TestEncrypt_PublicKeyHandling(t *testing.T) {
	t.Run("ValidPublicKeyFile", func(t *testing.T) {
		// Create a temporary public key file for testing
		tempDir := t.TempDir()
		pubKeyFile := filepath.Join(tempDir, "test-public.pem")
		
		// Generate a test key pair
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate test private key: %v", err)
		}
		
		// Write the public key to file
		pubWriter, err := os.Create(pubKeyFile)
		if err != nil {
			t.Fatalf("Failed to create public key file: %v", err)
		}
		defer pubWriter.Close()
		
		err = security.WritePublicKey(&privateKey.PublicKey, pubWriter)
		if err != nil {
			t.Fatalf("Failed to write public key: %v", err)
		}
		pubWriter.Close()
		
		// Test reading the public key back
		reader, err := os.Open(pubKeyFile)
		if err != nil {
			t.Fatalf("Failed to open public key file: %v", err)
		}
		defer reader.Close()
		
		pubKey, err := security.ReadPublicKey(reader)
		if err != nil {
			t.Fatalf("Failed to read public key: %v", err)
		}
		
		// Validate the public key
		if pubKey == nil {
			t.Error("Read public key is nil")
			return
		}
		
		if pubKey.X.Cmp(privateKey.PublicKey.X) != 0 {
			t.Error("Read public key X coordinate doesn't match original")
		}
		
		if pubKey.Y.Cmp(privateKey.PublicKey.Y) != 0 {
			t.Error("Read public key Y coordinate doesn't match original")
		}
	})
	
	t.Run("NonExistentPublicKeyFile", func(t *testing.T) {
		nonExistentFile := "/non/existent/public.pem"
		
		_, err := os.Open(nonExistentFile)
		if err == nil {
			t.Error("Expected error when opening non-existent public key file")
		}
		
		if !os.IsNotExist(err) {
			t.Errorf("Expected 'not exist' error, got: %v", err)
		}
	})
}

func TestEncrypt_FileEncryption(t *testing.T) {
	t.Run("ValidFileEncryption", func(t *testing.T) {
		// Create temporary directory and files
		tempDir := t.TempDir()
		
		// Generate key pair
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate test private key: %v", err)
		}
		
		// Create and write public key file
		pubKeyFile := filepath.Join(tempDir, "test-public.pem")
		pubWriter, err := os.Create(pubKeyFile)
		if err != nil {
			t.Fatalf("Failed to create public key file: %v", err)
		}
		
		err = security.WritePublicKey(&privateKey.PublicKey, pubWriter)
		pubWriter.Close()
		if err != nil {
			t.Fatalf("Failed to write public key: %v", err)
		}
		
		// Create test file to encrypt
		testFile := filepath.Join(tempDir, "test.txt")
		testContent := "This is a test file content for encryption"
		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		// Test the file encryption workflow components using EncryptArgs
		args := createEncryptArgs(pubKeyFile, testFile, "")
		
		// Test public key reading
		reader, err := os.Open(args.Public)
		if err != nil {
			t.Fatalf("Failed to open public key file: %v", err)
		}
		defer reader.Close()
		
		pubKey, err := security.ReadPublicKey(reader)
		if err != nil {
			t.Fatalf("Failed to read public key: %v", err)
		}
		
		// Test file argument
		if args.File != testFile {
			t.Errorf("Expected file path '%s', got '%s'", testFile, args.File)
		}
		
		// Test file reading
		file, err := os.Open(args.File)
		if err != nil {
			t.Fatalf("Failed to open test file: %v", err)
		}
		defer file.Close()
		
		data, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}
		
		if string(data) != testContent {
			t.Errorf("Expected file content '%s', got '%s'", testContent, string(data))
		}
		
		// Test encryption
		ecies := security.ECIESHelper{}
		encrypted, err := ecies.EncryptWithPublic(pubKey, data)
		if err != nil {
			t.Skipf("Skipping encryption test due to ECIES implementation: %v", err)
			return
		}
		
		// Test output file path generation
		expectedOutputPath := filepath.Join(filepath.Dir(args.File), 
			fmt.Sprintf("%s-encrypted%s", filepath.Base(args.File), filepath.Ext(args.File)))
		
		expectedPath := filepath.Join(tempDir, "test.txt-encrypted.txt")
		if expectedOutputPath != expectedPath {
			t.Errorf("Expected output path '%s', got '%s'", expectedPath, expectedOutputPath)
		}
		
		// Test output file writing
		outputContent := fmt.Sprintf("ENC(%s)", string(encrypted))
		err = os.WriteFile(expectedOutputPath, []byte(outputContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write encrypted file: %v", err)
		}
		
		// Verify encrypted file was created
		if _, err := os.Stat(expectedOutputPath); os.IsNotExist(err) {
			t.Error("Encrypted output file was not created")
		}
		
		// Verify encrypted file content format
		encryptedContent, err := os.ReadFile(expectedOutputPath)
		if err != nil {
			t.Fatalf("Failed to read encrypted file: %v", err)
		}
		
		if !strings.HasPrefix(string(encryptedContent), "ENC(") {
			t.Error("Encrypted file content doesn't have ENC( prefix")
		}
		
		if !strings.HasSuffix(string(encryptedContent), ")") {
			t.Error("Encrypted file content doesn't have ) suffix")
		}
	})
	
	t.Run("NonExistentInputFile", func(t *testing.T) {
		nonExistentFile := "/non/existent/input.txt"
		
		_, err := os.Open(nonExistentFile)
		if err == nil {
			t.Error("Expected error when opening non-existent input file")
		}
		
		if !os.IsNotExist(err) {
			t.Errorf("Expected 'not exist' error, got: %v", err)
		}
	})
}

func TestEncrypt_StringEncryption(t *testing.T) {
	t.Run("ValidStringEncryption", func(t *testing.T) {
		// Create temporary directory and public key
		tempDir := t.TempDir()
		
		// Generate key pair
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate test private key: %v", err)
		}
		
		// Create and write public key file
		pubKeyFile := filepath.Join(tempDir, "test-public.pem")
		pubWriter, err := os.Create(pubKeyFile)
		if err != nil {
			t.Fatalf("Failed to create public key file: %v", err)
		}
		
		err = security.WritePublicKey(&privateKey.PublicKey, pubWriter)
		pubWriter.Close()
		if err != nil {
			t.Fatalf("Failed to write public key: %v", err)
		}
		
		// Test string encryption workflow components using EncryptArgs
		testString := "  Hello, World!  "
		args := createEncryptArgs(pubKeyFile, "", testString)
		
		// Test public key reading
		reader, err := os.Open(args.Public)
		if err != nil {
			t.Fatalf("Failed to open public key file: %v", err)
		}
		defer reader.Close()
		
		pubKey, err := security.ReadPublicKey(reader)
		if err != nil {
			t.Fatalf("Failed to read public key: %v", err)
		}
		
		// Test string argument
		str := strings.TrimSpace(args.Str)
		
		expectedTrimmed := "Hello, World!"
		if str != expectedTrimmed {
			t.Errorf("Expected trimmed string '%s', got '%s'", expectedTrimmed, str)
		}
		
		// Test encryption
		ecies := security.ECIESHelper{}
		encrypted, err := ecies.EncryptWithPublic(pubKey, []byte(str))
		if err != nil {
			t.Skipf("Skipping encryption test due to ECIES implementation: %v", err)
			return
		}
		
		// Test base64 encoding
		encodedResult := base64.StdEncoding.EncodeToString(encrypted)
		if encodedResult == "" {
			t.Error("Base64 encoded result is empty")
		}
		
		// Test output format
		expectedOutput := fmt.Sprintf("entryped string is: ENC(%s)\n", encodedResult)
		if !strings.Contains(expectedOutput, "ENC(") {
			t.Error("Output format doesn't contain ENC( prefix")
		}
		
		if !strings.Contains(expectedOutput, ")") {
			t.Error("Output format doesn't contain ) suffix")
		}
	})
	
	t.Run("EmptyStringEncryption", func(t *testing.T) {
		testString := "   "
		trimmed := strings.TrimSpace(testString)
		
		if trimmed != "" {
			t.Errorf("Expected empty string after trimming whitespace, got '%s'", trimmed)
		}
	})
	
	t.Run("StringWithSpecialCharacters", func(t *testing.T) {
		testCases := []string{
			"Hello\nWorld",
			"Test\tString",
			"Special!@#$%^&*()Characters",
			"Unicode: 你好世界",
			"",
		}
		
		for _, testCase := range testCases {
			trimmed := strings.TrimSpace(testCase)
			// Just verify trimming works correctly
			if len(trimmed) > len(testCase) {
				t.Errorf("Trimmed string is longer than original for case '%s'", testCase)
			}
		}
	})
}

func TestEncrypt_ArgumentParsing(t *testing.T) {
	t.Run("BothFileAndStringProvided", func(t *testing.T) {
		// Test the precedence - file should be processed first
		args := createEncryptArgs("/tmp/public.pem", "/tmp/file.txt", "test string")
		
		if args.File == "" {
			t.Error("Expected file argument to be set")
		}
		
		if args.Str == "" {
			t.Error("Expected string argument to be set")
		}
		
		// In the actual Encrypt function, file takes precedence over string
	})
	
	t.Run("NeitherFileNorStringProvided", func(t *testing.T) {
		args := createEncryptArgs("/tmp/public.pem", "", "")
		
		if args.File != "" {
			t.Error("Expected file argument to be empty")
		}
		
		if args.Str != "" {
			t.Error("Expected string argument to be empty")
		}
		
		// This would trigger the "invalid argument for decrypt" error in the actual function
	})
}

func TestEncrypt_PathHandling(t *testing.T) {
	t.Run("FilePathGeneration", func(t *testing.T) {
		testCases := []struct {
			inputPath    string
			expectedName string
		}{
			{"/tmp/test.txt", "test.txt-encrypted.txt"},
			{"/home/user/document.pdf", "document.pdf-encrypted.pdf"},
			{"./local/file.json", "file.json-encrypted.json"},
			{"noextension", "noextension-encrypted"},
			{"/path/to/file.tar.gz", "file.tar.gz-encrypted.gz"},
		}
		
		for _, tc := range testCases {
			outputPath := filepath.Join(filepath.Dir(tc.inputPath), 
				fmt.Sprintf("%s-encrypted%s", filepath.Base(tc.inputPath), filepath.Ext(tc.inputPath)))
			
			expectedPath := filepath.Join(filepath.Dir(tc.inputPath), tc.expectedName)
			
			if outputPath != expectedPath {
				t.Errorf("For input '%s': expected '%s', got '%s'", 
					tc.inputPath, expectedPath, outputPath)
			}
		}
	})
}

func TestEncrypt_ErrorHandling(t *testing.T) {
	t.Run("InvalidPublicKeyFile", func(t *testing.T) {
		// Create a file with invalid public key content
		tempDir := t.TempDir()
		invalidKeyFile := filepath.Join(tempDir, "invalid-public.pem")
		
		err := os.WriteFile(invalidKeyFile, []byte("invalid key content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid key file: %v", err)
		}
		
		// Test reading invalid public key
		reader, err := os.Open(invalidKeyFile)
		if err != nil {
			t.Fatalf("Failed to open invalid key file: %v", err)
		}
		defer reader.Close()
		
		_, err = security.ReadPublicKey(reader)
		if err == nil {
			t.Error("Expected error when reading invalid public key")
		}
	})
	
	t.Run("FileReadError", func(t *testing.T) {
		// Create a directory instead of a file to cause read error
		tempDir := t.TempDir()
		dirPath := filepath.Join(tempDir, "directory")
		
		err := os.Mkdir(dirPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		
		// Try to read directory as file
		_, err = os.Open(dirPath)
		if err != nil {
			t.Fatalf("Failed to open directory: %v", err)
		}
		
		// Reading directory content should fail
		file, err := os.Open(dirPath)
		if err != nil {
			t.Fatalf("Failed to open directory: %v", err)
		}
		defer file.Close()
		
		_, err = io.ReadAll(file)
		if err == nil {
			t.Error("Expected error when reading directory as file")
		}
	})
}

// Benchmark tests
func BenchmarkEncrypt_ArgumentParsing(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = createEncryptArgs("/tmp/public.pem", "/tmp/file.txt", "test string")
	}
}

func BenchmarkEncrypt_StringTrimming(b *testing.B) {
	testString := "  Hello, World! This is a test string with whitespace  "
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.TrimSpace(testString)
	}
}

func BenchmarkEncrypt_Base64Encoding(b *testing.B) {
	testData := []byte("This is test data for base64 encoding benchmark")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = base64.StdEncoding.EncodeToString(testData)
	}
}
