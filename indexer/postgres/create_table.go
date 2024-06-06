package postgres

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	indexerbase "cosmossdk.io/indexer/base"
)

func (i indexer) createTableStatement(tableSchema indexerbase.Table) (string, error) {
	w := &bytes.Buffer{}
	_, err := fmt.Fprintf(w, "CREATE TABLE %s (\n\t", tableSchema.Name)
	if err != nil {
		return "", err
	}

	for _, col := range tableSchema.KeyColumns {
		err = i.createColumnDef(w, col)
		if err != nil {
			return "", err
		}
	}

	for _, col := range tableSchema.ValueColumns {
		err = i.createColumnDef(w, col)
		if err != nil {
			return "", err
		}
	}

	var pKeys []string
	for _, col := range tableSchema.KeyColumns {
		pKeys = append(pKeys, col.Name)
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

func (i indexer) createColumnDef(w io.Writer, col indexerbase.Column) error {
	typeStr, err := i.colType(col)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "%s %s,\n\t", col.Name, typeStr)
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
