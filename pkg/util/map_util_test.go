package util

import (
	"reflect"
	"testing"
)

func TestHasKey(t *testing.T) {
	t.Run("StringIntMap", func(t *testing.T) {
		m := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		
		// Test existing keys
		if !HasKey(m, "one") {
			t.Error("Expected HasKey to return true for existing key 'one'")
		}
		
		if !HasKey(m, "two") {
			t.Error("Expected HasKey to return true for existing key 'two'")
		}
		
		if !HasKey(m, "three") {
			t.Error("Expected HasKey to return true for existing key 'three'")
		}
		
		// Test non-existing keys
		if HasKey(m, "four") {
			t.Error("Expected HasKey to return false for non-existing key 'four'")
		}
		
		if HasKey(m, "") {
			t.Error("Expected HasKey to return false for empty string key")
		}
	})
	
	t.Run("IntStringMap", func(t *testing.T) {
		m := map[int]string{
			1: "one",
			2: "two",
			3: "three",
		}
		
		// Test existing keys
		if !HasKey(m, 1) {
			t.Error("Expected HasKey to return true for existing key 1")
		}
		
		if !HasKey(m, 2) {
			t.Error("Expected HasKey to return true for existing key 2")
		}
		
		// Test non-existing keys
		if HasKey(m, 0) {
			t.Error("Expected HasKey to return false for non-existing key 0")
		}
		
		if HasKey(m, 4) {
			t.Error("Expected HasKey to return false for non-existing key 4")
		}
	})
	
	t.Run("EmptyMap", func(t *testing.T) {
		m := make(map[string]int)
		
		if HasKey(m, "any") {
			t.Error("Expected HasKey to return false for any key in empty map")
		}
		
		if HasKey(m, "") {
			t.Error("Expected HasKey to return false for empty string in empty map")
		}
	})
	
	t.Run("NilMap", func(t *testing.T) {
		var m map[string]int
		
		if HasKey(m, "any") {
			t.Error("Expected HasKey to return false for any key in nil map")
		}
	})
	
	t.Run("ZeroValues", func(t *testing.T) {
		m := map[string]int{
			"zero":  0,
			"empty": 0,
		}
		
		// Should return true even if value is zero
		if !HasKey(m, "zero") {
			t.Error("Expected HasKey to return true for key with zero value")
		}
		
		if !HasKey(m, "empty") {
			t.Error("Expected HasKey to return true for key with zero value")
		}
	})
	
	t.Run("BoolKeyMap", func(t *testing.T) {
		m := map[bool]string{
			true:  "yes",
			false: "no",
		}
		
		if !HasKey(m, true) {
			t.Error("Expected HasKey to return true for boolean key true")
		}
		
		if !HasKey(m, false) {
			t.Error("Expected HasKey to return true for boolean key false")
		}
	})
	
	t.Run("StructValueMap", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		
		m := map[string]Person{
			"alice": {Name: "Alice", Age: 30},
			"bob":   {Name: "Bob", Age: 25},
		}
		
		if !HasKey(m, "alice") {
			t.Error("Expected HasKey to return true for existing key 'alice'")
		}
		
		if HasKey(m, "charlie") {
			t.Error("Expected HasKey to return false for non-existing key 'charlie'")
		}
	})
	
	t.Run("InterfaceValueMap", func(t *testing.T) {
		m := map[string]interface{}{
			"string": "hello",
			"int":    42,
			"bool":   true,
			"nil":    nil,
		}
		
		if !HasKey(m, "string") {
			t.Error("Expected HasKey to return true for string value")
		}
		
		if !HasKey(m, "int") {
			t.Error("Expected HasKey to return true for int value")
		}
		
		if !HasKey(m, "bool") {
			t.Error("Expected HasKey to return true for bool value")
		}
		
		if !HasKey(m, "nil") {
			t.Error("Expected HasKey to return true for nil value")
		}
		
		if HasKey(m, "missing") {
			t.Error("Expected HasKey to return false for missing key")
		}
	})
}

