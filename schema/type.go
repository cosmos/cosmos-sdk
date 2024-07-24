package schema

// Type is an interface that all types in the schema implement.
// Currently these are ObjectType and EnumDefinition.
type Type interface {
	isType()
}
