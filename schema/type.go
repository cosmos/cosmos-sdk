package schema

// Type is an interface that all types in the schema implement.
// Currently these are ObjectType and EnumType.
type Type interface {
	// TypeName returns the type's name.
	TypeName() string

	isType()
}
