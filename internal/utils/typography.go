package utils

import (
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
