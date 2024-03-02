package workflow

import (
	"fmt"
	"testing"
)

func TestWorkflow(t *testing.T) {
	f := "/Users/jun/magicworldz/github/nadleeh/cmd/backup.yml"
	workflow, err := ParseWorkflow(f)

	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(workflow)
}
