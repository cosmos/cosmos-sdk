package indexerbase

import "fmt"

// ModuleSchema represents the logical schema of a module for purposes of indexing and querying.
type ModuleSchema struct {
	// ObjectTypes describe the types of objects that are part of the module's schema.
	ObjectTypes map[string]ObjectType

	// EnumTypes describe the enum types that are part of the module's schema.
	EnumTypes map[string]EnumDefinition
}

func (m ModuleSchema) ValidateObjectUpdate(update ObjectUpdate) error {
	if err := m.ValidateObjectKey(update.TypeName, update.Key); err != nil {
		return err
	}

	if update.Delete {
		return nil
	}

	return m.ValidateObjectValue(update.TypeName, update.Value)
}

func (m ModuleSchema) ValidateObjectKey(objectType string, value interface{}) error {
	return m.validateFieldsValue(m.ObjectTypes[objectType].KeyFields, value)
}

func (m ModuleSchema) ValidateObjectValue(objectType string, value interface{}) error {
	valueFields := m.ObjectTypes[objectType].ValueFields

	valueUpdates, ok := value.(ValueUpdates)
	if !ok {
		return m.validateFieldsValue(valueFields, value)
	}

	values := map[string]interface{}{}
	err := valueUpdates.Iterate(func(fieldName string, value interface{}) bool {
		values[fieldName] = value
		return true
	})
	if err != nil {
		return err
	}

	for _, field := range valueFields {
		v, ok := values[field.Name]
		if !ok {
			continue
		}

		if err := m.validateFieldValue(field, v); err != nil {
			return err
		}

		delete(values, field.Name)
	}

	if len(values) > 0 {
		return fmt.Errorf("unexpected fields in ValueUpdates: %v", values)
	}

	return nil
}

func (m ModuleSchema) validateFieldValue(field Field, value interface{}) error {
	if err := field.ValidateValue(value); err != nil {
		return fmt.Errorf("invalid value for key field %q: %w", field.Name, err)
	}

	if field.Kind == EnumKind {
		enumType, ok := m.EnumTypes[field.EnumType]
		if !ok {
			return fmt.Errorf("unknown enum type %q for field %q", field.EnumType, field.Name)
		}
		if err := enumType.ValidateValue(value.(string)); err != nil {
			return err
		}
	}

	return nil
}

func (m ModuleSchema) validateFieldsValue(fields []Field, value interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	if len(fields) == 1 {
		return m.validateFieldValue(fields[0], value)
	}

	values, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected slice of values for key fields, got %T", value)
	}

	if len(fields) != len(values) {
		return fmt.Errorf("expected %d key fields, got %d values", len(fields), len(value.([]interface{})))
	}
	for i, field := range fields {
		if err := m.validateFieldValue(field, values[i]); err != nil {
			return err
		}
	}
	return nil
}
