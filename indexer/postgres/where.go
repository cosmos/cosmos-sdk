package postgres

import (
	"fmt"
	"io"
)

func (tm *TableManager) WhereSqlAndParams(w io.Writer, key interface{}, startParamIdx int) (endParamIdx int, keyParams []interface{}, err error) {
	var keyCols []string
	keyParams, keyCols, err = tm.bindKeyParams(key)
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(w, " WHERE ")
	if err != nil {
		return
	}

	endParamIdx = startParamIdx
	for i, col := range keyCols {
		if i > 0 {
			_, err = fmt.Fprintf(w, " AND ")
			if err != nil {
				return
			}
		}
		_, err = fmt.Fprintf(w, "%s = $%d", col, endParamIdx)
		if err != nil {
			return
		}

		endParamIdx++
	}

	return
}
