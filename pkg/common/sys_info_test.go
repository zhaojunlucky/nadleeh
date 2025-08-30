package common

import (
	"fmt"
	"testing"
)

func Test_SysInfo(t *testing.T) {
	fmt.Println(Sys.GetInfo().GetAll())
}
