package sqllite

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"cosmossdk.io/store/v2"
	_ "modernc.org/sqlite"
)

const (
	driverName       = "sqlite"
	dbName           = "ss.db"
	reservedStoreKey = "RESERVED"
	keyLatestHeight  = "latest_height"

	latestVersionStmt = `
	INSERT INTO state_storage(store_key, key, value, version)
    VALUES(?, ?, ?, ?)
  ON CONFLICT(store_key, key, version) DO UPDATE SET
    value = ?;
	`
	upsertStmt = `
	INSERT INTO state_storage(store_key, key, value, version)
    VALUES(?, ?, ?, ?)
  ON CONFLICT(store_key, key, version) DO UPDATE SET
    value = ?;
	`
	delStmt = "DELETE FROM state_storage WHERE store_key = ? AND key = ? AND version = ?;"
)

var _ store.VersionedDatabase = (*Database)(nil)

type Database struct {
	storage *sql.DB
}

func New(dataDir string) (*Database, error) {
	db, err := sql.Open(driverName, filepath.Join(dataDir, dbName))
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite DB: %w", err)
	}

	stmt := `
	CREATE TABLE IF NOT EXISTS state_storage (
		id integer not null primary key, 
		store_key varchar not null,
		key varchar not null,
		value varchar not null,
		version integer unsigned not null,
		unique (store_key, key, version)
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_store_key_version ON state_storage (store_key, key, version);
	`
	_, err = db.Exec(stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	return &Database{
		storage: db,
	}, nil
}

func (db *Database) Close() error {
	return db.storage.Close()
}

func (db *Database) GetLatestVersion() (uint64, error) {
	stmt, err := db.storage.Prepare("SELECT value FROM state_storage WHERE store_key = ? AND key = ?")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var latestHeight uint64
	if err := stmt.QueryRow(reservedStoreKey, keyLatestHeight).Scan(&latestHeight); err != nil {
		return 0, fmt.Errorf("failed to query row: %w", err)
	}

	if latestHeight == 0 {
		return 0, store.ErrVersionNotFound
	}

	return latestHeight, nil
}

func (db *Database) SetLatestVersion(version uint64) error {
	_, err := db.storage.Exec(latestVersionStmt, reservedStoreKey, keyLatestHeight, version, 0, version)
	if err != nil {
		return fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	return nil
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	stmt, err := db.storage.Prepare("SELECT EXISTS(SELECT 1 FROM state_storage WHERE store_key = ? AND key = ? AND version = ?);")
	if err != nil {
		return false, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var exists bool
	if err := stmt.QueryRow(storeKey, key, version).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to query row: %w", err)
	}

	return exists, nil
}

func (db *Database) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	stmt, err := db.storage.Prepare("SELECT value FROM state_storage WHERE store_key = ? AND key = ? AND version = ?;")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var value []byte
	if err := stmt.QueryRow(storeKey, key, version).Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to query row: %w", err)
	}

	return value, nil
}

func (db *Database) Set(storeKey string, version uint64, key, value []byte) error {
	_, err := db.storage.Exec(upsertStmt, storeKey, key, value, version, value)
	if err != nil {
		return fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	return nil
}

func (db *Database) Delete(storeKey string, version uint64, key []byte) error {
	_, err := db.storage.Exec(delStmt, storeKey, key, version)
	if err != nil {
		return fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	return nil
}

func (db *Database) NewBatch(version uint64) (store.Batch, error) {
	return NewBatch(db.storage, version)
}

func (db *Database) NewIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	panic("not implemented!")
}

func (db *Database) NewReverseIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	panic("not implemented!")
}

func (db *Database) printRowsDebug() {
	stmt, err := db.storage.Prepare("SELECT store_key, key, value, version FROM state_storage")
	if err != nil {
		panic(fmt.Errorf("failed to prepare SQL statement: %w", err))
	}

	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		panic(fmt.Errorf("failed to execute SQL query: %w", err))
	}

	var sb strings.Builder
	for rows.Next() {
		var (
			storeKey string
			key      string
			value    string
			version  uint64
		)
		if err := rows.Scan(&storeKey, &key, &value, &version); err != nil {
			panic(fmt.Sprintf("failed to scan row: %s", err))
		}

		sb.WriteString(fmt.Sprintf("STORE_KEY: %s, KEY: %s, VALUE: %s, VERSION: %d\n", storeKey, key, value, version))
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("received unexpected error: %w", err))
	}

	fmt.Println(sb.String())
}
