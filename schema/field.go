package schema

import "fmt"

// Field represents a field in an object type.
type Field struct {
	// Name is the name of the field. It must conform to the NameFormat regular expression.
	Name string

	// Kind is the basic type of the field.
	Kind Kind

	// Nullable indicates whether null values are accepted for the field. Key fields CANNOT be nullable.
	Nullable bool

	// EnumType is the definition of the enum type and is only valid when Kind is EnumKind.
	// The same enum types can be reused in the same module schema, but they always must contain
	// the same values for the same enum name. This possibly introduces some duplication of
	// definitions but makes it easier to reason about correctness and validation in isolation.
	EnumType EnumType

	// Width specifies the width of the field for IntNKind and UintNKind fields.
	// It is invalid to have a non-zero Width for other kinds.
	// Width must be a multiple of 8 and values of 8, 16, 32, and 64 are invalid because there
	// are more specific kinds for these widths.
	Width uint16

	// MaxLength specifies a maximum length for StringKind, BytesKind, AddressKind, and JSONKind fields.
	// If it is 0, the field has no maximum length.
	// It is invalid to have a non-zero MaxLength for other kinds.
	// Negative values are invalid.
	MaxLength int
}

// Validate validates the field.
func (c Field) Validate() error {
	// valid name
	if !ValidateName(c.Name) {
		return fmt.Errorf("invalid field name %q", c.Name)
	}

	// valid kind
	if err := c.Kind.Validate(); err != nil {
		return fmt.Errorf("invalid field kind for %q: %v", c.Name, err) //nolint:errorlint // false positive due to using go1.12
	}

	// enum definition only valid with EnumKind
	if c.Kind == EnumKind {
		if err := c.EnumType.Validate(); err != nil {
			return fmt.Errorf("invalid enum definition for field %q: %v", c.Name, err) //nolint:errorlint // false positive due to using go1.12
		}
	} else if c.Kind != EnumKind && (c.EnumType.Name != "" || c.EnumType.Values != nil) {
		return fmt.Errorf("enum definition is only valid for field %q with type EnumKind", c.Name)
	}

	switch c.Kind {
	case IntNKind, UIntNKind:
		switch c.Width {
		case 0, 8, 16, 32, 64:
			return fmt.Errorf("invalid width %d for field %q, use a more specific type", c.Width, c.Name)
		default:
			if c.Width%8 != 0 {
				return fmt.Errorf("invalid width %d for field %q, must be a multiple of 8", c.Width, c.Name)
			}
		}
	default:
		if c.Width != 0 {
			return fmt.Errorf("width %d is only valid for IntNKind and UIntNKind fields", c.Width)
		}
	}

	switch c.Kind {
	case StringKind, BytesKind, AddressKind, JSONKind:
		if c.MaxLength < 0 {
			return fmt.Errorf("negative max length %d for field %q", c.MaxLength, c.Name)
		}
	default:
		if c.MaxLength != 0 {
			return fmt.Errorf("max length %d is only valid for StringKind, BytesKind, AddressKind, and JSONKind fields", c.MaxLength)
		}
	}

	return nil
}

// ValidateValue validates that the value conforms to the field's kind and nullability.
// Unlike Kind.ValidateValue, it also checks that the value conforms to the EnumType
// if the field is an EnumKind.
func (c Field) ValidateValue(value interface{}) error {
	if value == nil {
		if !c.Nullable {
			return fmt.Errorf("field %q cannot be null", c.Name)
		}
		return nil
	}
	err := c.Kind.ValidateValue(value)
	if err != nil {
		return fmt.Errorf("invalid value for field %q: %v", c.Name, err) //nolint:errorlint // false positive due to using go1.12
	}

	switch c.Kind {
	case EnumKind:
		return c.EnumType.ValidateValue(value.(string))
	case IntNKind, UIntNKind:
		bz := value.([]byte)
		if len(bz) != int(c.Width/8) {
			return fmt.Errorf("invalid byte length %d for field %q, expected %d", len(bz), c.Name, c.Width/8)
		}
	case StringKind:
		if c.MaxLength > 0 {
			str := value.(string)
			if len(str) > c.MaxLength {
				return fmt.Errorf("value for field %q exceeds max length %d", c.Name, c.MaxLength)
			}
		}
	case BytesKind, AddressKind, JSONKind:
		if c.MaxLength > 0 {
			bz := value.([]byte)
			if len(bz) > c.MaxLength {
				return fmt.Errorf("value for field %q exceeds max length %d", c.Name, c.MaxLength)
			}
		}
	}

	return nil
}
