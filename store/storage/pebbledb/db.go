package pebbledb

import (
	"encoding/binary"
	"errors"
	"fmt"

	"cosmossdk.io/store/v2"
	"github.com/cockroachdb/pebble"
)

const (
	VersionSize = 8

	StorePrefixTpl   = "s/k:%d/%s/" // s/k:<version>/<storeKey>/<key>
	latestVersionKey = "s/latest"
)

var (
	_ store.VersionedDatabase = (*Database)(nil)

	defaultWriteOpts = pebble.Sync
)

type Database struct {
	storage *pebble.DB
}

func New(dataDir string) (*Database, error) {
	opts := &pebble.Options{}
	opts = opts.EnsureDefaults()

	db, err := pebble.Open(dataDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open PebbleDB: %w", err)
	}

	return &Database{
		storage: db,
	}, nil
}

func NewWithDB(storage *pebble.DB) *Database {
	return &Database{
		storage: storage,
	}
}

func (db *Database) Close() error {
	return db.storage.Close()
}

func (db *Database) SetLatestVersion(version uint64) error {
	var ts [VersionSize]byte
	binary.LittleEndian.PutUint64(ts[:], version)
	return db.storage.Set([]byte(latestVersionKey), ts[:], defaultWriteOpts)
}

func (db *Database) GetLatestVersion() (uint64, error) {
	bz, closer, err := db.storage.Get([]byte(latestVersionKey))
	if err != nil {
		return 0, err
	}

	if len(bz) == 0 {
		return 0, closer.Close()
	}

	return binary.LittleEndian.Uint64(bz), closer.Close()
}

func (db *Database) Has(storeKey string, version uint64, key []byte) (bool, error) {
	_, closer, err := db.storage.Get(prependStoreKey(storeKey, version, key))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, closer.Close()
}

func storePrefix(storeKey string, version uint64) []byte {
	return []byte(fmt.Sprintf(StorePrefixTpl, version, storeKey))
}

func prependStoreKey(storeKey string, version uint64, key []byte) []byte {
	return append(storePrefix(storeKey, version), key...)
}
