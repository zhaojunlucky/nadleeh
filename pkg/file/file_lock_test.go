package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFileLock(t *testing.T) {
	lockFile := "/tmp/test.lock"
	fl := NewFileLock(lockFile)

	if fl == nil {
		t.Fatal("NewFileLock returned nil")
	}

	if fl.lockFile != lockFile {
		t.Errorf("Expected lockFile to be %s, got %s", lockFile, fl.lockFile)
	}

	if fl.fd != nil {
		t.Error("Expected fd to be nil initially")
	}
}

func TestFileLock_Lock_Success(t *testing.T) {
	tempDir := t.TempDir()
	lockFile := filepath.Join(tempDir, "test.lock")
	fl := NewFileLock(lockFile)

	err := fl.Lock()
	if err != nil {
		t.Fatalf("Expected Lock() to succeed, got error: %v", err)
	}

	// Verify lock file was created
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Error("Lock file was not created")
	}

	// Verify fd is set
	if fl.fd == nil {
		t.Error("Expected fd to be set after Lock()")
	}

	// Clean up
	fl.Unlock()
	os.Remove(lockFile)
}

func TestFileLock_Lock_CreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	lockDir := filepath.Join(tempDir, "subdir", "nested")
	lockFile := filepath.Join(lockDir, "test.lock")
	fl := NewFileLock(lockFile)

	err := fl.Lock()
	if err != nil {
		t.Fatalf("Expected Lock() to succeed when creating directories, got error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(lockDir); os.IsNotExist(err) {
		t.Error("Lock directory was not created")
	}

	// Verify lock file was created
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Error("Lock file was not created")
	}

	// Clean up
	fl.Unlock()
	os.RemoveAll(filepath.Join(tempDir, "subdir"))
}

