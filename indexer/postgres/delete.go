package postgres

import (
	"fmt"
	"strings"
)

func (tu *tableInfo) deleteSql() (string, error) {
	w := &strings.Builder{}

	keyCols := make([]string, 0, len(tu.table.KeyColumns))
	if len(tu.table.KeyColumns) == 0 {
		keyCols = append(keyCols, "_id")
	} else {
		for _, col := range tu.table.KeyColumns {
			keyCols = append(keyCols, col.Name)
		}
	}

	_, err := fmt.Fprintf(w, "DELETE FROM %s WHERE ", tu.table.Name)
	if err != nil {
		return "", err
	}

	for i, col := range keyCols {
		if i > 0 {
			_, err = fmt.Fprintf(w, " AND ")
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
