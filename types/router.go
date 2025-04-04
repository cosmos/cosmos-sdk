package types

import (
	"regexp"
)

// IsAlphaNumeric defines a regular expression for matching against alpha-numeric
// values.
var IsAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// IsAlphaNumericOrUnderscore defines a regular expression for matching against
// alpha-numeric or underscore values.
var IsAlphaNumericOrUnderscore = regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString
