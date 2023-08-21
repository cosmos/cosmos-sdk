package sqllite

import (
	"database/sql"
	"fmt"

	"cosmossdk.io/store/v2"
	_ "modernc.org/sqlite"
)

var _ store.Iterator = (*iterator)(nil)

type iterator struct {
	statement  *sql.Stmt
	rows       *sql.Rows
	storeKey   string
	version    uint64
	start, end []byte
}

func newIterator(storage *sql.DB, storeKey string, version uint64, start, end []byte) (*iterator, error) {
	stmt, err := storage.Prepare(`
	SELECT x.key, x.value, x.version 
	FROM (
			SELECT key, value, version,
					row_number() OVER (PARTITION BY key ORDER BY version DESC) AS _rn
			FROM state_storage WHERE store_key = ? AND version <= ? AND key >= ? AND key < ?
	) x
	WHERE x._rn = 1 ORDER BY x.key ASC;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	rows, err := stmt.Query(storeKey, version, start, end)
	if err != nil {
		_ = stmt.Close()
		return nil, fmt.Errorf("failed to execute SQL query: %w", err)
	}

	return &iterator{
		statement: stmt,
		rows:      rows,
		storeKey:  storeKey,
		version:   version,
		start:     start,
		end:       end,
	}, nil
}

func (itr *iterator) Close() {
	itr.statement.Close()
	itr.statement = nil
	itr.rows = nil
}

func (itr *iterator) Domain() ([]byte, []byte) {
	panic("not implemented!")
}

func (itr *iterator) Valid() bool {
	panic("not implemented!")
}

func (itr *iterator) Key() []byte {
	panic("not implemented!")
}

func (itr *iterator) Value() []byte {
	panic("not implemented!")
}

func (itr *iterator) Next() bool {
	panic("not implemented!")
}

func (itr *iterator) Error() error {
	panic("not implemented!")
}
