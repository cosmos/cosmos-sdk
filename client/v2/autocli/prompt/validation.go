package prompt

import (
	"errors"
	"fmt"
	"net/url"
	"unicode"

	"cosmossdk.io/core/address"
)

// ValidatePromptNotEmpty validates that the input is not empty.
func ValidatePromptNotEmpty(input string) error {
	if input == "" {
		return errors.New("input cannot be empty")
	}

	return nil
}

// ValidateAddress returns a validation function that checks if a string is a valid address
// for the given address codec.
func ValidateAddress(ac address.Codec) func(string) error {
	return func(i string) error {
		if _, err := ac.StringToBytes(i); err != nil {
			return fmt.Errorf("invalid address")
		}

		return nil
	}
}

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
