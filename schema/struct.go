package schema

// StructType represents a struct type.
type StructType struct {
	// Name is the name of the struct type.
	Name string

	// Fields is the list of fields in the struct. ObjectKind fields are not allowed.
	// It is an INCOMPATIBLE change to add, remove or update fields in a struct.
	Fields []Field
}
