package schema

import (
	"fmt"
	"time"
)

// Field represents a field in an object type.
type Field struct {
	// Name is the name of the field. It must conform to the NameFormat regular expression.
	Name string `json:"name"`

	// Kind is the basic type of the field.
	Kind Kind `json:"kind"`

	// Nullable indicates whether null values are accepted for the field. Key fields CANNOT be nullable.
	Nullable bool `json:"nullable,omitempty"`

	// ReferencedType is the referenced type name when Kind is EnumKind.
	ReferencedType string `json:"referenced_type,omitempty"`
}

// Validate validates the field.
func (c Field) Validate(typeSet TypeSet) error {
	// valid name
	if !ValidateName(c.Name) {
		return fmt.Errorf("invalid field name %q", c.Name)
	}

	// valid kind
	if err := c.Kind.Validate(); err != nil {
		return fmt.Errorf("invalid field kind for %q: %v", c.Name, err) //nolint:errorlint // false positive due to using go1.12
	}

	// enum definition only valid with EnumKind
	switch c.Kind {
	case EnumKind:
		if c.ReferencedType == "" {
			return fmt.Errorf("enum field %q must have a referenced type", c.Name)
		}

		ty, ok := typeSet.LookupType(c.ReferencedType)
		if !ok {
			return fmt.Errorf("enum field %q references unknown type %q", c.Name, c.ReferencedType)
		}

		if _, ok := ty.(EnumType); !ok {
			return fmt.Errorf("enum field %q references non-enum type %q", c.Name, c.ReferencedType)
		}
	default:
		if c.ReferencedType != "" {
			return fmt.Errorf("field %q with kind %q cannot have a referenced type", c.Name, c.Kind)
		}
	}

	return nil
}

// ValidateValue validates that the value conforms to the field's kind and nullability.
// Unlike Kind.ValidateValue, it also checks that the value conforms to the EnumType
// if the field is an EnumKind.
func (c Field) ValidateValue(value interface{}, typeSet TypeSet) error {
	if value == nil {
		if !c.Nullable {
			return fmt.Errorf("field %q cannot be null", c.Name)
		}
		return nil
	}
	err := c.Kind.ValidateValueType(value)
	if err != nil {
		return fmt.Errorf("invalid value for field %q: %v", c.Name, err) //nolint:errorlint // false positive due to using go1.12
	}

	switch c.Kind {
	case EnumKind:
		ty, ok := typeSet.LookupType(c.ReferencedType)
		if !ok {
			return fmt.Errorf("enum field %q references unknown type %q", c.Name, c.ReferencedType)
		}
		enumType, ok := ty.(EnumType)
		if !ok {
			return fmt.Errorf("enum field %q references non-enum type %q", c.Name, c.ReferencedType)
		}
		err := enumType.ValidateValue(value.(string))
		if err != nil {
			return fmt.Errorf("invalid value for enum field %q: %v", c.Name, err)
		}
	default:
	}

	return nil
}

// DefaultValue returns the default value for the field if one is known.
// If validDefault is false, the field does not have a known default value.
// TimeKind and JSONKind values do not have valid default values.
func (c Field) DefaultValue(typeSet TypeSet) (value interface{}, validDefault bool) {
	switch c.Kind {
	case StringKind:
		return "", true
	case Int8Kind:
		return int8(0), true
	case Int16Kind:
		return int16(0), true
	case Int32Kind:
		return int32(0), true
	case Int64Kind:
		return int64(0), true
	case Uint8Kind:
		return uint8(0), true
	case Uint16Kind:
		return uint16(0), true
	case Uint32Kind:
		return uint32(0), true
	case Uint64Kind:
		return uint64(0), true
	case Float32Kind:
		return float32(0), true
	case Float64Kind:
		return float64(0), true
	case BoolKind:
		return false, true
	case EnumKind:
		t, found := typeSet.LookupEnumType(c.ReferencedType)
		if !found {
			return nil, false
		}
		return t.DefaultValue()
	case BytesKind:
		return []byte{}, true
	case AddressKind:
		return []byte{}, true
	case DurationKind:
		return time.Duration(0), true
	case IntegerStringKind:
		return "0", true
	case DecimalStringKind:
		return "0", true
	case JSONKind:
		return nil, false
	case TimeKind:
		return nil, false
	default:
		return nil, false
	}
}
