package validation

import (
	"regexp"
	"strings"
)

// regular expression to check string is lowercase alphabetic characters only
var isAlphaLower = regexp.MustCompile(`^[a-z]+$`).MatchString

// regular expression to check string is alphanumeric
var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// Validator function type to validate path and identifier bytestrings
type Validator func(string) bool

// Default validator function for Client, Connection, and Channel
// identifiers
// Valid Identifier must be between 10-20 characters and only
// contain lowercase alphabetic characters
func DefaultIdentifierValidator(id string) bool {
	// valid id must be between 10 and 20 characters
	if len(id) < 10 || len(id) > 20 {
		return false
	}
	// valid id must contain only lower alphabetic characters
	if !isAlphaLower(string(id)) {
		return false
	}
	return true
}

// NewPathValidator takes in a Identifier Validator function and returns
// a Path Validator function which requires path only has valid identifiers
// alphanumeric character strings, and "/" separators
func NewPathValidator(idValidator Validator) Validator {
	return func(path string) bool {
		pathArr := strings.Split(path, "/")
		for _, p := range pathArr {
			// Each path element must either be valid identifier or alphanumeric
			if !idValidator(p) && !isAlphaNumeric(p) {
				return false
			}
		}
		return true
	}
}

// Default Path Validator takes in path string and validates
// with default identifier rules. This is optimized by simply
// checking that all path elements are alphanumeric
func DefaultPathValidator(path string) bool {
	pathArr := strings.Split(path, "/")
	for _, p := range pathArr {
		// Each path element must either be alphanumeric
		if !isAlphaNumeric(p) {
			return false
		}
	}
	return true
}
