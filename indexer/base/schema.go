package indexerbase

import "fmt"

type Schema struct {
	Tables []Table
}

type Table struct {
	Name         string
	KeyColumns   []Column
	ValueColumns []Column

	// RetainDeletions is a flag that indicates whether the indexer should retain
	// deleted rows in the database and flag them as deleted rather than actually
	// deleting the row. For many types of data in state, the data is deleted even
	// though it is still valid in order to save space. Indexers will want to have
	// the option of retaining such data and distinguishing from other "true" deletions.
	RetainDeletions bool
}

type Column struct {
	Name           string
	Type           Type
	Nullable       bool
	AddressPrefix  string
	EnumDefinition EnumDefinition
}

type EnumDefinition struct {
	Name   string
	Values []string
}

func (c Column) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("column name cannot be empty")
	}
	if err := c.Type.Validate(); err != nil {
		return fmt.Errorf("invalid column type for %q: %w", c.Name, err)
	}
	if c.Type == TypeBech32Address && c.AddressPrefix == "" {
		return fmt.Errorf("missing address prefix for column %q", c.Name)
	}
	if c.Type == TypeEnum {
		if err := c.EnumDefinition.Validate(); err != nil {
			return fmt.Errorf("invalid enum definition for column %q: %w", c.Name, err)
		}
	}
	return nil
}

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

func (c Column) ValidateValue(value any) error {
	if value == nil {
		if !c.Nullable {
			return fmt.Errorf("column %q cannot be null", c.Name)
		}
		return nil
	}
	return c.Type.ValidateValue(value)
}

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
