package indexerbase

import "fmt"

// EnumDefinition represents the definition of an enum type.
type EnumDefinition struct {

	// Values is a list of distinct values that are part of the enum type.
	Values []string
}

func (e EnumDefinition) ValidateValue(value string) error {
	for _, v := range e.Values {
		if v == value {
			return nil
		}
	}
	return fmt.Errorf("value %q is not a valid enum value for %s", value, e.Name)
}
