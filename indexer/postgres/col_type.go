package postgres

import (
	"fmt"

	indexerbase "cosmossdk.io/indexer/base"
)

func columnType(field indexerbase.Field) (string, error) {
	switch field.Kind {
	case indexerbase.StringKind:
		return "TEXT", nil
	case indexerbase.BoolKind:
		return "BOOLEAN", nil
	case indexerbase.BytesKind:
		return "BYTEA", nil
	case indexerbase.Int8Kind:
		return "SMALLINT", nil
	case indexerbase.Int16Kind:
		return "SMALLINT", nil
	case indexerbase.Int32Kind:
		return "INTEGER", nil
	case indexerbase.Int64Kind:
		return "BIGINT", nil
	case indexerbase.Uint8Kind:
		return "SMALLINT", nil
	case indexerbase.Uint16Kind:
		return "INTEGER", nil
	case indexerbase.Uint32Kind:
		return "BIGINT", nil
	case indexerbase.Uint64Kind:
		return "NUMERIC", nil
	case indexerbase.DecimalKind:
		return "NUMERIC", nil
	case indexerbase.Float32Kind:
		return "REAL", nil
	case indexerbase.Float64Kind:
		return "DOUBLE PRECISION", nil
	case indexerbase.EnumKind:
		return "TEXT", fmt.Errorf("enums not supported yet")
	case indexerbase.JSONKind:
		return "JSONB", nil
	case indexerbase.Bech32AddressKind:
		return "TEXT", nil
	default:
		return "", fmt.Errorf("unsupported kind %v for field %s", field.Kind, field.Name)
	}
}
