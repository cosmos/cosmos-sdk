package sqllite

import (
	"bytes"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"cosmossdk.io/store/v2"
)

var _ store.Iterator = (*iterator)(nil)

type iterator struct {
	statement  *sql.Stmt
	rows       *sql.Rows
	key, val   []byte
	start, end []byte
	valid      bool
	err        error
}

func newIterator(storage *sql.DB, storeKey string, version uint64, start, end []byte) (*iterator, error) {
	// XXX(bez): Think of a cleaner way of creating the SQL statement to handle
	// end domain existing or not without resorting to fmt.Sprintf or relying on
	// a SQL builder (which can be a last resort option). For now, this will suffice.
	var (
		stmt      *sql.Stmt
		queryArgs []any
		err       error
	)
	if len(end) > 0 {
		stmt, err = storage.Prepare(`
		SELECT x.key, x.value
		FROM (
				SELECT key, value, version,
						row_number() OVER (PARTITION BY key ORDER BY version DESC) AS _rn
				FROM state_storage WHERE store_key = ? AND version <= ? AND key >= ? AND key < ?
		) x
		WHERE x._rn = 1 ORDER BY x.key ASC;
		`)
		queryArgs = []any{storeKey, version, start, end}
	} else {
		stmt, err = storage.Prepare(`
		SELECT x.key, x.value
		FROM (
				SELECT key, value, version,
						row_number() OVER (PARTITION BY key ORDER BY version DESC) AS _rn
				FROM state_storage WHERE store_key = ? AND version <= ? AND key >= ?
		) x
		WHERE x._rn = 1 ORDER BY x.key ASC;
		`)
		queryArgs = []any{storeKey, version, start}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	rows, err := stmt.Query(queryArgs...)
	if err != nil {
		_ = stmt.Close()
		return nil, fmt.Errorf("failed to execute SQL query: %w", err)
	}

	itr := &iterator{
		statement: stmt,
		rows:      rows,
		start:     start,
		end:       end,
		valid:     rows.Next(),
	}
	if !itr.valid {
		itr.err = fmt.Errorf("iterator invalid: %w", sql.ErrNoRows)
		return itr, nil
	}

	// read the first row
	itr.parseRow()
	if !itr.valid {
		itr.err = fmt.Errorf("iterator invalid: %w", itr.err)
		return itr, nil
	}

	return itr, nil
}

func (itr *iterator) Close() {
	_ = itr.statement.Close()
	itr.statement = nil
	itr.rows = nil
}

// Domain returns the domain of the iterator. The caller must not modify the
// return values.
func (itr *iterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

func (itr *iterator) Key() []byte {
	itr.assertIsValid()

	keyCopy := make([]byte, len(itr.key))
	_ = copy(keyCopy, itr.key)

	return keyCopy
}

func (itr *iterator) Value() []byte {
	itr.assertIsValid()

	valCopy := make([]byte, len(itr.val))
	_ = copy(valCopy, itr.val)

	return valCopy
}

func (itr *iterator) Valid() bool {
	if !itr.valid || itr.rows.Err() != nil {
		itr.valid = false
		return itr.valid
	}

	// if key is at the end or past it, consider it invalid
	if end := itr.end; end != nil {
		if bytes.Compare(end, itr.Key()) <= 0 {
			itr.valid = false
			return itr.valid
		}
	}

	return true
}

func (itr *iterator) Next() bool {
	if itr.rows.Next() {
		itr.parseRow()
		return itr.Valid()
	}

	itr.valid = false
	return itr.valid
}

func (itr *iterator) Error() error {
	if err := itr.rows.Err(); err != nil {
		return err
	}

	return itr.err
}

func (itr *iterator) parseRow() {
	var (
		key   []byte
		value []byte
	)
	if err := itr.rows.Scan(&key, &value); err != nil {
		itr.err = fmt.Errorf("failed to scan row: %s", err)
		itr.valid = false
		return
	}

	itr.key = key
	itr.val = value
}

func (itr *iterator) assertIsValid() {
	if !itr.valid {
		panic("iterator is invalid")
	}
}
