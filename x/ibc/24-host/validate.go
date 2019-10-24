package host

import (
	"regexp"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ICS 024 Identifier and Path Validation Implementation
//
// This file defines ValidateFn to validate identifier and path strings
// The spec for ICS 024 can be located here:
// https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements

// regular expression to check string is lowercase alphabetic characters only
var isAlphaLower = regexp.MustCompile(`^[a-z]+$`).MatchString

// regular expression to check string is alphanumeric
var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// ValidateFn function type to validate path and identifier bytestrings
type ValidateFn func(string) error

// DefaultIdentifierValidator is the default validator function for Client,
// Connection and Channel identifiers.
// A valid Identifier must be between 10-20 characters and only contain lowercase
// alphabetic characters,
func DefaultIdentifierValidator(id string) error {
	// valid id MUST NOT contain "/" separator
	if strings.Contains(id, "/") {
		return sdkerrors.Wrap(ErrInvalidID, "identifier cannot contain separator: /")
	}
	// valid id must be between 10 and 20 characters
	if len(id) < 10 || len(id) > 20 {
		return sdkerrors.Wrapf(ErrInvalidID, "identifier has invalid length: %d, must be between 10-20 characters", len(id))
	}
	// valid id must contain only lower alphabetic characters
	if !isAlphaLower(id) {
		return sdkerrors.Wrap(ErrInvalidID, "identifier must contain only lowercase alphabetic characters")
	}
	return nil
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
			if err != nil && !isAlphaNumeric(p) {
				return sdkerrors.Wrapf(ErrInvalidPath, "path contains invalid identifier or non-alphanumeric path element: %s", p)
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
	for _, p := range pathArr {
		// Each path element must either be alphanumeric
		if !isAlphaNumeric(p) {
			return sdkerrors.Wrapf(ErrInvalidPath, "invalid path element containing non-alphanumeric characters: %s", p)
		}
	}
	return nil
}
