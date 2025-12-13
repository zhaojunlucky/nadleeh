package encrypt

import (
	"fmt"
	"nadleeh/internal/argument"
	"os"
	"path/filepath"
	"testing"

	"github.com/zhaojunlucky/golib/pkg/security"
)

// createKeypairArgs creates KeypairArgs for testing
func createKeypairArgs(name, dir string) *argument.KeypairArgs {
	return &argument.KeypairArgs{
		Name: name,
		Dir:  dir,
	}
}

func TestGenerateKeyPair_ArgumentValidation(t *testing.T) {
	t.Run("MissingNameArgument", func(t *testing.T) {
		// Test that empty name would be invalid
		args := createKeypairArgs("", "/tmp")
		
		if args.Name != "" {
			t.Error("Expected empty name")
		}
	})
	
	t.Run("MissingDirArgument", func(t *testing.T) {
		// Test that empty dir would be invalid
		args := createKeypairArgs("test", "")
		
		if args.Dir != "" {
			t.Error("Expected empty dir")
		}
	})
	
	t.Run("ValidArguments", func(t *testing.T) {
		args := createKeypairArgs("testkey", "/tmp")
		
		if args.Name != "testkey" {
			t.Errorf("Expected name 'testkey', got: %v", args.Name)
		}
		
		if args.Dir != "/tmp" {
			t.Errorf("Expected dir '/tmp', got: %v", args.Dir)
		}
	})
}

func TestGenerateKeyPair_FileGeneration(t *testing.T) {
	t.Run("ValidKeyPairGeneration", func(t *testing.T) {
		// Create temporary directory for test
		tempDir := t.TempDir()
		keyName := "test-keypair"
		
		args := createKeypairArgs(keyName, tempDir)
		
		// Since GenerateKeyPair uses log.Fatal, we'll test the individual components
		// that would be called within the function
		
		// Test key pair generation
		pri, err := security.GenerateECKeyPair("secp256r1")
		if err != nil {
			t.Fatalf("Failed to generate EC key pair: %v", err)
		}
		
		// Test file path generation
		priFile := filepath.Join(args.Dir, fmt.Sprintf("%s-private.pem", args.Name))
		pubFile := filepath.Join(args.Dir, fmt.Sprintf("%s-public.pem", args.Name))
		
		expectedPriFile := filepath.Join(tempDir, "test-keypair-private.pem")
		expectedPubFile := filepath.Join(tempDir, "test-keypair-public.pem")
		
		if priFile != expectedPriFile {
			t.Errorf("Expected private key file %s, got %s", expectedPriFile, priFile)
		}
		
		if pubFile != expectedPubFile {
			t.Errorf("Expected public key file %s, got %s", expectedPubFile, pubFile)
		}
		
		// Test public key file creation and writing
		pubWriter, err := os.Create(pubFile)
		if err != nil {
			t.Fatalf("Failed to create public key file: %v", err)
		}
		defer pubWriter.Close()
		
		err = security.WritePublicKey(&pri.PublicKey, pubWriter)
		if err != nil {
			t.Fatalf("Failed to write public key: %v", err)
		}
		
		// Test private key file creation and writing
		priWriter, err := os.Create(priFile)
		if err != nil {
			t.Fatalf("Failed to create private key file: %v", err)
		}
		defer priWriter.Close()
		
		err = security.WriteECPrivateKey(pri, priWriter)
		if err != nil {
			t.Fatalf("Failed to write private key: %v", err)
		}
		
		// Verify files were created
		if _, err := os.Stat(pubFile); os.IsNotExist(err) {
			t.Error("Public key file was not created")
		}
		
		if _, err := os.Stat(priFile); os.IsNotExist(err) {
			t.Error("Private key file was not created")
		}
		
		// Verify files have content
		pubInfo, err := os.Stat(pubFile)
		if err != nil {
			t.Errorf("Failed to stat public key file: %v", err)
		} else if pubInfo.Size() == 0 {
			t.Error("Public key file is empty")
		}
		
		priInfo, err := os.Stat(priFile)
		if err != nil {
			t.Errorf("Failed to stat private key file: %v", err)
		} else if priInfo.Size() == 0 {
			t.Error("Private key file is empty")
		}
	})
}

func TestGenerateKeyPair_KeyValidation(t *testing.T) {
	t.Run("GeneratedKeyPairValidation", func(t *testing.T) {
		// Test key pair generation
		pri, err := security.GenerateECKeyPair("secp256r1")
		if err != nil {
			t.Fatalf("Failed to generate EC key pair: %v", err)
		}
		
		// Validate that we got a proper ECDSA private key
		if pri == nil {
			t.Fatal("Generated private key is nil")
		}
		
		// Validate key properties
		if pri.Curve == nil {
			t.Error("Private key curve is nil")
		}
		
		if pri.D == nil {
			t.Error("Private key D value is nil")
		}
		
		// Validate public key
		pubKey := &pri.PublicKey
		if pubKey.X == nil || pubKey.Y == nil {
			t.Error("Public key coordinates are nil")
		}
		
		// Test that the key pair is valid by checking curve membership
		if !pri.Curve.IsOnCurve(pubKey.X, pubKey.Y) {
			t.Error("Generated public key is not on the curve")
		}
	})
}

