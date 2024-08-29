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

// Schema represents something that has types and allows them to be looked up by name.
// Currently, the only implementation is ModuleSchema.
type Schema interface {
	// LookupType looks up a type by name.
	LookupType(name string) (Type, bool)

	// Types calls the given function for each type in the schema.
	Types(f func(Type) bool)
}

// EmptySchema is a schema that contains no types.
// It can be used in Validate methods when there is no schema needed or available.
type EmptySchema struct{}

// LookupType always returns false because there are no types in an EmptySchema.
func (EmptySchema) LookupType(name string) (Type, bool) {
	return nil, false
}

// Types does nothing because there are no types in an EmptySchema.
func (EmptySchema) Types(f func(Type) bool) {}
