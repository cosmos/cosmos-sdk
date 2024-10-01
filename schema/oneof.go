package schema

// OneOfType represents a oneof type.
// Support for this is currently UNIMPLEMENTED, this notice will be removed when it is added.
type OneOfType struct {
	// Name is the name of the oneof type. It must conform to the NameFormat regular expression.
	Name string

	// Cases is a list of cases in the oneof type.
	// It is a COMPATIBLE change to add new cases to a oneof type.
	// If a newer client tries to send a message with a case that an older server does not recognize,
	// the older server will simply reject it in a switch statement.
	// It is INCOMPATIBLE to remove existing cases from a oneof type.
	Cases []OneOfCase

	// DiscriminantKind is the kind of the discriminant field.
	// It must be Uint8Kind, Int8Kind, Uint16Kind, Int16Kind, or Int32Kind.
	DiscriminantKind Kind
}

// OneOfCase represents a case in a oneof type. It is represented by a struct type internally with a discriminant value.
type OneOfCase struct {
	// Name is the name of the case. It must conform to the NameFormat regular expression.
	Name string

	// Discriminant is the discriminant value for the case.
	Discriminant int32

	// Kind is the kind of the case. ListKind is not allowed.
	Kind Kind

	// Reference is the referenced type if Kind is EnumKind, StructKind, or OneOfKind.
	ReferencedType string
}

// OneOfValue is the golang runtime representation of a oneof value.
type OneOfValue = struct {
	// Case is the name of the case.
	Case string

	// Value is the value of the case.
	Value interface{}
}
