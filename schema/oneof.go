package schema

// OneOfType represents a oneof type.
type OneOfType struct {
	// Name is the name of the oneof type. It must conform to the NameFormat regular expression.
	Name string

	// Cases is a list of cases in the oneof type.
	// It is a COMPATIBLE change to add new cases to a oneof type.
	// If a newer client tries to send a message with a case that an older server does not recognize,
	// the older server will simply reject it in a switch statement.
	// Existing cases should not be removed or modified.
	Cases []OneOfCase

	// DiscriminantKind is the kind of the discriminant field.
	// It must be Uint8Kind, Int8Kind, Uint16Kind, Int16Kind, or Int32Kind.
	DiscriminantKind Kind
}

// OneOfCase represents a case in a oneof type. It is represented by a struct type internally with a discriminant value.
type OneOfCase struct {
	// StructType represents the name and fields of the case.
	// As with normal structs, it is an INCOMPATIBLE change to add, remove or update fields.
	// If a newer client tries to send a message to an older server and the case has new fields,
	// the message will be incomprehensible.
	StructType

	Name string

	// Discriminant is the discriminant value for the case.
	Discriminant int32
}

type OneOfValue = struct {
	// Case is the name of the case.
	Case string

	// Value is the value of the case.
	Value interface{}
}
