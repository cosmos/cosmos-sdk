package host

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ICS 024 Identifier and Path Validation Implementation
//
// This file defines ValidateFn to validate identifier and path strings
// The spec for ICS 024 can be located here:
// https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements

// ValidateFn function type to validate path and identifier bytestrings
type ValidateFn func(string) error

func defaultIdentifierValidator(id string, min, max int) error {
	// valid id MUST NOT contain "/" separator
	if strings.Contains(id, "/") {
		return sdkerrors.Wrapf(ErrInvalidID, "identifier %s cannot contain separator '/'", id)
	}
	// valid id must be between 10 and 20 characters
	if len(id) < min || len(id) > max {
		return sdkerrors.Wrapf(ErrInvalidID, "identifier %s has invalid length: %d, must be between %d-%d characters", id, len(id), min, max)
	}
	// valid id must contain only lower alphabetic characters
	if !sdk.IsAlphaLower(id) {
		return sdkerrors.Wrapf(ErrInvalidID, "identifier %s must contain only lowercase alphabetic characters", id)
	}
	return nil
}

// DefaultClientIdentifierValidator is the default validator function for Client identifiers
// A valid Identifier must be between 10-20 characters and only contain lowercase
// alphabetic characters,
func DefaultClientIdentifierValidator(id string) error {
	return defaultIdentifierValidator(id, 10, 20)
}

// DefaultConnectionIdentifierValidator is the default validator function for Connection identifiers
// A valid Identifier must be between 10-20 characters and only contain lowercase
// alphabetic characters,
func DefaultConnectionIdentifierValidator(id string) error {
	return defaultIdentifierValidator(id, 10, 20)
}

// DefaultChannelIdentifierValidator is the default validator function for Channel identifiers
// A valid Identifier must be between 10-20 characters and only contain lowercase
// alphabetic characters,
func DefaultChannelIdentifierValidator(id string) error {
	return defaultIdentifierValidator(id, 10, 20)
}

// DefaultPortIdentifierValidator is the default validator function for Port identifiers
// A valid Identifier must be between 2-20 characters and only contain lowercase
// alphabetic characters,
func DefaultPortIdentifierValidator(id string) error {
	return defaultIdentifierValidator(id, 2, 20)
}

// NewPathValidator takes in a Identifier Validator function and returns
// a Path Validator function which requires path only has valid identifiers
// alphanumeric character strings, and "/" separators
func NewPathValidator(idValidator ValidateFn) ValidateFn {
	return func(path string) error {
		pathArr := strings.Split(path, "/")
		for _, p := range pathArr {
			// Each path element must either be valid identifier or alphanumeric
			err := idValidator(p)
			if err != nil && !sdk.IsAlphaNumeric(p) {
				return sdkerrors.Wrapf(ErrInvalidPath, "path %s contains invalid identifier or non-alphanumeric path element: %s", path, p)
			}
		}
		return nil
	}
}

// DefaultPathValidator takes in path string and validates
// with default identifier rules. This is optimized by simply
// checking that all path elements are alphanumeric
func DefaultPathValidator(path string) error {
	pathArr := strings.Split(path, "/")
	if pathArr[0] == path {
		return sdkerrors.Wrapf(ErrInvalidPath, "path %s doesn't contain any separator '/'", path)
	}

	for _, p := range pathArr {
		// Each path element must be alphanumeric and non-blank
		if strings.TrimSpace(p) == "" || !sdk.IsAlphaNumeric(p) {
			return sdkerrors.Wrapf(ErrInvalidPath, "path %s contains an invalid non-alphanumeric character: '%s'", path, p)
		}
	}
	return nil
}
