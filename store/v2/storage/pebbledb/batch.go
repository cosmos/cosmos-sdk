package pebbledb

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/cockroachdb/pebble"

	"cosmossdk.io/store/v2"
)

var _ store.Batch = (*Batch)(nil)

type Batch struct {
	storage *pebble.DB
	batch   *pebble.Batch
	version uint64
	sync    bool
}

func NewBatch(storage *pebble.DB, version uint64, sync bool) (*Batch, error) {
	var versionBz [VersionSize]byte
	binary.LittleEndian.PutUint64(versionBz[:], version)

	batch := storage.NewBatch()

	if err := batch.Set([]byte(latestVersionKey), versionBz[:], nil); err != nil {
		return nil, fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}

	return &Batch{
		storage: storage,
		batch:   batch,
		version: version,
		sync:    sync,
	}, nil
}

func (b *Batch) Size() int {
	return b.batch.Len()
}

func (b *Batch) Reset() error {
	b.batch.Reset()
	return nil
}

func (b *Batch) set(storeKey []byte, tombstone uint64, key, value []byte) error {
	prefixedKey := MVCCEncode(prependStoreKey(storeKey, key), b.version)
	prefixedVal := MVCCEncode(value, tombstone)

	if err := b.batch.Set(prefixedKey, prefixedVal, nil); err != nil {
		return fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}

	return nil
}

func (b *Batch) Set(storeKey, key, value []byte) error {
	return b.set(storeKey, 0, key, value)
}

func (b *Batch) Delete(storeKey, key []byte) error {
	return b.set(storeKey, b.version, key, []byte(tombstoneVal))
}

func (b *Batch) Write() (err error) {
	defer func() {
		err = errors.Join(err, b.batch.Close())
	}()

	return b.batch.Commit(&pebble.WriteOptions{Sync: b.sync})
}
