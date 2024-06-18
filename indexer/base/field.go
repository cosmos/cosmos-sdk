package indexerbase

// Field represents a field in an object type.
type Field struct {
	// Name is the name of the field.
	Name string

	// Kind is the basic type of the field.
	Kind Kind

	// Nullable indicates whether null values are accepted for the field.
	Nullable bool

	// AddressPrefix is the address prefix of the field's kind, currently only used for Bech32AddressKind.
	AddressPrefix string

	// EnumDefinition is the definition of the enum type and is only valid when Kind is EnumKind.
	EnumDefinition EnumDefinition
}
