package shell

import (
	"bytes"
	"fmt"
)

type SdtOutputWriter struct {
	b bytes.Buffer
}

func (w SdtOutputWriter) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return w.b.Write(p)
}

func (w SdtOutputWriter) String() string {
	return w.b.String()
}

func NewStdOutputWriter() *SdtOutputWriter {
	return &SdtOutputWriter{
		b: bytes.Buffer{},
	}
}
