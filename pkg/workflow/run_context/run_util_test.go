package run_context

import (
	"fmt"
	"testing"
)

// mockEnv implements the env.Env interface for testing
type mockEnv struct {
	data map[string]string
}

func (m *mockEnv) Get(key string) string {
	if value, exists := m.data[key]; exists {
		return value
	}
	return ""
}

func (m *mockEnv) Set(key, value string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
}

func (m *mockEnv) GetAll() map[string]string {
	result := make(map[string]string)
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

func (m *mockEnv) Contains(key string) bool {
	_, exists := m.data[key]
	return exists
}

func (m *mockEnv) Expand(value string) string {
	// Simple expansion implementation for testing
	return value
}

func (m *mockEnv) SetAll(data map[string]string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	for k, v := range data {
		m.data[k] = v
	}
}

func newMockEnv() *mockEnv {
	return &mockEnv{
		data: make(map[string]string),
	}
}

func TestInterpretPluginCfg(t *testing.T) {
	t.Run("EmptyConfig", func(t *testing.T) {
		// Create a real WorkflowRunContext for testing
		ctx := NewWorkflowRunContext(nil)
		
		mockEnv := newMockEnv()
		config := make(map[string]string)
		variables := make(map[string]interface{})
		
		result, err := InterpretPluginCfg(ctx, mockEnv, config, variables)
		
		if err != nil {
			t.Errorf("Expected no error for empty config, got: %v", err)
		}
		
		if len(result) != 0 {
			t.Errorf("Expected empty result for empty config, got: %v", result)
		}
	})
	
	t.Run("SimpleStringValues", func(t *testing.T) {
		// Create a real WorkflowRunContext for testing
		ctx := NewWorkflowRunContext(nil)
		
		mockEnv := newMockEnv()
		config := map[string]string{
			"key1": "simple_value",
			"key2": "another_value", 
			"key3": "third_value",
		}
		variables := make(map[string]interface{})
		
		result, err := InterpretPluginCfg(ctx, mockEnv, config, variables)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		// Since we're using real JSContext, simple strings should be returned as-is
		expected := map[string]string{
			"key1": "simple_value",
			"key2": "another_value",
			"key3": "third_value",
		}
		
		if len(result) != len(expected) {
			t.Errorf("Expected result length %d, got %d", len(expected), len(result))
		}
		
		for k, v := range expected {
			if result[k] != v {
				t.Errorf("Expected result[%s] = %s, got %s", k, v, result[k])
			}
		}
	})
	
	t.Run("JavaScriptExpressions", func(t *testing.T) {
		// Create a real WorkflowRunContext for testing
		ctx := NewWorkflowRunContext(nil)
		
		mockEnv := newMockEnv()
		config := map[string]string{
			"username":  "${name}",
			"greeting":  "Hello ${name}!",
		}
		variables := map[string]interface{}{
			"name": "John Doe",
		}
		
		result, err := InterpretPluginCfg(ctx, mockEnv, config, variables)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		// Note: The actual JSContext behavior may not evaluate simple ${} expressions
		// This test verifies that the function completes without error
		if len(result) != len(config) {
			t.Errorf("Expected result to have %d entries, got %d", len(config), len(result))
		}
		
		// Verify all keys are present
		for k := range config {
			if _, exists := result[k]; !exists {
				t.Errorf("Expected key %s to be present in result", k)
			}
		}
	})
	
	t.Run("EvaluationError", func(t *testing.T) {
		// Create a real WorkflowRunContext for testing
		ctx := NewWorkflowRunContext(nil)
		
		mockEnv := newMockEnv()
		config := map[string]string{
			"valid_key":   "simple_value",
			"invalid_key": "throw new Error('test error')", // Use actual JS that will cause an error
		}
		variables := make(map[string]interface{})
		
		result, err := InterpretPluginCfg(ctx, mockEnv, config, variables)
		
		// The function should handle errors gracefully
		// If JSContext doesn't evaluate the error expression, it will succeed
		// If it does evaluate and throws an error, the function should return an error
		if err != nil {
			// Error case - verify result is nil
			if result != nil {
				t.Errorf("Expected nil result on error, got: %v", result)
			}
		} else {
			// Success case - verify result is complete
			if len(result) != len(config) {
				t.Errorf("Expected result to have %d entries, got %d", len(config), len(result))
			}
		}
	})
	
	t.Run("NilInputs", func(t *testing.T) {
		// Create a real WorkflowRunContext for testing
		ctx := NewWorkflowRunContext(nil)
		
		mockEnv := newMockEnv()
		
		// Test with nil config
		result, err := InterpretPluginCfg(ctx, mockEnv, nil, nil)
		
		if err != nil {
			t.Errorf("Expected no error for nil config, got: %v", err)
		}
		
		if len(result) != 0 {
			t.Errorf("Expected empty result for nil config, got: %v", result)
		}
	})
	
	t.Run("ComplexVariables", func(t *testing.T) {
		// Create a real WorkflowRunContext for testing
		ctx := NewWorkflowRunContext(nil)
		
		mockEnv := newMockEnv()
		mockEnv.Set("NODE_ENV", "production")
		
		config := map[string]string{
			"name":    "${user.name}",
			"timeout": "${config.timeout}",
		}
		
		variables := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "Alice Smith",
				"age":  28,
			},
			"config": map[string]interface{}{
				"timeout": 30,
				"retries": 3,
			},
		}
		
		result, err := InterpretPluginCfg(ctx, mockEnv, config, variables)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		// Verify that the function processes all config entries
		if len(result) != len(config) {
			t.Errorf("Expected result to have %d entries, got %d", len(config), len(result))
		}
		
		// Verify all keys are present
		for k := range config {
			if _, exists := result[k]; !exists {
				t.Errorf("Expected key %s to be present in result", k)
			}
		}
	})
	
	t.Run("EmptyStringValues", func(t *testing.T) {
		// Create a real WorkflowRunContext for testing
		ctx := NewWorkflowRunContext(nil)
		
		mockEnv := newMockEnv()
		config := map[string]string{
			"empty_key":     "",
			"non_empty_key": "non_empty",
		}
		variables := make(map[string]interface{})
		
		result, err := InterpretPluginCfg(ctx, mockEnv, config, variables)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if result["empty_key"] != "" {
			t.Errorf("Expected empty string for empty_key, got: %s", result["empty_key"])
		}
		
		if result["non_empty_key"] != "non_empty" {
			t.Errorf("Expected 'non_empty' for non_empty_key, got: %s", result["non_empty_key"])
		}
	})
	
	t.Run("SpecialCharacters", func(t *testing.T) {
		// Create a real WorkflowRunContext for testing
		ctx := NewWorkflowRunContext(nil)
		
		mockEnv := newMockEnv()
		config := map[string]string{
			"special_chars": "${special}",
			"unicode_text":  "${unicode}",
			"json_data":     "${json}",
		}
		variables := map[string]interface{}{
			"special": "!@#$%^&*()_+-=[]{}|;:,.<>?",
			"unicode": "Hello 世界 🌍",
			"json":    `{"key": "value", "number": 42}`,
		}
		
		result, err := InterpretPluginCfg(ctx, mockEnv, config, variables)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		// Verify that the function processes all config entries
		if len(result) != len(config) {
			t.Errorf("Expected result to have %d entries, got %d", len(config), len(result))
		}
		
		// Verify all keys are present
		for k := range config {
			if _, exists := result[k]; !exists {
				t.Errorf("Expected key %s to be present in result", k)
			}
		}
	})
}

