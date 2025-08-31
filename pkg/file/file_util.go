package file

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func DirExists(path string) (bool, error) {
	fiInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if fiInfo.IsDir() {
		return true, nil
	} else {
		return false, fmt.Errorf("path %s is a file", path)
	}
}

func FileExists(path string) (bool, error) {
	fiInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if !fiInfo.IsDir() {
		return true, nil
	} else {
		return false, fmt.Errorf("path %s is a file", path)
	}
}

func LogFileWithLineNo(name, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return LogStrWithLineNo(name, string(data))
}

func LogStrWithLineNo(name, str string) error {
	lines := strings.Split(str, "\n")

	// Iterate over the slice with the index and value.
	// The index `i` gives us the line number.
	fmt.Printf("======%s file======\n", name)
	for i, line := range lines {
		// A common issue is an empty string at the beginning due to the initial newline.
		// This condition skips any empty lines.
		if line != "" {
			fmt.Printf("%d: %s\n", i+1, line)
		}
	}
	fmt.Printf("======end %s file======\n", name)
	return nil
}

func GetProjectRootDir() (string, error) {
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// The command returns the path to go.mod. We need its parent directory.
	goModPath := strings.TrimSpace(string(out))
	if goModPath == "" {
		return "", err
	}

	// Get the directory of the go.mod file.
	projectRoot := filepath.Dir(goModPath)
	return projectRoot, nil
}
