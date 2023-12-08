package sqlite

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/storage"
)

const (
	driverName       = "sqlite3"
	dbName           = "file:ss.db?cache=shared&mode=rwc&_journal_mode=WAL"
	reservedStoreKey = "_RESERVED_"
	keyLatestHeight  = "latest_height"
	keyPruneHeight   = "prune_height"

	reservedUpsertStmt = `
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
	delStmt = `
	UPDATE state_storage SET tombstone = ?
	WHERE id = (
		SELECT id FROM state_storage WHERE store_key = ? AND key = ? AND version <= ? ORDER BY version DESC LIMIT 1
	) AND tombstone = 0;
	`
)

var _ storage.Database = (*Database)(nil)

type Database struct {
	storage *sql.DB

	// earliestVersion defines the earliest version set in the database, which is
	// only updated when the database is pruned.
	earliestVersion uint64
}

func New(dataDir string) (*Database, error) {
	storage, err := sql.Open(driverName, filepath.Join(dataDir, dbName))
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
		tombstone integer unsigned default 0,
		unique (store_key, key, version)
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_store_key_version ON state_storage (store_key, key, version);
	`
	_, err = storage.Exec(stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	pruneHeight, err := getPruneHeight(storage)
	if err != nil {
		return nil, fmt.Errorf("failed to get prune height: %w", err)
	}

	return &Database{
		storage:         storage,
		earliestVersion: pruneHeight + 1,
	}, nil
}

func (db *Database) Close() error {
	err := db.storage.Close()
	db.storage = nil
	return err
}

func (db *Database) NewBatch(version uint64) (store.Batch, error) {
	return NewBatch(db.storage, version)
}

func (db *Database) GetLatestVersion() (uint64, error) {
	stmt, err := db.storage.Prepare("SELECT value FROM state_storage WHERE store_key = ? AND key = ?")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var latestHeight uint64
	if err := stmt.QueryRow(reservedStoreKey, keyLatestHeight).Scan(&latestHeight); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// in case of a fresh database
			return 0, nil
		}

		return 0, fmt.Errorf("failed to query row: %w", err)
	}

	return latestHeight, nil
}

func (db *Database) SetLatestVersion(version uint64) error {
	_, err := db.storage.Exec(reservedUpsertStmt, reservedStoreKey, keyLatestHeight, version, 0, version)
	if err != nil {
		return fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	return nil
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	val, err := db.Get(storeKey, version, key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (db *Database) Get(storeKey string, targetVersion uint64, key []byte) ([]byte, error) {
	if targetVersion < db.earliestVersion {
		return nil, store.ErrVersionPruned{EarliestVersion: db.earliestVersion}
	}

	stmt, err := db.storage.Prepare(`
	SELECT value, tombstone FROM state_storage
	WHERE store_key = ? AND key = ? AND version <= ?
	ORDER BY version DESC LIMIT 1;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var (
		value []byte
		tomb  uint64
	)
	if err := stmt.QueryRow(storeKey, key, targetVersion).Scan(&value, &tomb); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to query row: %w", err)
	}

	// A tombstone of zero or a target version that is less than the tombstone
	// version means the key is not deleted at the target version.
	if tomb == 0 || targetVersion < tomb {
		return value, nil
	}

	// the value is considered deleted
	return nil, nil
}

// Prune removes all versions of all keys that are <= the given version. It keeps
// the latest (non-tombstoned) version of each key/value tuple to handle queries
// above the prune version. This is analogous to RocksDB full_history_ts_low.
//
// We perform the prune by deleting all versions of a key, excluding reserved keys,
// that are <= the given version, except for the latest version of the key.
func (db *Database) Prune(version uint64) error {
	tx, err := db.storage.Begin()
	if err != nil {
		return fmt.Errorf("failed to create SQL transaction: %w", err)
	}

	pruneStmt := `DELETE FROM state_storage
	WHERE version < (
		SELECT max(version) FROM state_storage t2 WHERE
		t2.store_key = state_storage.store_key AND
		t2.key = state_storage.key AND
		t2.version <= ?
	) AND store_key != ?;
	`

	_, err = tx.Exec(pruneStmt, version, reservedStoreKey)
	if err != nil {
		return fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	// set the prune height so we can return <nil> for queries below this height
	_, err = tx.Exec(reservedUpsertStmt, reservedStoreKey, keyPruneHeight, version, 0, version)
	if err != nil {
		return fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to write SQL transaction: %w", err)
	}

	db.earliestVersion = version + 1

	return nil
}

func (db *Database) Iterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, store.ErrStartAfterEnd
	}

	return newIterator(db, storeKey, version, start, end, false)
}

func (db *Database) ReverseIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, store.ErrStartAfterEnd
	}

	return newIterator(db, storeKey, version, start, end, true)
}

func (db *Database) PrintRowsDebug() {
	stmt, err := db.storage.Prepare("SELECT store_key, key, value, version, tombstone FROM state_storage")
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
			key      []byte
			value    []byte
			version  uint64
			tomb     uint64
		)
		if err := rows.Scan(&storeKey, &key, &value, &version, &tomb); err != nil {
			panic(fmt.Sprintf("failed to scan row: %s", err))
		}

		sb.WriteString(fmt.Sprintf("STORE_KEY: %s, KEY: %s, VALUE: %s, VERSION: %d, TOMBSTONE: %d\n", storeKey, key, value, version, tomb))
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("received unexpected error: %w", err))
	}

	fmt.Println(strings.TrimSpace(sb.String()))
}

func getPruneHeight(storage *sql.DB) (uint64, error) {
	stmt, err := storage.Prepare(`SELECT value FROM state_storage WHERE store_key = ? AND key = ?`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var value uint64
	if err := stmt.QueryRow(reservedStoreKey, keyPruneHeight).Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}

		return 0, fmt.Errorf("failed to query row: %w", err)
	}

	return value, nil
}
