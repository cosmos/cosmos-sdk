package client

import (
	"fmt"
	"net/url"
	"unicode"
)

// ValidatePromptURL validates that the input is a valid URL.
func ValidatePromptURL(input string) error {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	return nil
}

// CamelCaseToString converts a camel case string to a string with spaces.
func CamelCaseToString(str string) string {
	w := []rune(str)
	for i := len(w) - 1; i > 1; i-- {
		if unicode.IsUpper(w[i]) {
			w = append(w[:i], append([]rune{' '}, w[i:]...)...)
		}
	}
	return string(w)
}
