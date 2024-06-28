package schema

import "fmt"

// ObjectType describes an object type a module schema.
type ObjectType struct {
	// Name is the name of the object type. It must be unique within the module schema
	// and conform to the NameFormat regular expression.
	Name string

	// KeyFields is a list of fields that make up the primary key of the object.
	// It can be empty in which case indexers should assume that this object is
	// a singleton and only has one value. Field names must be unique within the
	// object between both key and value fields.
	KeyFields []Field

	// ValueFields is a list of fields that are not part of the primary key of the object.
	// It can be empty in the case where all fields are part of the primary key.
	// Field names must be unique within the object between both key and value fields.
	ValueFields []Field

	// RetainDeletions is a flag that indicates whether the indexer should retain
	// deleted rows in the database and flag them as deleted rather than actually
	// deleting the row. For many types of data in state, the data is deleted even
	// though it is still valid in order to save space. Indexers will want to have
	// the option of retaining such data and distinguishing from other "true" deletions.
	RetainDeletions bool
}

// Validate validates the object type.
func (o ObjectType) Validate() error {
	if !ValidateName(o.Name) {
		return fmt.Errorf("invalid object type name %q", o.Name)
	}

	fieldNames := map[string]bool{}
	for _, field := range o.KeyFields {
		if err := field.Validate(); err != nil {
			return fmt.Errorf("invalid key field %q: %w", field.Name, err)
		}

		if fieldNames[field.Name] {
			return fmt.Errorf("duplicate field name %q", field.Name)
		}
		fieldNames[field.Name] = true
	}

	for _, field := range o.ValueFields {
		if err := field.Validate(); err != nil {
			return fmt.Errorf("invalid value field %q: %w", field.Name, err)
		}

		if fieldNames[field.Name] {
			return fmt.Errorf("duplicate field name %q", field.Name)
		}
		fieldNames[field.Name] = true
	}

	if len(o.KeyFields) == 0 && len(o.ValueFields) == 0 {
		return fmt.Errorf("object type %q has no key or value fields", o.Name)
	}

	return nil
}

// ValidateObjectUpdate validates that the update conforms to the object type.
func (o ObjectType) ValidateObjectUpdate(update ObjectUpdate) error {
	if o.Name != update.TypeName {
		return fmt.Errorf("object type name %q does not match update type name %q", o.Name, update.TypeName)
	}

	if err := ValidateForKeyFields(o.KeyFields, update.Key); err != nil {
		return fmt.Errorf("invalid key for object type %q: %w", update.TypeName, err)
	}

	if update.Delete {
		return nil
	}

	return ValidateForValueFields(o.ValueFields, update.Value)
}
