package shell

import (
	"bytes"
	"os"
	"strings"
	"sync"
)

type SdtOutputWriter struct {
	b      bytes.Buffer
	mu     sync.Mutex
	out    *os.File
	masker *OutputMasker
}

type OutputMasker struct {
	values []string
}

func NewOutputMasker(envs map[string]string) *OutputMasker {
	seen := make(map[string]struct{})
	var values []string
	for key, value := range envs {
		if !isSensitiveKey(key) || len(value) < 4 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return &OutputMasker{values: values}
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	return strings.Contains(key, "secret") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "password") ||
		strings.Contains(key, "private") ||
		strings.Contains(key, "credential")
}

func (m *OutputMasker) Mask(value string) string {
	if m == nil {
		return value
	}
	for _, secret := range m.values {
		value = strings.ReplaceAll(value, secret, "***")
	}
	return value
}

func (w *SdtOutputWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	masked := w.masker.Mask(string(p))
	if w.out != nil {
		if _, err = w.out.Write([]byte(masked)); err != nil {
			return 0, err
		}
	}
	_, err = w.b.WriteString(masked)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *SdtOutputWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.String()
}

func NewStdOutputWriter(maskers ...*OutputMasker) *SdtOutputWriter {
	var masker *OutputMasker
	if len(maskers) > 0 {
		masker = maskers[0]
	}
	return &SdtOutputWriter{
		b:      bytes.Buffer{},
		out:    os.Stdout,
		masker: masker,
	}
}
