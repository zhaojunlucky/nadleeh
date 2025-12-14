package script

import (
	"runtime"
	"testing"
)

func TestNJSCore_RunCmd(t *testing.T) {
	core := &NJSCore{}

	t.Run("successful command", func(t *testing.T) {
		var cmd string
		var args []string
		
		if runtime.GOOS == "windows" {
			cmd = "cmd"
			args = []string{"/c", "echo", "hello world"}
		} else {
			cmd = "echo"
			args = []string{"hello world"}
		}

		result := core.RunCmd(cmd, &args, nil)

		if result.Status != 0 {
			t.Errorf("Expected status 0, got %d", result.Status)
		}
		if result.Stdout == "" {
			t.Error("Expected stdout to contain output")
		}
		if result.Stderr != "" {
			t.Errorf("Expected empty stderr, got: %s", result.Stderr)
		}
	})

	t.Run("command with exit code", func(t *testing.T) {
		var cmd string
		var args []string
		
		if runtime.GOOS == "windows" {
			cmd = "cmd"
			args = []string{"/c", "exit", "42"}
		} else {
			cmd = "sh"
			args = []string{"-c", "exit 42"}
		}

		result := core.RunCmd(cmd, &args, nil)

		if result.Status != 42 {
			t.Errorf("Expected status 42, got %d", result.Status)
		}
	})

	t.Run("nonexistent command", func(t *testing.T) {
		result := core.RunCmd("nonexistent-command-12345", nil, nil)

		if result.Status == 0 {
			t.Error("Expected non-zero status for nonexistent command")
		}
		// Should get status 255 for non-ExitError cases
		if result.Status != 255 {
			t.Errorf("Expected status 255 for nonexistent command, got %d", result.Status)
		}
	})

	t.Run("command with stderr output", func(t *testing.T) {
		var cmd string
		var args []string
		
		if runtime.GOOS == "windows" {
			cmd = "cmd"
			args = []string{"/c", "echo error message 1>&2 && exit 1"}
		} else {
			cmd = "sh"
			args = []string{"-c", "echo 'error message' >&2; exit 1"}
		}

		result := core.RunCmd(cmd, &args, nil)

		if result.Status != 1 {
			t.Errorf("Expected status 1, got %d", result.Status)
		}
		if result.Stderr == "" {
			t.Error("Expected stderr to contain error message")
		}
	})

	t.Run("command with both stdout and stderr", func(t *testing.T) {
		var cmd string
		var args []string
		
		if runtime.GOOS == "windows" {
			cmd = "cmd"
			args = []string{"/c", "echo stdout && echo stderr 1>&2"}
		} else {
			cmd = "sh"
			args = []string{"-c", "echo 'stdout'; echo 'stderr' >&2"}
		}

		result := core.RunCmd(cmd, &args, nil)

		if result.Status != 0 {
			t.Errorf("Expected status 0, got %d", result.Status)
		}
		if result.Stdout == "" {
			t.Error("Expected stdout to contain output")
		}
		if result.Stderr == "" {
			t.Error("Expected stderr to contain output")
		}
	})

	t.Run("empty command name", func(t *testing.T) {
		result := core.RunCmd("", nil, nil)

		if result.Status == 0 {
			t.Error("Expected non-zero status for empty command")
		}
	})

	t.Run("nil args", func(t *testing.T) {
		var cmd string
		
		if runtime.GOOS == "windows" {
			cmd = "cmd"
		} else {
			cmd = "echo"
		}

		result := core.RunCmd(cmd, nil, nil)

		// Should not panic and should handle nil args gracefully
		if result == nil {
			t.Error("Expected result to not be nil")
		}
	})
}
