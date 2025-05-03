package js_token

import (
	"fmt"
	"strings"
)

// TokenType represents the type of token
type TokenType int

const (
	// RawString represents a raw string token
	RawString TokenType = iota
	// VarString represents a variable string token (surrounded by ${{ }})
	VarString
)

// Token represents a token in the input string
type Token struct {
	Type  TokenType
	Value string
}

// JSTokenScanner scans a string and extracts tokens
type JSTokenScanner struct {
	input string
}

// Scan scans the input string and returns a list of tokens
func (s *JSTokenScanner) Scan(input string) ([]Token, error) {
	var tokens []Token

	// Read character by character to validate variable syntax
	var i int
	var currentRaw strings.Builder
	inVariable := false
	varStartPos := -1

	// Helper function to check if a character is whitespace
	isWhitespace := func(c byte) bool {
		return c == ' ' || c == '\t' || c == '\n' || c == '\r'
	}

	// Helper function to check if we have a valid variable start
	isValidVarStart := func(pos int) bool {
		// Need at least "${{" (3 chars)
		if pos+2 >= len(input) {
			return false
		}

		// Must be exactly "${{" without whitespace
		return input[pos] == '$' && input[pos+1] == '{' && input[pos+2] == '{'
	}

	for i < len(input) {
		// Check for variable start sequence "${{" (must be exactly this, no whitespace)
		if !inVariable && isValidVarStart(i) {
			// If we have accumulated raw string, add it as a token
			if currentRaw.Len() > 0 {
				tokens = append(tokens, Token{Type: RawString, Value: currentRaw.String()})
				currentRaw.Reset()
			}

			inVariable = true
			varStartPos = i
			i += 3 // Skip past "${{"
			continue
		}

		// Check for invalid variable start with whitespace (like "$ { {")
		if !inVariable && i+4 < len(input) && input[i] == '$' &&
			(isWhitespace(input[i+1]) || isWhitespace(input[i+3])) {
			// Look ahead to see if this might be an attempt at a variable with whitespace
			j := i + 1
			for j < len(input) && isWhitespace(input[j]) {
				j++
			}
			if j < len(input) && input[j] == '{' {
				k := j + 1
				for k < len(input) && isWhitespace(input[k]) {
					k++
				}
				if k < len(input) && input[k] == '{' {
					// This is an invalid variable pattern with whitespace
					// Just treat it as raw text and continue
					currentRaw.WriteByte(input[i])
					i++
					continue
				}
			}
		}

		// Check for nested variable start - this is invalid
		if inVariable && isValidVarStart(i) {
			// Return error for nested variable
			return nil, fmt.Errorf("nested variable at position %d: %s", i, input[varStartPos:i+3]+"...")
		}

		// Check for variable end sequence "}}" (must be exactly this, no whitespace)
		if inVariable && i+1 < len(input) && input[i] == '}' && input[i+1] == '}' {
			// Extract the variable name (without ${{ and }})
			varContent := input[varStartPos+3 : i]

			// Check if the variable content contains newlines
			if strings.Contains(varContent, "\n") {
				return nil, fmt.Errorf("variable contains newline at position %d: %s", varStartPos, input[varStartPos:i+2])
			}

			varName := strings.TrimSpace(varContent)
			tokens = append(tokens, Token{Type: VarString, Value: varName})

			inVariable = false
			i += 2 // Skip past "}}"
			continue
		}

		// Add to current raw string if not in variable
		if !inVariable {
			currentRaw.WriteByte(input[i])
		}

		i++
	}

	// Handle any unclosed variable (return error)
	if inVariable {
		return nil, fmt.Errorf("unclosed variable starting at position %d: %s", varStartPos, input[varStartPos:])
	}

	// Add any remaining raw content
	if currentRaw.Len() > 0 {
		tokens = append(tokens, Token{Type: RawString, Value: currentRaw.String()})
	}

	// If no tokens were created, return the entire input as raw
	if len(tokens) == 0 {
		return []Token{{Type: RawString, Value: input}}, nil
	}

	return tokens, nil
}
