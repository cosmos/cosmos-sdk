package postgres

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	indexerbase "cosmossdk.io/indexer/base"
)

func (i *indexer) createTableStatement(tableSchema indexerbase.Table) (string, error) {
	w := &bytes.Buffer{}
	_, err := fmt.Fprintf(w, "CREATE TABLE IF NOT EXISTS %s (\n\t", tableSchema.Name)
	if err != nil {
		return "", err
	}

	isSingleton := false
	if len(tableSchema.KeyColumns) == 0 {
		isSingleton = true
		_, err = fmt.Fprintf(w, "_id INTEGER NOT NULL CHECK (_id = 1),\n\t")
	} else {
		for _, col := range tableSchema.KeyColumns {
			err = i.createColumnDef(w, col)
			if err != nil {
				return "", err
			}
		}
	}

	for _, col := range tableSchema.ValueColumns {
		err = i.createColumnDef(w, col)
		if err != nil {
			return "", err
		}
	}

	var pKeys []string
	if !isSingleton {
		for _, col := range tableSchema.KeyColumns {
			pKeys = append(pKeys, col.Name)
		}
	} else {
		pKeys = []string{"_id"}
	}

	_, err = fmt.Fprintf(w, "PRIMARY KEY (%s)\n", strings.Join(pKeys, ", "))
	if err != nil {
		return "", err
	}

	_, err = fmt.Fprintf(w, ");")
	if err != nil {
		return "", err
	}

	return w.String(), nil
}

func (i *indexer) createColumnDef(w io.Writer, col indexerbase.Column) error {
	typeStr, err := i.colType(col)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "%s %s,\n\t", col.Name, typeStr)
	return err
}

func (i *indexer) colType(col indexerbase.Column) (string, error) {
	switch col.Type {
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
		return "", fmt.Errorf("unsupported type %v", col.Type)
	}
}
