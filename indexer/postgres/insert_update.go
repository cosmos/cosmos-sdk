package postgres

import (
	"fmt"
	"io"
	"strings"
)

func (tm *TableManager) InsertUpdateSql(w io.Writer) error {
	// TODO: timestamp nanos column

	keyCols := make([]string, 0, len(tm.typ.KeyFields))
	if len(tm.typ.KeyFields) == 0 {
		keyCols = append(keyCols, "_id")
	} else {
		for _, col := range tm.typ.KeyFields {
			keyCols = append(keyCols, col.Name)
		}
	}

	valueCols := make([]string, 0, len(tm.typ.ValueFields))
	for _, col := range tm.typ.ValueFields {
		valueCols = append(valueCols, col.Name)
	}

	allCols := make([]string, 0, len(tm.typ.KeyFields)+len(tm.typ.ValueFields))
	allCols = append(allCols, keyCols...)
	allCols = append(allCols, valueCols...)

	_, err := fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (", tm.typ.Name, strings.Join(allCols, ", "))
	if err != nil {
		return err
	}

	for i, col := range allCols {
		if i > 0 {
			_, err = fmt.Fprintf(w, ", ")
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(w, "@%s", col)
		if err != nil {
			return err
		}

	}

	_, err = fmt.Fprintf(w, ") ON CONFLICT (%s) DO ", strings.Join(keyCols, ", "))
	if err != nil {
		return err
	}

	if len(valueCols) == 0 {
		_, err = fmt.Fprintf(w, "NOTHING")
		if err != nil {
			return err
		}
		return nil
	}

	_, err = fmt.Fprintf(w, "UPDATE SET ")
	if err != nil {
		return err
	}

	for i, col := range valueCols {
		if i > 0 {
			_, err = fmt.Fprintf(w, ", ")
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(w, "%s = @%s", col, col)
		if err != nil {
			return err
		}
	}

	return nil
}
