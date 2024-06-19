package postgres

import (
	"fmt"
	"io"

	indexerbase "cosmossdk.io/indexer/base"
)

func (tm *TableManager) createColumnDef(writer io.Writer, field indexerbase.Field) error {
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
		case indexerbase.EnumKind:
			_, err = fmt.Fprintf(writer, "TEXT") // TODO: enum type
			if err != nil {
				return err
			}
		case indexerbase.Bech32AddressKind:
			_, err = fmt.Fprintf(writer, "BYTEA") // TODO: string conversion
			if err != nil {
				return err
			}

		case indexerbase.TimeKind:
			nanosCol := fmt.Sprintf("%s_nanos", field.Name)
			_, err = fmt.Fprintf(writer, "TIMESTAMPTZ GENERATED ALWAYS AS (to_timestamp(%q)) STORED,\n\t", nanosCol)
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

func simpleColumnType(kind indexerbase.Kind) string {
	switch kind {
	case indexerbase.StringKind:
		return "TEXT"
	case indexerbase.BoolKind:
		return "BOOLEAN"
	case indexerbase.BytesKind:
		return "BYTEA"
	case indexerbase.Int8Kind:
		return "SMALLINT"
	case indexerbase.Int16Kind:
		return "SMALLINT"
	case indexerbase.Int32Kind:
		return "INTEGER"
	case indexerbase.Int64Kind:
		return "BIGINT"
	case indexerbase.Uint8Kind:
		return "SMALLINT"
	case indexerbase.Uint16Kind:
		return "INTEGER"
	case indexerbase.Uint32Kind:
		return "BIGINT"
	case indexerbase.Uint64Kind:
		return "NUMERIC"
	case indexerbase.IntegerKind:
		return "NUMERIC"
	case indexerbase.DecimalKind:
		return "NUMERIC"
	case indexerbase.Float32Kind:
		return "REAL"
	case indexerbase.Float64Kind:
		return "DOUBLE PRECISION"
	case indexerbase.JSONKind:
		return "JSONB"
	case indexerbase.DurationKind:
		// TODO: set COMMENT on field indicating nanoseconds unit
		return "BIGINT"
	default:
		return ""
	}
}