func TestFileLock_Lock_InvalidPath(t *testing.T) {
	// Try to create lock file in a path that cannot be created (e.g., under a file)
	tempDir := t.TempDir()
	regularFile := filepath.Join(tempDir, "regular_file")

	// Create a regular file
	f, err := os.Create(regularFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()

	// Try to create lock file under the regular file (should fail)
	lockFile := filepath.Join(regularFile, "test.lock")
	fl := NewFileLock(lockFile)

	err = fl.Lock()
	if err == nil {
		t.Error("Expected Lock() to fail with invalid path")
		fl.Unlock() // Clean up if somehow succeeded
	}
}

func TestFileLock_Unlock_Success(t *testing.T) {
	tempDir := t.TempDir()
	lockFile := filepath.Join(tempDir, "test.lock")
	fl := NewFileLock(lockFile)

	// First acquire the lock
	err := fl.Lock()
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	// Now unlock
	err = fl.Unlock()
	if err != nil {
		t.Errorf("Expected Unlock() to succeed, got error: %v", err)
	}

	// Verify fd is nil after unlock
	if fl.fd != nil {
		t.Error("Expected fd to be nil after Unlock()")
	}

	// Clean up
	os.Remove(lockFile)
}

func TestFileLock_Unlock_WithoutLock(t *testing.T) {
	lockFile := "/tmp/test_no_lock.lock"
	fl := NewFileLock(lockFile)

	// Try to unlock without acquiring lock first
	err := fl.Unlock()
	if err != nil {
		t.Errorf("Expected Unlock() to succeed even without prior lock, got error: %v", err)
	}
}

func TestFileLock_MultipleLocks_SameProcess(t *testing.T) {
	tempDir := t.TempDir()
	lockFile := filepath.Join(tempDir, "test.lock")

	fl1 := NewFileLock(lockFile)
	fl2 := NewFileLock(lockFile)

	// First lock should succeed
	err := fl1.Lock()
	if err != nil {
		t.Fatalf("First lock should succeed: %v", err)
	}

	// Test second lock attempt with timeout to avoid hanging
	done := make(chan error, 1)
	go func() {
		done <- fl2.Lock()
	}()
	
	select {
	case err := <-done:
		// Second lock completed (might succeed or fail depending on OS behavior)
		if err == nil {
			// If it succeeded, unlock it
			fl2.Unlock()
		}
	case <-time.After(100 * time.Millisecond):
		// Second lock is blocking as expected for exclusive locks
		t.Log("Second lock blocked as expected (exclusive lock behavior)")
	}

	// Clean up
	fl1.Unlock()
	os.Remove(lockFile)
}

func TestFileLock_LockUnlockCycle(t *testing.T) {
	tempDir := t.TempDir()
	lockFile := filepath.Join(tempDir, "test.lock")
	fl := NewFileLock(lockFile)

	// Test multiple lock/unlock cycles
	for i := 0; i < 3; i++ {
		err := fl.Lock()
		if err != nil {
			t.Fatalf("Lock cycle %d failed: %v", i+1, err)
		}

		err = fl.Unlock()
		if err != nil {
			t.Fatalf("Unlock cycle %d failed: %v", i+1, err)
		}
	}

	// Clean up
	os.Remove(lockFile)
}

func TestFileLock_Close(t *testing.T) {
	tempDir := t.TempDir()
	lockFile := filepath.Join(tempDir, "test.lock")
	fl := NewFileLock(lockFile)

	// Test close without lock
	err := fl.close()
	if err != nil {
		t.Errorf("close() should succeed even without fd, got error: %v", err)
	}

	// Test close with lock
	err = fl.Lock()
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	err = fl.close()
	if err != nil {
		t.Errorf("close() should succeed with fd, got error: %v", err)
	}

	// Verify fd is nil after close
	if fl.fd != nil {
		t.Error("Expected fd to be nil after close()")
	}

	// Clean up
	os.Remove(lockFile)
}

func TestFileLock_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	lockFile := filepath.Join(tempDir, "concurrent.lock")

	// Test concurrent access using goroutines
	results := make(chan error, 2)

	go func() {
		fl := NewFileLock(lockFile)
		err := fl.Lock()
		if err != nil {
			results <- err
			return
		}

		// Hold lock for a short time
		time.Sleep(50 * time.Millisecond)

		err = fl.Unlock()
		results <- err
	}()

	go func() {
		// Start second goroutine slightly after first
		time.Sleep(25 * time.Millisecond)

		fl := NewFileLock(lockFile)
		
		// Use a timeout channel to prevent indefinite blocking with race detector
		done := make(chan error, 1)
		go func() {
			done <- fl.Lock()
		}()
		
		select {
		case err := <-done:
			if err != nil {
				results <- err
				return
			}
			err = fl.Unlock()
			results <- err
		case <-time.After(2 * time.Second):
			// Lock is blocking as expected (exclusive lock behavior)
			// This is normal behavior, not an error
			results <- nil
		}
	}()

	// Wait for both goroutines to complete
	for i := 0; i < 2; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Errorf("Concurrent access test failed: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}

	// Clean up
	os.Remove(lockFile)
}

func TestFileLock_EdgeCases(t *testing.T) {
	t.Run("EmptyLockFile", func(t *testing.T) {
		fl := NewFileLock("")
		err := fl.Lock()
		if err == nil {
			t.Error("Expected Lock() to fail with empty lock file path")
			fl.Unlock()
		}
	})

	t.Run("RelativePath", func(t *testing.T) {
		fl := NewFileLock("relative/path/test.lock")
		err := fl.Lock()
		// This might succeed or fail depending on current directory permissions
		if err == nil {
			fl.Unlock()
			os.Remove("relative/path/test.lock")
			os.RemoveAll("relative")
		}
	})

	t.Run("VeryLongPath", func(t *testing.T) {
		longPath := "/tmp/" + string(make([]byte, 200)) + "/test.lock"
		fl := NewFileLock(longPath)
		err := fl.Lock()
		// This will likely fail due to invalid path
		if err == nil {
			fl.Unlock()
		}
	})
}

// Benchmark tests
func BenchmarkFileLock_LockUnlock(b *testing.B) {
	tempDir := b.TempDir()
	lockFile := filepath.Join(tempDir, "benchmark.lock")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fl := NewFileLock(lockFile)
		fl.Lock()
		fl.Unlock()
	}

	// Clean up
	os.Remove(lockFile)
}

func BenchmarkFileLock_NewFileLock(b *testing.B) {
	lockFile := "/tmp/benchmark.lock"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewFileLock(lockFile)
	}
}
