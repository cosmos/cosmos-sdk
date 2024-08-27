package schema

// Type is an interface that all types in the schema implement.
// Currently, these are ObjectType and EnumType.
type Type interface {
	// TypeName returns the type's name.
	TypeName() string

	// Validate validates the type.
	Validate(Schema) error

	// isType is a private method that ensures that only types in this package can be marked as types.
	isType()
}

type Schema interface {
	LookupType(name string) (Type, bool)
	Types(f func(Type) bool)
}

// EmptySchema is a schema that contains no types.
// It can be used in Validate methods when there is no schema needed or available.
type EmptySchema struct{}

func (EmptySchema) LookupType(name string) (Type, bool) {
	return nil, false
}

func (EmptySchema) Types(f func(Type) bool) {}
