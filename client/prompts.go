package client

import (
	"fmt"
	"net/url"
	"unicode"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/manifoldco/promptui"
)

// Prompts

var (
	AddressPrompt = promptui.Prompt{
		Label: "Enter address",
		Validate: func(input string) error {
			if _, err := sdk.AccAddressFromBech32(input); err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}

			return nil
		},
	}

	CoinsAmountPrompt = promptui.Prompt{
		Label: "Enter coin(s) amount",
		Validate: func(input string) error {
			_, err := sdk.ParseCoinsNormalized(input)
			if err != nil {
				return fmt.Errorf("invalid coin amount: %w", err)
			}

			return nil
		},
	}
)

// Validation

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
