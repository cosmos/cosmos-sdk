package schema

// OneOfType represents a oneof type.
type OneOfType struct {
	// Name is the name of the oneof type. It must conform to the NameFormat regular expression.
	Name string

	// Cases is a list of cases in the oneof type.
	// It is a COMPATIBLE change to add new cases to a oneof type, but existing cases should not be removed or modified.
	Cases []OneOfCase
}

// OneOfCase represents a case in a oneof type. It is represented by a struct type internally with a discriminant value.
type OneOfCase struct {
	// StructType represents the name and fields of the case.
	// As with normal structs, it is an INCOMPATIBLE change to add, remove or update fields.
	StructType

	// Discriminant is the discriminant value for the case.
	Discriminant int32
}
