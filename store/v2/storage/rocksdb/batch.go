//go:build rocksdb
// +build rocksdb

package rocksdb

import (
	"encoding/binary"

	"github.com/linxGnu/grocksdb"

	"cosmossdk.io/store/v2"
)

var _ store.Batch = (*Batch)(nil)

type Batch struct {
	version  uint64
	ts       [TimestampSize]byte
	storage  *grocksdb.DB
	cfHandle *grocksdb.ColumnFamilyHandle
	batch    *grocksdb.WriteBatch
}

// NewBatch creates a new versioned batch used for batch writes. The caller
// must ensure to call Write() on the returned batch to commit the changes and to
// destroy the batch when done.
func NewBatch(db *Database, version uint64) Batch {
	var ts [TimestampSize]byte
	binary.LittleEndian.PutUint64(ts[:], version)

	batch := grocksdb.NewWriteBatch()
	batch.Put([]byte(latestVersionKey), ts[:])

	return Batch{
		version:  version,
		ts:       ts,
		storage:  db.storage,
		cfHandle: db.cfHandle,
		batch:    batch,
	}
}

func (b Batch) Size() int {
	return len(b.batch.Data())
}

func (b Batch) Reset() error {
	b.batch.Clear()
	return nil
}

func (b Batch) Set(storeKey, key, value []byte) error {
	prefixedKey := prependStoreKey(storeKey, key)
	b.batch.PutCFWithTS(b.cfHandle, prefixedKey, b.ts[:], value)
	return nil
}

func (b Batch) Delete(storeKey, key []byte) error {
	prefixedKey := prependStoreKey(storeKey, key)
	b.batch.DeleteCFWithTS(b.cfHandle, prefixedKey, b.ts[:])
	return nil
}

func (b Batch) Write() error {
	defer b.batch.Destroy()
	return b.storage.Write(defaultWriteOpts, b.batch)
}
