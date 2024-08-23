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

	// ElementKind is the element type when Kind is ListKind.
	ElementKind Kind

	// String is the referenced type name when Kind is EnumKind, StructKind, OneOfKind or ObjectType.
	// When the main kind is ListKind, this type name is the referenced type of the ElementKind.
	Type string
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
	err := c.Kind.ValidateValueType(value)
	if err != nil {
		return fmt.Errorf("invalid value for field %q: %v", c.Name, err) //nolint:errorlint // false positive due to using go1.12
	}

	if c.Kind == EnumKind {
		return c.EnumType.ValidateValue(value.(string))
	}

	return nil
}
