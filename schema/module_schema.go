package schema

import (
	"fmt"
	"sort"
)

// ModuleSchema represents the logical schema of a module for purposes of indexing and querying.
type ModuleSchema struct {
	types map[string]Type
}

// CompileModuleSchema compiles the types into a ModuleSchema and validates it.
// Any module schema returned without an error is guaranteed to be valid.
func CompileModuleSchema(types ...Type) (ModuleSchema, error) {
	typeMap := map[string]Type{}

	for _, typ := range types {
		if _, ok := typeMap[typ.TypeName()]; ok {
			return ModuleSchema{}, fmt.Errorf("duplicate type %q", typ.TypeName())
		}

		typeMap[typ.TypeName()] = typ
	}

	res := ModuleSchema{types: typeMap}

	err := res.Validate()
	if err != nil {
		return ModuleSchema{}, err
	}

	return res, nil
}

// MustCompileModuleSchema constructs a new ModuleSchema and panics if it is invalid.
// This should only be used in test code or static initialization where it is safe to panic!
func MustCompileModuleSchema(types ...Type) ModuleSchema {
	sch, err := CompileModuleSchema(types...)
	if err != nil {
		panic(err)
	}
	return sch
}

// Validate validates the module schema.
func (s ModuleSchema) Validate() error {
	for _, typ := range s.types {
		err := typ.Validate(s)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateObjectUpdate validates that the update conforms to the module schema.
func (s ModuleSchema) ValidateObjectUpdate(update ObjectUpdate) error {
	typ, ok := s.types[update.TypeName]
	if !ok {
		return fmt.Errorf("object type %q not found in module schema", update.TypeName)
	}

	objTyp, ok := typ.(ObjectType)
	if !ok {
		return fmt.Errorf("type %q is not an object type", update.TypeName)
	}

	return objTyp.ValidateObjectUpdate(update, s)
}

// LookupType looks up a type by name in the module schema.
func (s ModuleSchema) LookupType(name string) (Type, bool) {
	typ, ok := s.types[name]
	return typ, ok
}

// Types calls the provided function for each type in the module schema and stops if the function returns false.
// The types are iterated over in sorted order by name. This function is compatible with go 1.23 iterators.
func (s ModuleSchema) Types(f func(Type) bool) {
	keys := make([]string, 0, len(s.types))
	for k := range s.types {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !f(s.types[k]) {
			break
		}
	}
}

// ObjectTypes iterators over all the object types in the schema in alphabetical order.
func (s ModuleSchema) ObjectTypes(f func(ObjectType) bool) {
	s.Types(func(t Type) bool {
		objTyp, ok := t.(ObjectType)
		if ok {
			return f(objTyp)
		}
		return true
	})
}

// EnumTypes iterators over all the enum types in the schema in alphabetical order.
func (s ModuleSchema) EnumTypes(f func(EnumType) bool) {
	s.Types(func(t Type) bool {
		enumType, ok := t.(EnumType)
		if ok {
			return f(enumType)
		}
		return true
	})
}

func (ModuleSchema) isTypeSet() {}

var _ TypeSet = ModuleSchema{}
