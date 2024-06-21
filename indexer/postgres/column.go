package postgres

import (
	"fmt"
	"io"

	"cosmossdk.io/schema"
)

func (tm *TableManager) createColumnDef(writer io.Writer, field schema.Field) error {
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
			_, err = fmt.Fprintf(writer, "%q", enumTypeName(tm.moduleName, field.EnumDefinition))
			if err != nil {
				return err
			}
		case schema.Bech32AddressKind:
			_, err = fmt.Fprintf(writer, "BYTEA") // TODO: string conversion
			if err != nil {
				return err
			}

		case schema.TimeKind:
			nanosCol := fmt.Sprintf("%s_nanos", field.Name)
			// TODO: retain at least microseconds in the timestamp
			_, err = fmt.Fprintf(writer, "TIMESTAMPTZ GENERATED ALWAYS AS (to_timestamp(%q / 1000000000)) STORED,\n\t", nanosCol)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(writer, `%q BIGINT`, nanosCol)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected kind: %v, this should have been handled earlier", field.Kind)
		}

		return writeNullability(writer, field.Nullable)
	}
}

func writeNullability(writer io.Writer, nullable bool) error {
	if nullable {
		_, err := fmt.Fprintf(writer, " NULL,\n\t")
		return err
	} else {
		_, err := fmt.Fprintf(writer, " NOT NULL,\n\t")
		return err
	}
}

func simpleColumnType(kind schema.Kind) string {
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
	case schema.IntegerStringKind:
		return "NUMERIC"
	case schema.DecimalStringKind:
		return "NUMERIC"
	case schema.Float32Kind:
		return "REAL"
	case schema.Float64Kind:
		return "DOUBLE PRECISION"
	case schema.JSONKind:
		return "JSONB"
	case schema.DurationKind:
		// TODO: set COMMENT on field indicating nanoseconds unit
		return "BIGINT"
	default:
		return ""
	}
}

func (tm *TableManager) updatableColumnName(field schema.Field) (name string, err error) {
	name = field.Name
	if field.Kind == schema.TimeKind {
		name = fmt.Sprintf("%s_nanos", name)
	}
	name = fmt.Sprintf("%q", name)
	return
}
