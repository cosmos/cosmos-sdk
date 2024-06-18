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

	// EnumType must refer to an enum definition in the module schema and must only be set when
	// the field's kind is EnumKind.
	EnumType string
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

	// enum type only valid with EnumKind
	if c.Kind == EnumKind {
		if c.EnumType == "" {
			return fmt.Errorf("missing EnumType for field %q", c.Name)
		}
	} else if c.Kind != EnumKind && c.EnumType != "" {
		return fmt.Errorf("EnumType is only valid for field with type EnumKind, found it on %s", c.Name)
	}

	return nil
}

// Validate validates the enum definition.
func (e EnumDefinition) Validate() error {
	if len(e.Values) == 0 {
		return fmt.Errorf("enum definition values cannot be empty")
	}
	seen := make(map[string]bool, len(e.Values))
	for i, v := range e.Values {
		if v == "" {
			return fmt.Errorf("enum definition value at index %d cannot be empty", i)
		}
		if seen[v] {
			return fmt.Errorf("duplicate enum definition value %q", v)
		}
		seen[v] = true
	}
	return nil
}

// ValidateValue validates that the value corresponds to the field's kind and nullability, but
// cannot check for enum value correctness.
func (c Field) ValidateValue(value interface{}) error {
	if value == nil {
		if !c.Nullable {
			return fmt.Errorf("field %q cannot be null", c.Name)
		}
		return nil
	}

	return c.Kind.ValidateValue(value)
}