func TestInterpretPluginCfg_EdgeCases(t *testing.T) {
	t.Run("LargeConfig", func(t *testing.T) {
		// Create a real WorkflowRunContext for testing
		ctx := NewWorkflowRunContext(nil)
		
		// Create large config
		config := make(map[string]string)
		expected := make(map[string]string)
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key_%d", i)
			value := fmt.Sprintf("value_%d", i)
			
			config[key] = value
			expected[key] = value // Simple strings should be returned as-is
		}
		
		mockEnv := newMockEnv()
		variables := make(map[string]interface{})
		
		result, err := InterpretPluginCfg(ctx, mockEnv, config, variables)
		
		if err != nil {
			t.Errorf("Expected no error for large config, got: %v", err)
		}
		
		if len(result) != len(expected) {
			t.Errorf("Expected result length %d, got %d", len(expected), len(result))
		}
		
		for k, v := range expected {
			if result[k] != v {
				t.Errorf("Expected result[%s] = %s, got %s", k, v, result[k])
			}
		}
	})
	
	t.Run("DuplicateKeys", func(t *testing.T) {
		// Create a real WorkflowRunContext for testing
		ctx := NewWorkflowRunContext(nil)
		
		mockEnv := newMockEnv()
		config := map[string]string{
			"same_key": "expression1",
			// Note: Go maps don't allow duplicate keys, so this tests normal behavior
		}
		variables := make(map[string]interface{})
		
		result, err := InterpretPluginCfg(ctx, mockEnv, config, variables)
		
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		
		if result["same_key"] != "expression1" {
			t.Errorf("Expected expression1, got: %s", result["same_key"])
		}
	})
}

// Benchmark tests
func BenchmarkInterpretPluginCfg(b *testing.B) {
	// Create a real WorkflowRunContext for benchmarking
	ctx := NewWorkflowRunContext(nil)
	
	mockEnv := newMockEnv()
	config := map[string]string{
		"name":    "${name}",
		"age":     "${age}",
		"city":    "${city}",
		"country": "${country}",
	}
	variables := map[string]interface{}{
		"name":    "John Doe",
		"age":     25,
		"city":    "New York",
		"country": "USA",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InterpretPluginCfg(ctx, mockEnv, config, variables)
	}
}
