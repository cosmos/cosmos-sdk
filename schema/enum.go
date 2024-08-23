package schema

import (
	"errors"
	"fmt"
)

// EnumType represents the definition of an enum type.
type EnumType struct {
	// Name is the name of the enum type. It must conform to the NameFormat regular expression.
	// Its name must be unique between all enum types and object types in the module.
	// The same enum, however, can be used in multiple object types and fields as long as the
	// definition is identical each time
	Name string

	// Values is a list of distinct, non-empty values that are part of the enum type.
	// Each value must conform to the NameFormat regular expression.
	Values []EnumValueDefinition

	// NumericKind is the numeric kind used to represent the enum values numerically.
	// If it is left empty, Int32Kind is used by default.
	// Valid values are Uint8Kind, Int8Kind, Uint16Kind, Int16Kind, and Int32Kind.
	NumericKind Kind
}
// EnumValueDefinitio represents a value in an enum type.
type EnumValueDefinition struct {
	// Name is the name of the enum value. It must conform to the NameFormat regular expression.
	Name string

	// Value is the numeric value of the enum.
	Value int32
}


// TypeName implements the Type interface.
func (e EnumType) TypeName() string {
	return e.Name
}

func (EnumType) isType() {}

// Validate validates the enum definition.
func (e EnumType) Validate() error {
	if !ValidateName(e.Name) {
		return fmt.Errorf("invalid enum definition name %q", e.Name)
	}

	if len(e.Values) == 0 {
		return errors.New("enum definition values cannot be empty")
	}
	seen := make(map[string]bool, len(e.Values))
	for i, v := range e.Values {
		if !ValidateName(v.Name) {
			return fmt.Errorf("invalid enum definition value %q at index %d for enum %s", v, i, e.Name)
		}

		if seen[v.Name] {
			return fmt.Errorf("duplicate enum definition value %q for enum %s", v, e.Name)
		}
		seen[v.Name] = true
	}
	return nil
}

// ValidateValue validates that the value is a valid enum value.
func (e EnumType) ValidateValue(value string) error {
	for _, v := range e.Values {
		if v.Name == value {
			return nil
		}
	}
	return fmt.Errorf("value %q is not a valid enum value for %s", value, e.Name)
}
