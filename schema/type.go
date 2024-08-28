package schema

// Type is an interface that all types in the schema implement.
// Currently, these are ObjectType and EnumType.
type Type interface {
	// TypeName returns the type's name.
	TypeName() string

	// Validate validates the type.
	Validate(TypeSet) error

	// isType is a private method that ensures that only types in this package can be marked as types.
	isType()
}

// TypeSet represents something that has types and allows them to be looked up by name.
// Currently, the only implementation is ModuleSchema.
type TypeSet interface {
	// LookupType looks up a type by name.
	LookupType(name string) (Type, bool)

	// Types calls the given function for each type in the schema.
	Types(f func(Type) bool)

	// isTypeSet is a private method that ensures that only types in this package can be marked as type sets.
	isTypeSet()
}

// EmptyTypeSet is a schema that contains no types.
// It can be used in Validate methods when there is no schema needed or available.
func EmptyTypeSet() TypeSet {
	return emptyTypeSetInst
}

var emptyTypeSetInst = emptyTypeSet{}

type emptyTypeSet struct{}

// LookupType always returns false because there are no types in an EmptyTypeSet.
func (emptyTypeSet) LookupType(string) (Type, bool) {
	return nil, false
}

// Types does nothing because there are no types in an EmptyTypeSet.
func (emptyTypeSet) Types(func(Type) bool) {}

func (emptyTypeSet) isTypeSet() {}
