package indexerbase

import "fmt"

// EnumDefinition represents the definition of an enum type.
type EnumDefinition struct {
	// Name is the name of the enum type.
	Name string

	// Values is a list of distinct, non-empty values that are part of the enum type.
	Values []string
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

// ValidateValue validates that the value is a valid enum value.
func (e EnumDefinition) ValidateValue(value string) error {
	for _, v := range e.Values {
		if v == value {
			return nil
		}
	}
	return fmt.Errorf("value %q is not a valid enum value for %s", value, e.Name)
}
