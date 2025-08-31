package common

import (
	"testing"

	"github.com/zhaojunlucky/golib/pkg/env"
)

func Test_WriteOnParentEnv(t *testing.T) {
	os := env.NewEmptyRWEnv()

	os.Set("AA", "BB")

	writeOnParentEnv := NewWriteOnParentEnv(os, map[string]string{"CC": "DD"})

	if writeOnParentEnv.Get("AA") != "BB" {
		t.Errorf("expect AA != BB")
	}

	if writeOnParentEnv.Get("CC") != "DD" {
		t.Errorf("expect CC != DD")
	}

	writeOnParentEnv.Set("EE", "FF")

	if writeOnParentEnv.Get("EE") != "FF" {
		t.Errorf("writeOnParentEnv expect EE != FF")
	}

	if os.Get("EE") != "FF" {
		t.Errorf("rwenv expect EE != FF")
	}
}