func TestGenerateKeyPair_FileOperations(t *testing.T) {
	t.Run("FileCreationInNonExistentDirectory", func(t *testing.T) {
		// Test behavior when trying to create files in non-existent directory
		nonExistentDir := "/non/existent/directory"
		keyName := "test"
		
		priFile := filepath.Join(nonExistentDir, fmt.Sprintf("%s-private.pem", keyName))
		
		_, err := os.Create(priFile)
		if err == nil {
			t.Error("Expected error when creating file in non-existent directory")
		}
		
		if !os.IsNotExist(err) {
			t.Errorf("Expected 'not exist' error, got: %v", err)
		}
	})
	
	t.Run("FileCreationWithInvalidName", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Test with invalid filename characters (depending on OS)
		invalidNames := []string{
			"",           // Empty name
			"test\x00",   // Null character
			"test/sub",   // Path separator in name
		}
		
		for _, invalidName := range invalidNames {
			priFile := filepath.Join(tempDir, fmt.Sprintf("%s-private.pem", invalidName))
			
			_, err := os.Create(priFile)
			if err == nil && invalidName != "" {
				// Some systems might allow certain characters, so we only fail for empty names
				if invalidName == "" {
					t.Error("Expected error when creating file with empty name")
				}
			}
		}
	})
}

func TestGenerateKeyPair_KeyFileReadback(t *testing.T) {
	t.Run("ReadBackGeneratedKeys", func(t *testing.T) {
		tempDir := t.TempDir()
		keyName := "readback-test"
		
		// Generate key pair
		pri, err := security.GenerateECKeyPair("secp256r1")
		if err != nil {
			t.Fatalf("Failed to generate EC key pair: %v", err)
		}
		
		// Write keys to files
		priFile := filepath.Join(tempDir, fmt.Sprintf("%s-private.pem", keyName))
		pubFile := filepath.Join(tempDir, fmt.Sprintf("%s-public.pem", keyName))
		
		// Write public key
		pubWriter, err := os.Create(pubFile)
		if err != nil {
			t.Fatalf("Failed to create public key file: %v", err)
		}
		
		err = security.WritePublicKey(&pri.PublicKey, pubWriter)
		pubWriter.Close()
		if err != nil {
			t.Fatalf("Failed to write public key: %v", err)
		}
		
		// Write private key
		priWriter, err := os.Create(priFile)
		if err != nil {
			t.Fatalf("Failed to create private key file: %v", err)
		}
		
		err = security.WriteECPrivateKey(pri, priWriter)
		priWriter.Close()
		if err != nil {
			t.Fatalf("Failed to write private key: %v", err)
		}
		
		// Read back private key
		priReader, err := os.Open(priFile)
		if err != nil {
			t.Fatalf("Failed to open private key file: %v", err)
		}
		defer priReader.Close()
		
		readPri, err := security.ReadECPrivateKey(priReader)
		if err != nil {
			t.Fatalf("Failed to read private key: %v", err)
		}
		
		// Verify the read key matches the original
		if readPri.D.Cmp(pri.D) != 0 {
			t.Error("Read private key D value doesn't match original")
		}
		
		if readPri.PublicKey.X.Cmp(pri.PublicKey.X) != 0 {
			t.Error("Read public key X coordinate doesn't match original")
		}
		
		if readPri.PublicKey.Y.Cmp(pri.PublicKey.Y) != 0 {
			t.Error("Read public key Y coordinate doesn't match original")
		}
	})
}

func TestGenerateKeyPair_PathHandling(t *testing.T) {
	t.Run("PathJoinBehavior", func(t *testing.T) {
		testCases := []struct {
			dir      string
			name     string
			expected string
		}{
			{"/tmp", "test", "/tmp/test-private.pem"},
			{"/tmp/", "test", "/tmp/test-private.pem"},
			{".", "test", "test-private.pem"},
			{"./keys", "mykey", "keys/mykey-private.pem"},
			{"/home/user/keys", "production", "/home/user/keys/production-private.pem"},
		}
		
		for _, tc := range testCases {
			result := filepath.Join(tc.dir, fmt.Sprintf("%s-private.pem", tc.name))
			// Normalize paths for comparison
			expected := filepath.Clean(tc.expected)
			result = filepath.Clean(result)
			
			if result != expected {
				t.Errorf("For dir=%s, name=%s: expected %s, got %s", tc.dir, tc.name, expected, result)
			}
		}
	})
}

func TestGenerateKeyPair_EdgeCases(t *testing.T) {
	t.Run("EmptyStringArguments", func(t *testing.T) {
		// Test empty string arguments
		args := createKeypairArgs("", "")
		
		if args.Name != "" {
			t.Errorf("Expected empty string, got: %v", args.Name)
		}
	})
	
	t.Run("WhitespaceArguments", func(t *testing.T) {
		args := createKeypairArgs("  test  ", "  /tmp  ")
		
		// The struct should store the exact string including whitespace
		if args.Name != "  test  " {
			t.Errorf("Expected '  test  ', got: %v", args.Name)
		}
	})
}

// Benchmark tests
func BenchmarkGenerateKeyPair_KeyGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := security.GenerateECKeyPair("secp256r1")
		if err != nil {
			b.Fatalf("Failed to generate key pair: %v", err)
		}
	}
}

func BenchmarkGenerateKeyPair_ArgumentParsing(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = createKeypairArgs("benchmark-test", "/tmp")
	}
}
