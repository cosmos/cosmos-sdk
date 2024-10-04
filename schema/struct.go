package schema

// StructType represents a struct type.
// Support for this is currently UNIMPLEMENTED, this notice will be removed when it is added.
type StructType struct {
	// Name is the name of the struct type.
	Name string

	// Fields is the list of fields in the struct.
	// It is a COMPATIBLE change to add new fields to an unsealed struct,
	// but it is an INCOMPATIBLE change to add new fields to a sealed struct.
	//
	// A sealed struct cannot reference any unsealed structs directly or
	// transitively because these types allow adding new fields.
	Fields []Field

	// Sealed is true if it is an INCOMPATIBLE change to add new fields to the struct.
	// It is a COMPATIBLE change to change an unsealed struct to sealed, but it is
	// an INCOMPATIBLE change to change a sealed struct to unsealed.
	Sealed bool
}
