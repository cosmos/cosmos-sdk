package schema

import "fmt"

// ModuleSchema represents the logical schema of a module for purposes of indexing and querying.
type ModuleSchema struct {
	// ObjectTypes describe the types of objects that are part of the module's schema.
	ObjectTypes []ObjectType
}

// Validate validates the module schema.
func (s ModuleSchema) Validate() error {
	for _, objType := range s.ObjectTypes {
		if err := objType.Validate(); err != nil {
			return err
		}
	}

	// validate that shared enum types are consistent across object types
	enumValueMap := map[string]map[string]bool{}
	for _, objType := range s.ObjectTypes {
		for _, field := range objType.KeyFields {
			err := checkEnum(enumValueMap, field)
			if err != nil {
				return err
			}
		}
		for _, field := range objType.ValueFields {
			err := checkEnum(enumValueMap, field)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func checkEnum(enumValueMap map[string]map[string]bool, field Field) error {
	if field.Kind != EnumKind {
		return nil
	}

	enum := field.EnumDefinition

	if existing, ok := enumValueMap[enum.Name]; ok {
		if len(existing) != len(enum.Values) {
			return fmt.Errorf("enum %q has different number of values in different object types", enum.Name)
		}

		for _, value := range enum.Values {
			if !existing[value] {
				return fmt.Errorf("enum %q has different values in different object types", enum.Name)
			}
		}
	} else {
		valueMap := map[string]bool{}
		for _, value := range enum.Values {
			valueMap[value] = true
		}
		enumValueMap[enum.Name] = valueMap
	}
	return nil
}

// ValidateObjectUpdate validates that the update conforms to the module schema.
func (s ModuleSchema) ValidateObjectUpdate(update ObjectUpdate) error {
	for _, objType := range s.ObjectTypes {
		if objType.Name == update.TypeName {
			return objType.ValidateObjectUpdate(update)
		}
	}
	return fmt.Errorf("object type %q not found in module schema", update.TypeName)
}
