package script

import (
	"bufio"
	"io"
	"io/fs"
	"os"
)

type NJSFile struct {
}

func (js *NJSFile) ReadFileAsLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func (js *NJSFile) ReadFileAsString(filePath string) (*string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
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
