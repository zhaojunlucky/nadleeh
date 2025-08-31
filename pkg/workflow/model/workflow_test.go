package workflow

import (
	"fmt"
	"nadleeh/pkg/file"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkflow(t *testing.T) {
	root, err := file.GetProjectRootDir()
	if err != nil {
		t.Fatal(err)
	}
	ymlFile, err := os.Open(filepath.Join(root, "examples/backup.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow, err := ParseWorkflow(ymlFile)

	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(workflow)
}
