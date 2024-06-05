package postgres

import (
	"fmt"
	"io"
	"strings"

	indexerbase "cosmossdk.io/indexer/base"
)

func (i indexer) createTableStatement(w io.Writer, tablePrefix string, tableSchema indexerbase.Table) error {
	_, err := fmt.Fprintf(w, "CREATE TABLE %s_%s (", tablePrefix, tableSchema.Name)
	if err != nil {
		return err
	}

	first := true
	for _, col := range tableSchema.KeyColumns {
		err = i.createColumnDef(w, col, &first)
		if err != nil {
			return err
		}
	}

	for _, col := range tableSchema.ValueColumns {
		err = i.createColumnDef(w, col, &first)
		if err != nil {
			return err
		}
	}

	var pKeys []string
	for _, col := range tableSchema.KeyColumns {
		pKeys = append(pKeys, col.Name)
	}
	_, err = fmt.Fprintf(w, "PRIMARY KEY (%s)", strings.Join(pKeys, ", "))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, ");")
	return err
}

func (i indexer) createColumnDef(w io.Writer, col indexerbase.Column, first *bool) error {
	typeStr, err := i.colType(col)
	if err != nil {
		return err
	}

	if !*first {
		_, err = fmt.Fprintf(w, ",\n")
		if err != nil {
			return err
		}

		*first = false
	}

	_, err = fmt.Fprintf(w, "%s %s", col.Name, typeStr)
	return err
}

func (i indexer) colType(col indexerbase.Column) (string, error) {
	switch col.Type {
	case indexerbase.TypeString:
		return "TEXT", nil
	case indexerbase.TypeBool:
		return "BOOLEAN", nil
	case indexerbase.TypeBytes:
		return "BYTEA", nil
	default:
		return "", fmt.Errorf("unsupported type %v", col.Type)
	}
}
