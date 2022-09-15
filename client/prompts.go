package client

import (
	"fmt"
	"net/url"
	"unicode"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Prompts Validation

func ValidatePromptNotEmpty(input string) error {
	if input == "" {
		return fmt.Errorf("input cannot be empty")
	}

	return nil
}

func ValidatePromptURL(input string) error {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	return nil
}

func ValidatePromptAddress(input string) error {
	if _, err := sdk.AccAddressFromBech32(input); err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	return nil
}

func ValidatePromptCoins(input string) error {
	if _, err := sdk.ParseCoinsNormalized(input); err != nil {
		return fmt.Errorf("invalid coins: %w", err)
	}

	return nil
}

// Prompts Helpers

func CamelCaseToString(str string) string {
	w := []rune(str)
	for i := len(w) - 1; i > 1; i-- {
		if unicode.IsUpper(w[i]) {
			w = append(w[:i], append([]rune{' '}, w[i:]...)...)
		}
	}
	return string(w)
}
