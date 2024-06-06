package postgres

import (
	"fmt"
	"io"
	"strings"

	indexerbase "cosmossdk.io/indexer/base"
)

func (i *indexer) insertOrUpdater(table indexerbase.Table) {

}

type tableUpdater struct {
	table indexerbase.Table
}

func (tu tableUpdater) insertOrUpdate(w io.Writer, update indexerbase.EntityUpdate) error {
	cols := make([]string, 0, len(tu.table.KeyColumns)+len(tu.table.ValueColumns))
	for _, col := range tu.table.KeyColumns {
		cols = append(cols, col.Name)
	}
	for _, col := range tu.table.ValueColumns {
		cols = append(cols, col.Name)
	}

	_, err := fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (", tu.table.Name, strings.Join(cols, ", "))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, ") ON CONFLICT (", tu.table.Name)
	if err != nil {
		return err
	}

	return nil
}
