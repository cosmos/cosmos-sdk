package indexerbase

import "fmt"

// Field represents a field in an object type.
type Field struct {
	// Name is the name of the field.
	Name string

	// Kind is the basic type of the field.
	Kind Kind

	// Nullable indicates whether null values are accepted for the field.
	Nullable bool

	// AddressPrefix is the address prefix of the field's kind, currently only used for Bech32AddressKind.
	AddressPrefix string

	// EnumDefinition is the definition of the enum type and is only valid when Kind is EnumKind.
	EnumDefinition EnumDefinition
}

// Validate validates the field.
func (c Field) Validate() error {
	// non-empty name
	if c.Name == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	// valid kind
	if err := c.Kind.Validate(); err != nil {
		return fmt.Errorf("invalid field type for %q: %w", c.Name, err)
	}

	// address prefix only valid with Bech32AddressKind
	if c.Kind == Bech32AddressKind && c.AddressPrefix == "" {
		return fmt.Errorf("missing address prefix for field %q", c.Name)
	} else if c.Kind != Bech32AddressKind && c.AddressPrefix != "" {
		return fmt.Errorf("address prefix is only valid for field %q with type Bech32AddressKind", c.Name)
	}

	// enum definition only valid with EnumKind
	if c.Kind == EnumKind {
		if err := c.EnumDefinition.Validate(); err != nil {
			return fmt.Errorf("invalid enum definition for field %q: %w", c.Name, err)
		}
	} else if c.Kind != EnumKind && c.EnumDefinition.Name != "" && c.EnumDefinition.Values != nil {
		return fmt.Errorf("enum definition is only valid for field %q with type EnumKind", c.Name)
	}

	return nil
}

// Validate validates the enum definition.
func (e EnumDefinition) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("enum definition name cannot be empty")
	}
	if len(e.Values) == 0 {
		return fmt.Errorf("enum definition values cannot be empty")
	}
	seen := make(map[string]bool, len(e.Values))
	for i, v := range e.Values {
		if v == "" {
			return fmt.Errorf("enum definition value at index %d cannot be empty for enum %s", i, e.Name)
		}
		if seen[v] {
			return fmt.Errorf("duplicate enum definition value %q for enum %s", v, e.Name)
		}
		seen[v] = true
	}
	return nil
}

// ValidateValue validates that the value conforms to the field's kind and nullability.
// It currently does not do any validation that IntegerKind, DecimalKind, Bech32AddressKind, or EnumKind
// values are valid for their respective types behind conforming to the correct go type.
func (c Field) ValidateValue(value interface{}) error {
	if value == nil {
		if !c.Nullable {
			return fmt.Errorf("field %q cannot be null", c.Name)
		}
		return nil
	}
	return c.Kind.ValidateValueType(value)
}

// ValidateKey validates that the value conforms to the set of fields as a Key in an EntityUpdate.
// See EntityUpdate.Key for documentation on the requirements of such values.
func ValidateKey(fields []Field, value interface{}) error {
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
			return fmt.Errorf("invalid value for key field %q: %w", field.Name, err)
		}
	}
	return nil
}

// ValidateValue validates that the value conforms to the set of fields as a Value in an EntityUpdate.
// See EntityUpdate.Value for documentation on the requirements of such values.
func ValidateValue(fields []Field, value interface{}) error {
	valueUpdates, ok := value.(ValueUpdates)
	if ok {
		fieldMap := map[string]Field{}
		for _, field := range fields {
			fieldMap[field.Name] = field
		}
		var errs []error
		valueUpdates.Iterate(func(fieldName string, value interface{}) bool {
			field, ok := fieldMap[fieldName]
			if !ok {
				errs = append(errs, fmt.Errorf("unknown field %q in value updates", fieldName))
			}
			if err := field.ValidateValue(value); err != nil {
				errs = append(errs, fmt.Errorf("invalid value for field %q: %w", fieldName, err))
			}
			return true
		})
		if len(errs) > 0 {
			return fmt.Errorf("validation errors: %v", errs)
		}
		return nil
	} else {
		return ValidateKey(fields, value)
	}
}
