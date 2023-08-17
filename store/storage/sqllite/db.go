package sqllite

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"cosmossdk.io/store/v2"
	_ "modernc.org/sqlite"
)

const (
	driverName = "sqlite"
	dbName     = "ss.db"
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

	sqlStmt := `
	create table if not exists state_storage (
		id integer not null primary key, 
		store_key varchar not null,
		key varchar not null,
		value varchar not null,
		height integer unsigned not null
	);
	create unique index if not exists idx_store_key_height on state_storage (store_key, key, height);
	`
	_, err = db.Exec(sqlStmt)
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
	panic("not implemented!")
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	panic("not implemented!")
}

func (db *Database) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	panic("not implemented!")
}

func (db *Database) Set(storeKey string, version uint64, key, value []byte) error {
	panic("not implemented!")
}

func (db *Database) Delete(storeKey string, version uint64, key []byte) error {
	panic("not implemented!")
}

func (db *Database) NewBatch(version uint64) (store.Batch, error) {
	panic("not implemented!")
}

func (db *Database) NewIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	panic("not implemented!")
}

func (db *Database) NewReverseIterator(storeKey string, version uint64, start, end []byte) (store.Iterator, error) {
	panic("not implemented!")
}
