package utils

import (
	"regexp"
	"strings"
	"unicode"
)

func ToPascalCase(s string) string {
	// If input is already in valid PascalCase, return as-is
	if isValidPascalCase(s) {
		return s
	}

	// Split on delimiters and camelCase boundaries
	words := splitWords(s)

	for i, word := range words {
		words[i] = toTitle(word)
	}

	return strings.Join(words, "")
}

// isValidPascalCase checks if string is already in proper PascalCase
func isValidPascalCase(s string) bool {
	if s == "" {
		return false
	}

	// Must start with uppercase
	if !unicode.IsUpper(rune(s[0])) {
		return false
	}

	// Check for valid characters (letters and numbers only)
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}

	return true
}

// splitWords splits a string into words based on delimiters and camelCase boundaries
func splitWords(s string) []string {
	var words []string
	var currentWord strings.Builder

	runes := []rune(s)
	for i, r := range runes {
		// Split on delimiters
		if r == '_' || r == '-' || r == ' ' {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
			continue
		}

		// Split on camelCase boundaries (lowercase followed by uppercase)
		if i > 0 && unicode.IsLower(runes[i-1]) && unicode.IsUpper(r) {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}

		currentWord.WriteRune(r)
	}

	// Add the last word
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if pascal == "" {
		return pascal
	}
	r := []rune(pascal)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func ToKebabCase(s string) string {
	return strings.ReplaceAll(ToSnakeCase(s), "_", "-")
}

func ToTitle(s string) string {
	return toTitle(s)
}

// ToGoIdentifier converts a string to a valid Go identifier
func ToGoIdentifier(s string) string {
	// Remove invalid characters and convert to camelCase
	reg := regexp.MustCompile("[^a-zA-Z0-9_]")
	cleaned := reg.ReplaceAllString(s, "")

	// Ensure it starts with a letter or underscore
	if cleaned == "" {
		return "pkg"
	}

	if unicode.IsDigit(rune(cleaned[0])) {
		cleaned = "_" + cleaned
	}

	return ToCamelCase(cleaned)
}

// toTitle converts the first character to uppercase and the rest to lowercase
// This replaces the deprecated strings.Title function
func toTitle(s string) string {
	if s == "" {
		return s
	}
	r := []rune(strings.ToLower(s))
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
