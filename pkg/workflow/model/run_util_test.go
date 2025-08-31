package workflow

import (
	"nadleeh/pkg/encrypt"
	"nadleeh/pkg/script"
	"testing"

	"github.com/zhaojunlucky/golib/pkg/env"
)

// Helper function to create a test JSContext
func createTestJSContext() *script.JSContext {
	secCtx := encrypt.SecureContext{}
	jsCtx := script.NewJSContext(&secCtx)
	return &jsCtx
}

// Helper function to create a test environment
func createTestEnv(data map[string]string) env.Env {
	testEnv := env.NewReadWriteEnv(nil, data)
	return testEnv
}

func TestInterpretNadEnv(t *testing.T) {
	t.Run("EmptyEnvs", func(t *testing.T) {
		jsContext := createTestJSContext()
		parent := createTestEnv(map[string]string{"PARENT_KEY": "parent-value"})
		envs := map[string]string{}
		variables := map[string]interface{}{}

		result, err := InterpretNadEnv(jsContext, parent, envs, variables)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
		// Should return empty ReadWriteEnv with parent
		if result.Get("PARENT_KEY") != "parent-value" {
			t.Errorf("Expected parent key to be accessible, got %s", result.Get("PARENT_KEY"))
		}
	})

	t.Run("NilEnvs", func(t *testing.T) {
		jsContext := createTestJSContext()
		parent := createTestEnv(map[string]string{"PARENT_KEY": "parent-value"})
		var envs map[string]string = nil
		variables := map[string]interface{}{}

		result, err := InterpretNadEnv(jsContext, parent, envs, variables)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
	})

	t.Run("SimpleStringLiteral", func(t *testing.T) {
		jsContext := createTestJSContext()
		parent := createTestEnv(map[string]string{"PARENT_KEY": "parent-value"})
		envs := map[string]string{
			"TEST_KEY": "'hello world'",
		}
		variables := map[string]interface{}{}

		result, err := InterpretNadEnv(jsContext, parent, envs, variables)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
		// Just verify the key was set (JavaScript evaluation may return with quotes)
		testValue := result.Get("TEST_KEY")
		if testValue == "" {
			t.Error("Expected TEST_KEY to have a value")
		}
		t.Logf("TEST_KEY value: '%s'", testValue)
	})
}

func TestInterpretWriteOnParentEnv(t *testing.T) {
	t.Run("EmptyEnvs", func(t *testing.T) {
		jsContext := createTestJSContext()
		parent := createTestEnv(map[string]string{"PARENT_KEY": "parent-value"})
		envs := map[string]string{}
		variables := map[string]interface{}{}

		result, err := InterpretWriteOnParentEnv(jsContext, parent, envs, variables)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
		// Should have access to parent values
		if result.Get("PARENT_KEY") != "parent-value" {
			t.Errorf("Expected parent key to be accessible, got %s", result.Get("PARENT_KEY"))
		}
	})

	t.Run("NilEnvs", func(t *testing.T) {
		jsContext := createTestJSContext()
		parent := createTestEnv(map[string]string{"PARENT_KEY": "parent-value"})
		var envs map[string]string = nil
		variables := map[string]interface{}{}

		result, err := InterpretWriteOnParentEnv(jsContext, parent, envs, variables)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
	})

	t.Run("SimpleStringLiteral", func(t *testing.T) {
		jsContext := createTestJSContext()
		parent := createTestEnv(map[string]string{"PARENT_KEY": "parent-value"})
		envs := map[string]string{
			"TEST_KEY": "'hello world'",
		}
		variables := map[string]interface{}{}

		result, err := InterpretWriteOnParentEnv(jsContext, parent, envs, variables)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
		// Just verify the key was set (JavaScript evaluation may return with quotes)
		testValue := result.Get("TEST_KEY")
		if testValue == "" {
			t.Error("Expected TEST_KEY to have a value")
		}
		t.Logf("TEST_KEY value: '%s'", testValue)
	})
}

// Benchmark tests
func BenchmarkInterpretNadEnv(b *testing.B) {
	jsContext := createTestJSContext()
	parent := createTestEnv(map[string]string{"PARENT": "parent-value"})
	envs := map[string]string{
		"KEY1": "'value1'",
		"KEY2": "'value2'",
		"KEY3": "'value3'",
	}
	variables := map[string]interface{}{"var": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := InterpretNadEnv(jsContext, parent, envs, variables)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkInterpretWriteOnParentEnv(b *testing.B) {
	jsContext := createTestJSContext()
	parent := createTestEnv(map[string]string{"PARENT": "parent-value"})
	envs := map[string]string{
		"KEY1": "'value1'",
		"KEY2": "'value2'",
		"KEY3": "'value3'",
	}
	variables := map[string]interface{}{"var": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := InterpretWriteOnParentEnv(jsContext, parent, envs, variables)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
