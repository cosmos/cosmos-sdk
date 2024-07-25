package schema

import (
	"fmt"
	"sort"
)

// ModuleSchema represents the logical schema of a module for purposes of indexing and querying.
type ModuleSchema struct {
	types map[string]Type
}

// NewModuleSchema constructs a new ModuleSchema and validates it. Any module schema returned without an error
// is guaranteed to be valid.
func NewModuleSchema(objectTypes []ObjectType) (ModuleSchema, error) {
	types := map[string]Type{}

	for _, objectType := range objectTypes {
		types[objectType.Name] = objectType
	}

	res := ModuleSchema{types: types}

	// validate adds all enum types to the type map
	err := res.Validate()
	if err != nil {
		return ModuleSchema{}, err
	}

	return res, nil
}

func addEnumType(types map[string]Type, field Field) error {
	enumDef := field.EnumType
	if enumDef.Name == "" {
		return nil
	}

	existing, ok := types[enumDef.Name]
	if !ok {
		types[enumDef.Name] = enumDef
		return nil
	}

	existingEnum, ok := existing.(EnumType)
	if !ok {
		return fmt.Errorf("enum %q already exists as a different non-enum type", enumDef.Name)
	}

	if len(existingEnum.Values) != len(enumDef.Values) {
		return fmt.Errorf("enum %q has different number of values in different fields", enumDef.Name)
	}

	existingValues := map[string]bool{}
	for _, value := range existingEnum.Values {
		existingValues[value] = true
	}

	for _, value := range enumDef.Values {
		_, ok := existingValues[value]
		if !ok {
			return fmt.Errorf("enum %q has different values in different fields", enumDef.Name)
		}
	}

	return nil
}

// Validate validates the module schema.
func (s ModuleSchema) Validate() error {
	for _, typ := range s.types {
		objTyp, ok := typ.(ObjectType)
		if !ok {
			continue
		}

		// all enum types get added to the type map when we call ObjectType.validate
		err := objTyp.validate(s.types)
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

	return objTyp.ValidateObjectUpdate(update)
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
