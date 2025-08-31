package util

import (
	"testing"
)

func TestBool2Int(t *testing.T) {
	t.Run("TrueValue", func(t *testing.T) {
		result := Bool2Int(true)
		expected := 1
		
		if result != expected {
			t.Errorf("Bool2Int(true) = %d; expected %d", result, expected)
		}
	})
	
	t.Run("FalseValue", func(t *testing.T) {
		result := Bool2Int(false)
		expected := 0
		
		if result != expected {
			t.Errorf("Bool2Int(false) = %d; expected %d", result, expected)
		}
	})
	
	t.Run("MultipleCallsConsistency", func(t *testing.T) {
		// Test that multiple calls with same input return same result
		for i := 0; i < 10; i++ {
			if Bool2Int(true) != 1 {
				t.Error("Bool2Int(true) should consistently return 1")
			}
			if Bool2Int(false) != 0 {
				t.Error("Bool2Int(false) should consistently return 0")
			}
		}
	})
}

func TestStr2Bool(t *testing.T) {
	t.Run("TrueString", func(t *testing.T) {
		result := Str2Bool("true")
		expected := true
		
		if result != expected {
			t.Errorf("Str2Bool(\"true\") = %t; expected %t", result, expected)
		}
	})
	
	t.Run("FalseString", func(t *testing.T) {
		result := Str2Bool("false")
		expected := false
		
		if result != expected {
			t.Errorf("Str2Bool(\"false\") = %t; expected %t", result, expected)
		}
	})
	
	t.Run("EmptyString", func(t *testing.T) {
		result := Str2Bool("")
		expected := false
		
		if result != expected {
			t.Errorf("Str2Bool(\"\") = %t; expected %t", result, expected)
		}
	})
	
	t.Run("CaseSensitivity", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected bool
		}{
			{"True", false},  // Capital T should return false
			{"TRUE", false},  // All caps should return false
			{"tRuE", false},  // Mixed case should return false
			{"true", true},   // Only exact lowercase "true" should return true
		}
		
		for _, tc := range testCases {
			result := Str2Bool(tc.input)
			if result != tc.expected {
				t.Errorf("Str2Bool(\"%s\") = %t; expected %t", tc.input, result, tc.expected)
			}
		}
	})
	
	t.Run("VariousStrings", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected bool
		}{
			{"1", false},
			{"0", false},
			{"yes", false},
			{"no", false},
			{"on", false},
			{"off", false},
			{"True", false},
			{"FALSE", false},
			{"t", false},
			{"f", false},
			{"true ", false}, // with trailing space
			{" true", false}, // with leading space
			{"true\n", false}, // with newline
			{"true\t", false}, // with tab
		}
		
		for _, tc := range testCases {
			result := Str2Bool(tc.input)
			if result != tc.expected {
				t.Errorf("Str2Bool(\"%s\") = %t; expected %t", tc.input, result, tc.expected)
			}
		}
	})
	
	t.Run("UnicodeStrings", func(t *testing.T) {
		testCases := []string{
			"真", // Chinese for "true"
			"правда", // Russian for "true"
			"vrai", // French for "true"
			"🔥", // Fire emoji
		}
		
		for _, input := range testCases {
			result := Str2Bool(input)
			if result != false {
				t.Errorf("Str2Bool(\"%s\") = %t; expected false", input, result)
			}
		}
	})
}

func TestInt2Bool(t *testing.T) {
	t.Run("PositiveValues", func(t *testing.T) {
		testCases := []int64{1, 2, 10, 100, 1000, 9223372036854775807} // max int64
		
		for _, val := range testCases {
			result := Int2Bool(val)
			if result != true {
				t.Errorf("Int2Bool(%d) = %t; expected true", val, result)
			}
		}
	})
	
	t.Run("ZeroValue", func(t *testing.T) {
		result := Int2Bool(0)
		expected := false
		
		if result != expected {
			t.Errorf("Int2Bool(0) = %t; expected %t", result, expected)
		}
	})
	
	t.Run("NegativeValues", func(t *testing.T) {
		testCases := []int64{-1, -2, -10, -100, -1000, -9223372036854775808} // min int64
		
		for _, val := range testCases {
			result := Int2Bool(val)
			if result != false {
				t.Errorf("Int2Bool(%d) = %t; expected false", val, result)
			}
		}
	})
	
	t.Run("BoundaryValues", func(t *testing.T) {
		testCases := []struct {
			input    int64
			expected bool
		}{
			{0, false},
			{1, true},
			{-1, false},
			{9223372036854775807, true},  // max int64
			{-9223372036854775808, false}, // min int64
		}
		
		for _, tc := range testCases {
			result := Int2Bool(tc.input)
			if result != tc.expected {
				t.Errorf("Int2Bool(%d) = %t; expected %t", tc.input, result, tc.expected)
			}
		}
	})
}

