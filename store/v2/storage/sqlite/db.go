package sqlite

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bvinc/go-sqlite-lite/sqlite3"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	storeerrors "cosmossdk.io/store/v2/errors"
	"cosmossdk.io/store/v2/storage"
)

const (
	driverName        = "sqlite3"
	dbName            = "ss.db"
	reservedStoreKey  = "_RESERVED_"
	keyLatestHeight   = "latest_height"
	keyPruneHeight    = "prune_height"
	valueRemovedStore = "removed_store"

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

var (
	_ storage.Database         = (*Database)(nil)
	_ store.UpgradableDatabase = (*Database)(nil)
)

type Database struct {
	storage   *sqlite3.Conn
	connStr   string
	writeLock *sync.Mutex
	// earliestVersion defines the earliest version set in the database, which is
	// only updated when the database is pruned.
	earliestVersion uint64
}

func New(dataDir string) (*Database, error) {
	connStr := fmt.Sprintf("file:%s", filepath.Join(dataDir, dbName))
	db, err := sqlite3.Open(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite DB: %w", err)
	}
	err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to set journal mode: %w", err)
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
	err = db.Exec(stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	pruneHeight, err := getPruneHeight(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get prune height: %w", err)
	}

	return &Database{
		storage:         db,
		connStr:         connStr,
		writeLock:       new(sync.Mutex),
		earliestVersion: pruneHeight,
	}, nil
}

func (db *Database) Close() error {
	err := db.storage.Close()
	db.storage = nil
	return err
}

func (db *Database) NewBatch(version uint64) (store.Batch, error) {
	conn, err := sqlite3.Open(db.connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite DB: %w", err)
	}

	return NewBatch(conn, db.writeLock, version)
}

func (db *Database) GetLatestVersion() (version uint64, err error) {
	stmt, err := db.storage.Prepare(`
	SELECT value
	FROM state_storage 
	WHERE store_key = ? AND key = ?
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer func(stmt *sqlite3.Stmt) {
		cErr := stmt.Close()
		if cErr != nil {
			err = fmt.Errorf("failed to close GetLatestVersion statement: %w", cErr)
		}
	}(stmt)

	err = stmt.Bind(reservedStoreKey, keyLatestHeight)
	if err != nil {
		return 0, fmt.Errorf("failed to bind GetLatestVersion statement: %w", err)
	}
	hasRow, err := stmt.Step()
	if err != nil {
		return 0, fmt.Errorf("failed to step through GetLatestVersion rows: %w", err)
	}
	if !hasRow {
		// in case of a fresh database
		return 0, nil
	}
	var v int64
	err = stmt.Scan(&v)
	if err != nil {
		return 0, fmt.Errorf("failed to scan GetLatestVersion row: %w", err)
	}
	version = uint64(v)
	return version, nil
}

func (db *Database) VersionExists(v uint64) (bool, error) {
	latestVersion, err := db.GetLatestVersion()
	if err != nil {
		return false, err
	}

	return latestVersion >= v && v >= db.earliestVersion, nil
}

func (db *Database) SetLatestVersion(version uint64) error {
	db.writeLock.Lock()
	defer db.writeLock.Unlock()
	err := db.storage.Exec(reservedUpsertStmt, reservedStoreKey, keyLatestHeight, int64(version), 0, int64(version))
	if err != nil {
		return fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	return nil
}

func (db *Database) Has(storeKey []byte, version uint64, key []byte) (bool, error) {
	val, err := db.Get(storeKey, version, key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (db *Database) Get(storeKey []byte, targetVersion uint64, key []byte) ([]byte, error) {
	if targetVersion < db.earliestVersion {
		return nil, storeerrors.ErrVersionPruned{EarliestVersion: db.earliestVersion, RequestedVersion: targetVersion}
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
		tomb  int64
	)
	err = stmt.Bind(storeKey, key, int64(targetVersion))
	if err != nil {
		return nil, fmt.Errorf("failed to bind SQL statement: %w", err)
	}
	hasRow, err := stmt.Step()
	if err != nil {
		return nil, fmt.Errorf("failed to step through SQL rows: %w", err)
	}
	if !hasRow {
		return nil, nil
	}
	err = stmt.Scan(&value, &tomb)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	// A tombstone of zero or a target version that is less than the tombstone
	// version means the key is not deleted at the target version.
	if tomb == 0 || targetVersion < uint64(tomb) {
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
	v := int64(version)
	db.writeLock.Lock()
	defer db.writeLock.Unlock()
	err := db.storage.Begin()
	if err != nil {
		return fmt.Errorf("failed to create SQL transaction: %w", err)
	}
	defer func() {
		if err != nil {
			err = db.storage.Rollback()
		}
	}()

	// prune all keys of old versions
	pruneStmt := `DELETE FROM state_storage
	WHERE version < (
		SELECT max(version) FROM state_storage t2 WHERE
		t2.store_key = state_storage.store_key AND
		t2.key = state_storage.key AND
		t2.version <= ?
	) AND store_key != ?;
	`
	if err := db.storage.Exec(pruneStmt, v, reservedStoreKey); err != nil {
		return fmt.Errorf("failed to exec prune keys statement: %w", err)
	}

	// prune removed stores
	pruneRemovedStoreKeysStmt := `DELETE FROM state_storage AS s
	WHERE EXISTS ( 
		SELECT 1 FROM
			(
			SELECT key, MAX(version) AS max_version
			FROM state_storage
			WHERE store_key = ? AND value = ? AND version <= ?
			GROUP BY key
			) AS t
		WHERE s.store_key = t.key AND s.version <= t.max_version LIMIT 1
	);
	`
	if err := db.storage.Exec(pruneRemovedStoreKeysStmt, reservedStoreKey, valueRemovedStore, v); err != nil {
		return fmt.Errorf("failed to exec prune store keys statement: %w", err)
	}

	// delete the removedKeys
	if err := db.storage.Exec("DELETE FROM state_storage WHERE store_key = ? AND value = ? AND version <= ?",
		reservedStoreKey, valueRemovedStore, v); err != nil {
		return fmt.Errorf("failed to exec remove keys statement: %w", err)
	}

	// set the prune height so we can return <nil> for queries below this height
	if err := db.storage.Exec(reservedUpsertStmt, reservedStoreKey, keyPruneHeight, v, 0, v); err != nil {
		return fmt.Errorf("failed to exec set prune height statement: %w", err)
	}

	if err := db.storage.Commit(); err != nil {
		return fmt.Errorf("failed to commit prune transaction: %w", err)
	}

	db.earliestVersion = version + 1
	return nil
}

func (db *Database) Iterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, storeerrors.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, storeerrors.ErrStartAfterEnd
	}

	return newIterator(db, storeKey, version, start, end, false)
}

func (db *Database) ReverseIterator(storeKey []byte, version uint64, start, end []byte) (corestore.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, storeerrors.ErrKeyEmpty
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, storeerrors.ErrStartAfterEnd
	}

	return newIterator(db, storeKey, version, start, end, true)
}

func (db *Database) PruneStoreKeys(storeKeys []string, version uint64) (err error) {
	db.writeLock.Lock()
	defer db.writeLock.Unlock()
	err = db.storage.Begin()
	if err != nil {
		return fmt.Errorf("failed to create SQL transaction: %w", err)
	}
	defer func() {
		if err != nil {
			err = db.storage.Rollback()
		}
	}()

	// flush removed store keys
	flushRemovedStoreKeyStmt := `INSERT INTO state_storage(store_key, key, value, version) 
		VALUES (?, ?, ?, ?)`
	for _, storeKey := range storeKeys {
		if err := db.storage.Exec(flushRemovedStoreKeyStmt, reservedStoreKey, []byte(storeKey), valueRemovedStore, version); err != nil {
			return fmt.Errorf("failed to exec SQL statement: %w", err)
		}
	}

	return db.storage.Commit()
}

func (db *Database) PrintRowsDebug() {
	stmt, err := db.storage.Prepare("SELECT store_key, key, value, version, tombstone FROM state_storage")
	if err != nil {
		panic(fmt.Errorf("failed to prepare SQL statement: %w", err))
	}

	defer stmt.Close()

	err = stmt.Exec()
	if err != nil {
		panic(fmt.Errorf("failed to execute SQL query: %w", err))
	}

	var (
		sb strings.Builder
	)
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			panic(fmt.Errorf("failed to step through SQL rows: %w", err))
		}
		if !hasRow {
			break
		}
		var (
			storeKey []byte
			key      []byte
			value    []byte
			version  int64
			tomb     int64
		)
		if err := stmt.Scan(&storeKey, &key, &value, &version, &tomb); err != nil {
			panic(fmt.Sprintf("failed to scan row: %s", err))
		}

		sb.WriteString(fmt.Sprintf("STORE_KEY: %s, KEY: %s, VALUE: %s, VERSION: %d, TOMBSTONE: %d\n", storeKey, key, value, version, tomb))
	}

	fmt.Println(strings.TrimSpace(sb.String()))
}

func getPruneHeight(storage *sqlite3.Conn) (height uint64, err error) {
	stmt, err := storage.Prepare(`SELECT value FROM state_storage WHERE store_key = ? AND key = ?`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer func(stmt *sqlite3.Stmt) {
		cErr := stmt.Close()
		if cErr != nil {
			err = fmt.Errorf("failed to close SQL statement: %w", cErr)
		}
	}(stmt)

	if err = stmt.Bind(reservedStoreKey, keyPruneHeight); err != nil {
		return 0, fmt.Errorf("failed to bind prune height SQL statement: %w", err)
	}
	hasRows, err := stmt.Step()
	if err != nil {
		return 0, fmt.Errorf("failed to step prune height SQL statement: %w", err)
	}
	if !hasRows {
		return 0, nil
	}
	var h int64
	if err = stmt.Scan(&h); err != nil {
		return 0, fmt.Errorf("failed to scan prune height SQL statement: %w", err)
	}
	height = uint64(h)
	return height, nil
}
