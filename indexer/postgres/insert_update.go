package postgres

import (
	"fmt"
	"strings"
)

func (tu *tableInfo) insertOrUpdateSql() (string, error) {
	w := &strings.Builder{}

	keyCols := make([]string, 0, len(tu.table.KeyColumns))
	if len(tu.table.KeyColumns) == 0 {
		keyCols = append(keyCols, "_id")
	} else {
		for _, col := range tu.table.KeyColumns {
			keyCols = append(keyCols, col.Name)
		}
	}

	valueCols := make([]string, 0, len(tu.table.ValueColumns))
	for _, col := range tu.table.ValueColumns {
		valueCols = append(valueCols, col.Name)
	}

	allCols := make([]string, 0, len(tu.table.KeyColumns)+len(tu.table.ValueColumns))
	allCols = append(allCols, keyCols...)
	allCols = append(allCols, valueCols...)

	_, err := fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (", tu.table.Name, strings.Join(allCols, ", "))
	if err != nil {
		return "", err
	}

	for i, col := range allCols {
		if i > 0 {
			_, err = fmt.Fprintf(w, ", ")
			if err != nil {
				return "", err
			}
		}
		_, err = fmt.Fprintf(w, "@%s", col)
		if err != nil {
			return "", err
		}

	}

	_, err = fmt.Fprintf(w, ") ON CONFLICT (%s) DO ", strings.Join(keyCols, ", "))
	if err != nil {
		return "", err
	}

	if len(valueCols) == 0 {
		_, err = fmt.Fprintf(w, "NOTHING")
		if err != nil {
			return "", err
		}
		return w.String(), nil
	}

	_, err = fmt.Fprintf(w, "UPDATE SET ")
	if err != nil {
		return "", err
	}

	for i, col := range valueCols {
		if i > 0 {
			_, err = fmt.Fprintf(w, ", ")
			if err != nil {
				return "", err
			}
		}
		_, err = fmt.Fprintf(w, "%s = @%s", col, col)
		if err != nil {
			return "", err
		}
	}

	return w.String(), nil
}
