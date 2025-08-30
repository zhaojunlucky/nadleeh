package file

import (
	"fmt"
	"os"
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
