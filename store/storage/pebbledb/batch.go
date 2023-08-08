package pebbledb

import (
	"encoding/binary"
	"errors"
	"fmt"

	"cosmossdk.io/store/v2"
	"github.com/cockroachdb/pebble"
)

var _ store.Batch = (*Batch)(nil)

type Batch struct {
	storage *pebble.DB
	batch   *pebble.Batch
	version uint64
}

func NewBatch(storage *pebble.DB, version uint64) (*Batch, error) {
	var versionBz [VersionSize]byte
	binary.LittleEndian.PutUint64(versionBz[:], uint64(version))

	batch := storage.NewBatch()

	if err := batch.Set([]byte(latestVersionKey), versionBz[:], nil); err != nil {
		return nil, fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}

	return &Batch{
		storage: storage,
		batch:   batch,
		version: version,
	}, nil
}

func (b *Batch) Size() int {
	return b.batch.Len()
}

func (b *Batch) Reset() {
	b.batch.Reset()
}

func (b *Batch) Set(storeKey string, key, value []byte) error {
	prefixedKey := MVCCEncode(prependStoreKey(storeKey, key), b.version)
	return b.batch.Set(prefixedKey, value, nil)
}

func (b *Batch) Delete(storeKey string, key []byte) error {
	prefixedKey := MVCCEncode(prependStoreKey(storeKey, key), b.version)
	return b.batch.Delete(prefixedKey, nil)
}

func (b *Batch) Write() (err error) {
	defer func() {
		err = errors.Join(err, b.batch.Close())
	}()

	return b.batch.Commit(defaultWriteOpts)
}
