package indexerbase

// EnumDefinition represents the definition of an enum type.
type EnumDefinition struct {
	// Name is the name of the enum type.
	Name string

	// Values is a list of distinct values that are part of the enum type.
	Values []string
}