func TestConversionChaining(t *testing.T) {
	t.Run("Bool2Int2Bool", func(t *testing.T) {
		// Test chaining: bool -> int -> bool
		originalTrue := true
		originalFalse := false
		
		// true -> 1 -> true
		intFromTrue := Bool2Int(originalTrue)
		boolFromInt := Int2Bool(int64(intFromTrue))
		if boolFromInt != originalTrue {
			t.Errorf("Bool2Int2Bool chain failed for true: %t -> %d -> %t", originalTrue, intFromTrue, boolFromInt)
		}
		
		// false -> 0 -> false
		intFromFalse := Bool2Int(originalFalse)
		boolFromInt2 := Int2Bool(int64(intFromFalse))
		if boolFromInt2 != originalFalse {
			t.Errorf("Bool2Int2Bool chain failed for false: %t -> %d -> %t", originalFalse, intFromFalse, boolFromInt2)
		}
	})
	
	t.Run("Str2Bool2Int", func(t *testing.T) {
		// Test chaining: string -> bool -> int
		testCases := []struct {
			str         string
			expectedInt int
		}{
			{"true", 1},
			{"false", 0},
			{"anything", 0},
			{"", 0},
		}
		
		for _, tc := range testCases {
			boolVal := Str2Bool(tc.str)
			intVal := Bool2Int(boolVal)
			if intVal != tc.expectedInt {
				t.Errorf("Str2Bool2Int chain failed for \"%s\": -> %t -> %d (expected %d)", tc.str, boolVal, intVal, tc.expectedInt)
			}
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("LongStrings", func(t *testing.T) {
		// Test with very long strings
		longString := ""
		for i := 0; i < 10000; i++ {
			longString += "a"
		}
		
		result := Str2Bool(longString)
		if result != false {
			t.Errorf("Str2Bool with long string should return false, got %t", result)
		}
		
		// Long string that ends with "true"
		longStringWithTrue := longString + "true"
		result2 := Str2Bool(longStringWithTrue)
		if result2 != false {
			t.Errorf("Str2Bool with long string ending in 'true' should return false, got %t", result2)
		}
	})
	
	t.Run("SpecialCharacters", func(t *testing.T) {
		specialChars := []string{
			"\n", "\t", "\r", " ",
			"true\x00", // null byte
			"tr\x00ue", // null byte in middle
		}
		
		for _, char := range specialChars {
			result := Str2Bool(char)
			if result != false {
				t.Errorf("Str2Bool(\"%q\") should return false, got %t", char, result)
			}
		}
	})
}

// Benchmark tests
func BenchmarkBool2Int(b *testing.B) {
	b.Run("True", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Bool2Int(true)
		}
	})
	
	b.Run("False", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Bool2Int(false)
		}
	})
}

func BenchmarkStr2Bool(b *testing.B) {
	b.Run("TrueString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Str2Bool("true")
		}
	})
	
	b.Run("FalseString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Str2Bool("false")
		}
	})
	
	b.Run("EmptyString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Str2Bool("")
		}
	})
	
	b.Run("LongString", func(b *testing.B) {
		longStr := ""
		for i := 0; i < 1000; i++ {
			longStr += "x"
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Str2Bool(longStr)
		}
	})
}

func BenchmarkInt2Bool(b *testing.B) {
	b.Run("Zero", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Int2Bool(0)
		}
	})
	
	b.Run("One", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Int2Bool(1)
		}
	})
	
	b.Run("Negative", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Int2Bool(-1)
		}
	})
	
	b.Run("Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Int2Bool(9223372036854775807)
		}
	})
}

func BenchmarkConversionChaining(b *testing.B) {
	b.Run("Bool2Int2Bool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			val := true
			intVal := Bool2Int(val)
			Int2Bool(int64(intVal))
		}
	})
	
	b.Run("Str2Bool2Int", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			boolVal := Str2Bool("true")
			Bool2Int(boolVal)
		}
	})
}
