package postgres

import (
	"fmt"
	"io"

	"cosmossdk.io/schema"
)

// createColumnDefinition writes a column definition within a CREATE TABLE statement for the field.
func (tm *objectIndexer) createColumnDefinition(writer io.Writer, field schema.Field) error {
	_, err := fmt.Fprintf(writer, "%q ", field.Name)
	if err != nil {
		return err
	}

	simple := simpleColumnType(field.Kind)
	if simple != "" {
		_, err = fmt.Fprintf(writer, "%s", simple)
		if err != nil {
			return err
		}

		return writeNullability(writer, field.Nullable)
	} else {
		switch field.Kind {
		case schema.EnumKind:
			_, err = fmt.Fprintf(writer, "%q", enumTypeName(tm.moduleName, field.ReferencedType))
			if err != nil {
				return err
			}
		case schema.TimeKind:
			// for time fields, we generate two columns:
			// - one with nanoseconds precision for lossless storage, suffixed with _nanos
			// - one as a timestamptz (microsecond precision) for ease of use, that is GENERATED
			nanosColName := fmt.Sprintf("%s_nanos", field.Name)
			_, err = fmt.Fprintf(writer, "TIMESTAMPTZ GENERATED ALWAYS AS (nanos_to_timestamptz(%q)) STORED,\n\t", nanosColName)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(writer, `%q BIGINT`, nanosColName)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected kind: %v, this should have been handled earlier", field.Kind)
		}

		return writeNullability(writer, field.Nullable)
	}
}

// writeNullability writes column nullability.
func writeNullability(writer io.Writer, nullable bool) error {
	if nullable {
		_, err := fmt.Fprintf(writer, " NULL,\n\t")
		return err
	} else {
		_, err := fmt.Fprintf(writer, " NOT NULL,\n\t")
		return err
	}
}

// simpleColumnType returns the postgres column type for the kind for simple types.
func simpleColumnType(kind schema.Kind) string {
	//nolint:goconst // adding constants for these postgres type names would impede readability
	switch kind {
	case schema.StringKind:
		return "TEXT"
	case schema.BoolKind:
		return "BOOLEAN"
	case schema.BytesKind:
		return "BYTEA"
	case schema.Int8Kind:
		return "SMALLINT"
	case schema.Int16Kind:
		return "SMALLINT"
	case schema.Int32Kind:
		return "INTEGER"
	case schema.Int64Kind:
		return "BIGINT"
	case schema.Uint8Kind:
		return "SMALLINT"
	case schema.Uint16Kind:
		return "INTEGER"
	case schema.Uint32Kind:
		return "BIGINT"
	case schema.Uint64Kind:
		return "NUMERIC"
	case schema.IntegerKind:
		return "NUMERIC"
	case schema.DecimalKind:
		return "NUMERIC"
	case schema.Float32Kind:
		return "REAL"
	case schema.Float64Kind:
		return "DOUBLE PRECISION"
	case schema.JSONKind:
		return "JSONB"
	case schema.DurationKind:
		return "BIGINT"
	case schema.AddressKind:
		return "TEXT"
	default:
		return ""
	}
}

// updatableColumnName is the name of the insertable/updatable column name for the field.
// This is the field name in most cases, except for time columns which are stored as nanos
// and then converted to timestamp generated columns.
func (tm *objectIndexer) updatableColumnName(field schema.Field) (name string, err error) {
	name = field.Name
	if field.Kind == schema.TimeKind {
		name = fmt.Sprintf("%s_nanos", name)
	}
	name = fmt.Sprintf("%q", name)
	return
}
