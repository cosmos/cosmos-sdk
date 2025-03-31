package math

// DecFromLegacyDec converts a LegacyDec to the Dec type using a string intermediate representation.
func DecFromLegacyDec(legacyDec LegacyDec) (Dec, error) {
	return NewDecFromString(legacyDec.String())
}
