package types

import (
	"regexp"
)

var (
	// IsAlphaNumeric defines a regular expression for matching against alpha-numeric
	// values.
	IsAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

	// IsAlphaLower defines regular expression to check if the string has lowercase
	// alphabetic characters only.
	IsAlphaLower = regexp.MustCompile(`^[a-z]+$`).MatchString

	// IsAlphaUpper defines regular expression to check if the string has uppercase
	// alphabetic characters only.
	IsAlphaUpper = regexp.MustCompile(`^[A-Z]+$`).MatchString

	// IsAlpha defines regular expression to check if the string has alphabetic
	// characters only.
	IsAlpha = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

	// IsNumeric defines regular expression to check if the string has numeric
	// characters only.
	IsNumeric = regexp.MustCompile(`^[0-9]+$`).MatchString
)
