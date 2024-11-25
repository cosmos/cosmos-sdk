package postgres

import (
	"fmt"
	"io"
)

// whereSqlAndParams generates a WHERE clause for the provided key and returns the parameters.
func (tm *objectIndexer) whereSqlAndParams(w io.Writer, key interface{}, startParamIdx int) (endParamIdx int, keyParams []interface{}, err error) {
	var keyCols []string
	keyParams, keyCols, err = tm.bindKeyParams(key)
	if err != nil {
		return
	}

	endParamIdx, keyParams, err = tm.whereSql(w, keyParams, keyCols, startParamIdx)
	return
}

// whereSql generates a WHERE clause for the provided columns and returns the parameters.
func (tm *objectIndexer) whereSql(w io.Writer, params []interface{}, cols []string, startParamIdx int) (endParamIdx int, resParams []interface{}, err error) {
	_, err = fmt.Fprintf(w, " WHERE ")
	if err != nil {
		return 0, nil, err
	}

	endParamIdx = startParamIdx
	for i, col := range cols {
		if i > 0 {
			_, err = fmt.Fprintf(w, " AND ")
			if err != nil {
				return 0, nil, err
			}
		}

		_, err = fmt.Fprintf(w, "%s ", col)
		if err != nil {
			return 0, nil, err
		}

		if params[i] == nil {
			_, err = fmt.Fprintf(w, "IS NULL")
			if err != nil {
				return 0, nil, err
			}

		} else {
			_, err = fmt.Fprintf(w, "= $%d", endParamIdx)
			if err != nil {
				return 0, nil, err
			}

			resParams = append(resParams, params[i])

			endParamIdx++
		}
	}

	return endParamIdx, resParams, nil
}
