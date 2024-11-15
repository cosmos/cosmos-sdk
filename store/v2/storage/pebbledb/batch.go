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
	size    int
}

const (
	oneIf64Bit     = ^uint(0) >> 63
	maxUint32OrInt = (1<<31)<<oneIf64Bit - 1
	maxVarintLen32 = 5
)

func keyValueSize(key, value []byte) int {
	return len(key) + len(value) + 1 + 2*maxVarintLen32
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
		size:    keyValueSize([]byte(latestVersionKey), versionBz[:]),
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

	size := keyValueSize(prefixedKey, prefixedVal)
	if b.size+size > maxUint32OrInt {
		// 4 GB is huge, probably genesis; flush and reset
		if err := b.batch.Commit(&pebble.WriteOptions{Sync: b.sync}); err != nil {
			return fmt.Errorf("max batch size exceed: failed to write PebbleDB batch: %w", err)
		}
		b.batch.Reset()
		b.size = 0
	}

	if err := b.batch.Set(prefixedKey, prefixedVal, nil); err != nil {
		return fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}
	b.size += size

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
