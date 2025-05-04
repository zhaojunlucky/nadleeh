package js_token

import (
	"reflect"
	"strings"
	"testing"
)

func TestScanner_Scan(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []Token
		expectError bool
		errorSubstr string
	}{
		{
			name:  "Example from requirements",
			input: "tar -czvf nadleeh-PR-${{ github.event.pull_request.number }}-1.0.${{ github.run_number }}.tar.gz nadleeh_amd64 nadleeh_arm64",
			expected: []Token{
				{Type: RawString, Value: "tar -czvf nadleeh-PR-"},
				{Type: VarString, Value: "github.event.pull_request.number"},
				{Type: RawString, Value: "-1.0."},
				{Type: VarString, Value: "github.run_number"},
				{Type: RawString, Value: ".tar.gz nadleeh_amd64 nadleeh_arm64"},
			},
		},
		{
			name:  "No variables",
			input: "simple string with no variables",
			expected: []Token{
				{Type: RawString, Value: "simple string with no variables"},
			},
		},
		{
			name:        "Variable with newline",
			input:       "${{ just.a.variable \n }}",
			expectError: true,
			errorSubstr: "variable contains newline",
		},
		{
			name:  "Only variable",
			input: "${{ just.a.variable }}",
			expected: []Token{
				{Type: VarString, Value: "just.a.variable"},
			},
		},
		{
			name:  "Multiple adjacent variables",
			input: "${{ var1 }}${{ var2 }}",
			expected: []Token{
				{Type: VarString, Value: "var1"},
				{Type: VarString, Value: "var2"},
			},
		},
		{
			name:        "Unclosed variable",
			input:       "Start ${{ unclosed.variable",
			expectError: true,
			errorSubstr: "unclosed variable",
		},
		{
			name:        "Nested variable",
			input:       "Start ${{ outer ${{ inner ${{ inner2 }} }} }}",
			expectError: true,
			errorSubstr: "nested variable",
		},
		{
			name:  "Variable with whitespace in delimiter",
			input: "This is a $ { { invalid.variable } } pattern",
			expected: []Token{
				{Type: RawString, Value: "This is a $ { { invalid.variable } } pattern"},
			},
		},
		{
			name:  "Multiple variables with text in between",
			input: "${{ var1 }} some text ${{ var2 }}",
			expected: []Token{
				{Type: VarString, Value: "var1"},
				{Type: RawString, Value: " some text "},
				{Type: VarString, Value: "var2"},
			},
		},
	}
	scanner := JSTokenScanner{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := scanner.Scan(tt.input)

			// Check error expectations
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tt.errorSubstr)
				} else if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("Expected error containing '%s', but got '%s'", tt.errorSubstr, err.Error())
				}
				return
			}

			// If we're not expecting an error, it should be nil
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(tokens, tt.expected) {
				t.Errorf("Scan() = %v, want %v", tokens, tt.expected)
			}
		})
	}
}