func TestCopyMap(t *testing.T) {
	t.Run("StringIntMap", func(t *testing.T) {
		original := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		
		copied := CopyMap(original)
		
		// Check that all keys and values are copied
		if !reflect.DeepEqual(original, copied) {
			t.Errorf("Copied map does not match original. Original: %v, Copied: %v", original, copied)
		}
		
		// Check that it's a different map (different memory address)
		if &original == &copied {
			t.Error("Copied map should be a different instance")
		}
		
		// Modify original and ensure copy is not affected
		original["four"] = 4
		if _, exists := copied["four"]; exists {
			t.Error("Modifying original map should not affect copied map")
		}
		
		// Modify copy and ensure original is not affected
		copied["five"] = 5
		if _, exists := original["five"]; exists {
			t.Error("Modifying copied map should not affect original map")
		}
	})
	
	t.Run("EmptyMap", func(t *testing.T) {
		original := make(map[string]int)
		copied := CopyMap(original)
		
		if len(copied) != 0 {
			t.Errorf("Expected copied empty map to have length 0, got %d", len(copied))
		}
		
		// Add to original, should not affect copy
		original["test"] = 1
		if len(copied) != 0 {
			t.Error("Adding to original empty map should not affect copy")
		}
	})
	
	t.Run("NilMap", func(t *testing.T) {
		var original map[string]int
		copied := CopyMap(original)
		
		if copied == nil {
			t.Error("CopyMap of nil map should return empty map, not nil")
		}
		
		if len(copied) != 0 {
			t.Errorf("Expected copied nil map to have length 0, got %d", len(copied))
		}
	})
	
	t.Run("IntStringMap", func(t *testing.T) {
		original := map[int]string{
			1: "one",
			2: "two",
			3: "three",
		}
		
		copied := CopyMap(original)
		
		if !reflect.DeepEqual(original, copied) {
			t.Errorf("Copied map does not match original. Original: %v, Copied: %v", original, copied)
		}
		
		// Test independence
		original[4] = "four"
		if copied[4] == "four" {
			t.Error("Copied map should be independent of original")
		}
	})
	
	t.Run("BoolKeyMap", func(t *testing.T) {
		original := map[bool]string{
			true:  "yes",
			false: "no",
		}
		
		copied := CopyMap(original)
		
		if !reflect.DeepEqual(original, copied) {
			t.Errorf("Copied map does not match original. Original: %v, Copied: %v", original, copied)
		}
	})
	
	t.Run("StructValueMap", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		
		original := map[string]Person{
			"alice": {Name: "Alice", Age: 30},
			"bob":   {Name: "Bob", Age: 25},
		}
		
		copied := CopyMap(original)
		
		if !reflect.DeepEqual(original, copied) {
			t.Errorf("Copied map does not match original. Original: %v, Copied: %v", original, copied)
		}
		
		// Modify original struct value
		alice := original["alice"]
		alice.Age = 31
		original["alice"] = alice
		
		// Copy should not be affected (shallow copy behavior)
		if copied["alice"].Age == 31 {
			t.Error("Modifying struct in original should not affect copy (shallow copy expected)")
		}
	})
	
	t.Run("PointerValueMap", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		
		alice := &Person{Name: "Alice", Age: 30}
		bob := &Person{Name: "Bob", Age: 25}
		
		original := map[string]*Person{
			"alice": alice,
			"bob":   bob,
		}
		
		copied := CopyMap(original)
		
		if !reflect.DeepEqual(original, copied) {
			t.Errorf("Copied map does not match original. Original: %v, Copied: %v", original, copied)
		}
		
		// Both maps should point to the same Person objects (shallow copy)
		if original["alice"] != copied["alice"] {
			t.Error("Pointer values should be the same in both maps (shallow copy)")
		}
		
		// Modifying the pointed-to object should affect both maps
		alice.Age = 31
		if copied["alice"].Age != 31 {
			t.Error("Modifying pointed-to object should affect both maps")
		}
	})
	
	t.Run("InterfaceValueMap", func(t *testing.T) {
		original := map[string]interface{}{
			"string": "hello",
			"int":    42,
			"bool":   true,
			"nil":    nil,
		}
		
		copied := CopyMap(original)
		
		if !reflect.DeepEqual(original, copied) {
			t.Errorf("Copied map does not match original. Original: %v, Copied: %v", original, copied)
		}
		
		// Test independence
		original["new"] = "value"
		if _, exists := copied["new"]; exists {
			t.Error("Adding to original should not affect copy")
		}
	})
	
	t.Run("LargeMap", func(t *testing.T) {
		original := make(map[int]int)
		for i := 0; i < 1000; i++ {
			original[i] = i * i
		}
		
		copied := CopyMap(original)
		
		if len(copied) != len(original) {
			t.Errorf("Expected copied map length %d, got %d", len(original), len(copied))
		}
		
		if !reflect.DeepEqual(original, copied) {
			t.Error("Large map copy does not match original")
		}
	})
}

func TestMapUtilsIntegration(t *testing.T) {
	t.Run("HasKeyOnCopiedMap", func(t *testing.T) {
		original := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		
		copied := CopyMap(original)
		
		// HasKey should work the same on both maps
		if HasKey(original, "one") != HasKey(copied, "one") {
			t.Error("HasKey should return same result for original and copied map")
		}
		
		if HasKey(original, "missing") != HasKey(copied, "missing") {
			t.Error("HasKey should return same result for missing keys in both maps")
		}
		
		// Add key to copy only
		copied["four"] = 4
		
		if HasKey(original, "four") {
			t.Error("Key added to copy should not exist in original")
		}
		
		if !HasKey(copied, "four") {
			t.Error("Key added to copy should exist in copy")
		}
	})
}

// Benchmark tests
func BenchmarkHasKey(b *testing.B) {
	m := make(map[string]int)
	for i := 0; i < 1000; i++ {
		m[string(rune('a'+i%26))+string(rune('a'+(i/26)%26))] = i
	}
	
	b.Run("ExistingKey", func(b *testing.B) {
		key := "aa"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			HasKey(m, key)
		}
	})
	
	b.Run("NonExistingKey", func(b *testing.B) {
		key := "zzzz"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			HasKey(m, key)
		}
	})
}

func BenchmarkCopyMap(b *testing.B) {
	b.Run("SmallMap", func(b *testing.B) {
		m := map[string]int{
			"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CopyMap(m)
		}
	})
	
	b.Run("MediumMap", func(b *testing.B) {
		m := make(map[int]int)
		for i := 0; i < 100; i++ {
			m[i] = i * i
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CopyMap(m)
		}
	})
	
	b.Run("LargeMap", func(b *testing.B) {
		m := make(map[int]int)
		for i := 0; i < 10000; i++ {
			m[i] = i * i
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CopyMap(m)
		}
	})
	
	b.Run("EmptyMap", func(b *testing.B) {
		m := make(map[string]int)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CopyMap(m)
		}
	})
}

func BenchmarkMapUtilsChaining(b *testing.B) {
	original := make(map[string]int)
	for i := 0; i < 100; i++ {
		original[string(rune('a'+i%26))+string(rune('a'+(i/26)%26))] = i
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copied := CopyMap(original)
		HasKey(copied, "aa")
	}
}
