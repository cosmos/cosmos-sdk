package math

// DecFromLegacyDec converts a LegacyDec to the Dec type using a string intermediate representation.
//
// This function can be used when migrating LegacyDec types to the Dec type.
func DecFromLegacyDec(legacyDec LegacyDec) (Dec, error) {
	return NewDecFromString(legacyDec.String())
}
