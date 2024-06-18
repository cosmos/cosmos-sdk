package indexerbase

import "fmt"

// ObjectType describes an object type a module schema.
type ObjectType struct {
	// Name is the name of the object.
	Name string

	// KeyFields is a list of fields that make up the primary key of the object.
	// It can be empty in which case indexers should assume that this object is
	// a singleton and ony has one value.
	KeyFields []Field

	// ValueFields is a list of fields that are not part of the primary key of the object.
	// It can be empty in the case where all fields are part of the primary key.
	ValueFields []Field

	// RetainDeletions is a flag that indicates whether the indexer should retain
	// deleted rows in the database and flag them as deleted rather than actually
	// deleting the row. For many types of data in state, the data is deleted even
	// though it is still valid in order to save space. Indexers will want to have
	// the option of retaining such data and distinguishing from other "true" deletions.
	RetainDeletions bool
}

func (o ObjectType) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("object type name cannot be empty")
	}

	for _, field := range o.KeyFields {
		if err := field.Validate(); err != nil {
			return fmt.Errorf("invalid key field %q: %w", field.Name, err)
		}
	}

	for _, field := range o.ValueFields {
		if err := field.Validate(); err != nil {
			return fmt.Errorf("invalid value field %q: %w", field.Name, err)
		}
	}

	return nil

}

func (o ObjectType) ValidateObjectUpdate(update ObjectUpdate) error {
	if o.Name != update.TypeName {
		return fmt.Errorf("object type name %q does not match update type name %q", o.Name, update.TypeName)
	}

	if err := o.ValidateKey(update.Key); err != nil {
		return fmt.Errorf("invalid key for object type %q: %w", update.TypeName, err)
	}

	if update.Delete {
		return nil
	}

	return o.ValidateValue(update.Value)
}

func (o ObjectType) ValidateKey(value interface{}) error {
	return validateFieldsValue(o.KeyFields, value)
}

// ValidateValue validates that the value conforms to the set of fields as a Value in an EntityUpdate.
// See EntityUpdate.Value for documentation on the requirements of such values.
func (o ObjectType) ValidateValue(value interface{}) error {
	valueUpdates, ok := value.(ValueUpdates)
	if !ok {
		return validateFieldsValue(o.ValueFields, value)
	}

	values := map[string]interface{}{}
	err := valueUpdates.Iterate(func(fieldname string, value interface{}) bool {
		values[fieldname] = value
		return true
	})
	if err != nil {
		return err
	}

	for _, field := range o.ValueFields {
		v, ok := values[field.Name]
		if !ok {
			continue
		}

		if err := field.ValidateValue(v); err != nil {
			return err
		}

		delete(values, field.Name)
	}

	if len(values) > 0 {
		return fmt.Errorf("unexpected fields in ValueUpdates: %v", values)
	}

	return nil
}

func validateFieldsValue(fields []Field, value interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	if len(fields) == 1 {
		return fields[0].ValidateValue(value)
	}

	values, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected slice of values for key fields, got %T", value)
	}

	if len(fields) != len(values) {
		return fmt.Errorf("expected %d key fields, got %d values", len(fields), len(value.([]interface{})))
	}
	for i, field := range fields {
		if err := field.ValidateValue(values[i]); err != nil {
			return err
		}
	}
	return nil
}
