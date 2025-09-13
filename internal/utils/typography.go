package utils

import (
	"regexp"
	"strings"
	"unicode"
)

func ToPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(c rune) bool {
		return c == '_' || c == '-' || c == ' '
	})

	for i, word := range words {
		words[i] = toTitle(word)
	}

	return strings.Join(words, "")
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
