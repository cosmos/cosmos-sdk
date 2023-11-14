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
		tombstone integer unsigned default 0,
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
	err := db.storage.Close()
	db.storage = nil
	return err
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
	stmt, err := db.storage.Prepare(`
	SELECT value, tombstone, coalesce((SELECT value from state_storage WHERE store_key = ? AND key = ?), 0) as prune_height
	FROM state_storage
	WHERE store_key = ? AND key = ? AND version <= ?
	ORDER BY version DESC LIMIT 1;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var (
		value       []byte
		tomb        uint64
		pruneHeight uint64
	)
	if err := stmt.QueryRow(reservedStoreKey, keyPruneHeight, storeKey, key, targetVersion).Scan(&value, &tomb, &pruneHeight); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to query row: %w", err)
	}

	if pruneHeight > 0 && targetVersion <= pruneHeight {
		return nil, nil
	}

	// A tombstone of zero or a target version that is less than the tombstone
	// version means the key is not deleted at the target version.
	if tomb == 0 || targetVersion < tomb {
		return value, nil
	}

	// the value is considered deleted
	return nil, nil
}

func (db *Database) ApplyChangeset(version uint64, cs *store.Changeset) error {
	b, err := NewBatch(db.storage, version)
	if err != nil {
		return err
	}

	for _, kvPair := range cs.Pairs {
		if kvPair.Value == nil {
			if err := b.Delete(kvPair.StoreKey, kvPair.Key); err != nil {
				return err
			}
		} else {
			if err := b.Set(kvPair.StoreKey, kvPair.Key, kvPair.Value); err != nil {
				return err
			}
		}
	}

	return b.Write()
}

func (db *Database) Prune(version uint64) error {
	tx, err := db.storage.Begin()
	if err != nil {
		return fmt.Errorf("failed to create SQL transaction: %w", err)
	}

	// Delete all versions less than or equal to the target version, except we keep
	// the last one to handle above the prune version. This is analogous to RocksDB
	// full_history_ts_low.
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

	return nil
}

func (db *Database) getPruneHeight() (uint64, error) {
	stmt, err := db.storage.Prepare(`SELECT value FROM state_storage WHERE store_key = ? AND key = ?`)
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
