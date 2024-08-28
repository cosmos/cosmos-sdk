package schema

import "fmt"

// ValidateObjectKey validates that the value conforms to the set of fields as a Key in an ObjectUpdate.
// See ObjectUpdate.Key for documentation on the requirements of such keys.
func ValidateObjectKey(keyFields []Field, value interface{}, schema Schema) error {
	return validateFieldsValue(keyFields, value, schema)
}

// ValidateObjectValue validates that the value conforms to the set of fields as a Value in an ObjectUpdate.
// See ObjectUpdate.Value for documentation on the requirements of such values.
func ValidateObjectValue(valueFields []Field, value interface{}, schema Schema) error {
	valueUpdates, ok := value.(ValueUpdates)
	if !ok {
		return validateFieldsValue(valueFields, value, schema)
	}

	values := map[string]interface{}{}
	err := valueUpdates.Iterate(func(fieldname string, value interface{}) bool {
		values[fieldname] = value
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

		if err := field.ValidateValue(v, schema); err != nil {
			return err
		}

		delete(values, field.Name)
	}

	if len(values) > 0 {
		return fmt.Errorf("unexpected values in ValueUpdates: %v", values)
	}

	return nil
}

func validateFieldsValue(fields []Field, value interface{}, schema Schema) error {
	if len(fields) == 0 {
		return nil
	}

	if len(fields) == 1 {
		return fields[0].ValidateValue(value, schema)
	}

	values, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected slice of values for key fields, got %T", value)
	}

	if len(fields) != len(values) {
		return fmt.Errorf("expected %d key fields, got %d values", len(fields), len(value.([]interface{})))
	}
	for i, field := range fields {
		if err := field.ValidateValue(values[i], schema); err != nil {
			return err
		}
	}
	return nil
}
