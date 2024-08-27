package schema

import "regexp"

// NameFormat is the regular expression that a name must match.
// A name must start with a letter or underscore and can only contain letters, numbers, and underscores.
// A name must be at least one character long and can be at most 63 characters long.
const NameFormat = `^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`

var nameRegex = regexp.MustCompile(NameFormat)

// ValidateName checks if the given name is a valid name conforming to NameFormat.
func ValidateName(name string) bool {
	return nameRegex.MatchString(name)
}

// QualifiedNameFormat is the regular expression that a qualified name must match.
// A qualified name is a dot-separated list of names, where each name must match NameFormat.
// A qualified name must be at least one character long and can be at most 127 characters long.
const QualifiedNameFormat = `^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`

var qualifiedNameRegex = regexp.MustCompile(QualifiedNameFormat)

func ValidateQualifiedName(name string) bool {
	if !qualifiedNameRegex.MatchString(name) {
		return false
	}

	if len(name) > 127 {
		return false
	}

	return true
}
