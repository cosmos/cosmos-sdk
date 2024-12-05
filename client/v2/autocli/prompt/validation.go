package prompt

import (
	"errors"
	"fmt"

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
