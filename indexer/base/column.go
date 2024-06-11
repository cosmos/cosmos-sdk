package indexerbase

import "fmt"

// Column represents a column in a table schema.
type Column struct {
	// Name is the name of the column.
	Name string

	// Kind is the basic type of the column.
	Kind Kind

	// Nullable indicates whether null values are accepted for the column.
	Nullable bool

	// AddressPrefix is the address prefix of the column's kind, currently only used for Bech32AddressKind.
	AddressPrefix string

	// EnumDefinition is the definition of the enum type and is only valid when Kind is EnumKind.
	EnumDefinition EnumDefinition
}

// EnumDefinition represents the definition of an enum type.
type EnumDefinition struct {
	// Name is the name of the enum type.
	Name string

	// Values is a list of distinct values that are part of the enum type.
	Values []string
}

// Validate validates the column.
func (c Column) Validate() error {
	// non-empty name
	if c.Name == "" {
		return fmt.Errorf("column name cannot be empty")
	}

	// valid kind
	if err := c.Kind.Validate(); err != nil {
		return fmt.Errorf("invalid column type for %q: %w", c.Name, err)
	}

	// address prefix only valid with Bech32AddressKind
	if c.Kind == Bech32AddressKind && c.AddressPrefix == "" {
		return fmt.Errorf("missing address prefix for column %q", c.Name)
	} else if c.Kind != Bech32AddressKind && c.AddressPrefix != "" {
		return fmt.Errorf("address prefix is only valid for column %q with type Bech32AddressKind", c.Name)
	}

	// enum definition only valid with EnumKind
	if c.Kind == EnumKind {
		if err := c.EnumDefinition.Validate(); err != nil {
			return fmt.Errorf("invalid enum definition for column %q: %w", c.Name, err)
		}
	} else if c.Kind != EnumKind && c.EnumDefinition.Name != "" && c.EnumDefinition.Values != nil {
		return fmt.Errorf("enum definition is only valid for column %q with type EnumKind", c.Name)
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

// ValidateValue validates that the value conforms to the column's kind and nullability.
func (c Column) ValidateValue(value any) error {
	if value == nil {
		if !c.Nullable {
			return fmt.Errorf("column %q cannot be null", c.Name)
		}
		return nil
	}
	return c.Kind.ValidateValue(value)
}

// ValidateKey validates that the value conforms to the set of columns as a Key in an EntityUpdate.
// See EntityUpdate.Key for documentation on the requirements of such values.
func ValidateKey(cols []Column, value any) error {
	if len(cols) == 0 {
		return nil
	}

	if len(cols) == 1 {
		return cols[0].ValidateValue(value)
	}

	values, ok := value.([]any)
	if !ok {
		return fmt.Errorf("expected slice of values for key columns, got %T", value)
	}

	if len(cols) != len(values) {
		return fmt.Errorf("expected %d key columns, got %d values", len(cols), len(value.([]any)))
	}
	for i, col := range cols {
		if err := col.ValidateValue(values[i]); err != nil {
			return fmt.Errorf("invalid value for key column %q: %w", col.Name, err)
		}
	}
	return nil
}

// ValidateValue validates that the value conforms to the set of columns as a Value in an EntityUpdate.
// See EntityUpdate.Value for documentation on the requirements of such values.
func ValidateValue(cols []Column, value any) error {
	valueUpdates, ok := value.(ValueUpdates)
	if ok {
		colMap := map[string]Column{}
		for _, col := range cols {
			colMap[col.Name] = col
		}
		var errs []error
		valueUpdates.Iterate(func(colName string, value any) bool {
			col, ok := colMap[colName]
			if !ok {
				errs = append(errs, fmt.Errorf("unknown column %q in value updates", colName))
			}
			if err := col.ValidateValue(value); err != nil {
				errs = append(errs, fmt.Errorf("invalid value for column %q: %w", colName, err))
			}
			return true
		})
		if len(errs) > 0 {
			return fmt.Errorf("validation errors: %v", errs)
		}
		return nil
	} else {
		return ValidateKey(cols, value)
	}
}
