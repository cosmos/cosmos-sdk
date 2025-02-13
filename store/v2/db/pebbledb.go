package db

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/cockroachdb/pebble"
	"github.com/spf13/cast"

	coreserver "cosmossdk.io/core/server"
	corestore "cosmossdk.io/core/store"
	storeerrors "cosmossdk.io/store/v2/errors"
)

var _ corestore.KVStoreWithBatch = (*PebbleDB)(nil)

// PebbleDB implements `corestore.KVStoreWithBatch` using PebbleDB as the underlying storage engine.
// It is used for only store v2 migration, since some clients use PebbleDB as
// the IAVL v0/v1 backend.
type PebbleDB struct {
	storage *pebble.DB
}

func NewPebbleDB(name, dataDir string) (*PebbleDB, error) {
	return NewPebbleDBWithOpts(name, dataDir, nil)
}

func NewPebbleDBWithOpts(name, dataDir string, opts coreserver.DynamicConfig) (*PebbleDB, error) {
	do := &pebble.Options{
		Logger:                   &fatalLogger{},          // pebble info logs are messing up the logs (not a cosmossdk.io/log logger)
		MaxConcurrentCompactions: func() int { return 3 }, // default 1
	}

	do.EnsureDefaults()

	if opts != nil {
		files := cast.ToInt(opts.Get("maxopenfiles"))
		if files > 0 {
			do.MaxOpenFiles = files
		}
	}
	dbPath := filepath.Join(dataDir, name+DBFileSuffix)
	db, err := pebble.Open(dbPath, do)
	if err != nil {
		return nil, fmt.Errorf("failed to open PebbleDB: %w", err)
	}

	return &PebbleDB{storage: db}, nil
}

func (db *PebbleDB) Close() error {
	err := db.storage.Close()
	db.storage = nil
	return err
}

func (db *PebbleDB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, storeerrors.ErrKeyEmpty
	}

	bz, closer, err := db.storage.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			// in case of a fresh database
			return nil, nil
		}

		return nil, fmt.Errorf("failed to perform PebbleDB read: %w", err)
	}

	return slices.Clone(bz), closer.Close()
}

func (db *PebbleDB) Has(key []byte) (bool, error) {
	bz, err := db.Get(key)
	if err != nil {
		return false, err
	}

	return bz != nil, nil
}

func (db *PebbleDB) Set(key, value []byte) error {
	if len(key) == 0 {
		return storeerrors.ErrKeyEmpty
	}
	if value == nil {
		return storeerrors.ErrValueNil
	}

	return db.storage.Set(key, value, &pebble.WriteOptions{Sync: false})
}

func (db *PebbleDB) Delete(key []byte) error {
	if len(key) == 0 {
		return storeerrors.ErrKeyEmpty
	}

	return db.storage.Delete(key, &pebble.WriteOptions{Sync: false})
}

func (db *PebbleDB) Iterator(start, end []byte) (corestore.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, storeerrors.ErrKeyEmpty
	}

	itr, err := db.storage.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	if err != nil {
		return nil, fmt.Errorf("failed to create PebbleDB iterator: %w", err)
	}

	return newPebbleDBIterator(itr, start, end, false), nil
}

func (db *PebbleDB) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, storeerrors.ErrKeyEmpty
	}

	itr, err := db.storage.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	if err != nil {
		return nil, fmt.Errorf("failed to create PebbleDB iterator: %w", err)
	}

	return newPebbleDBIterator(itr, start, end, true), nil
}

func (db *PebbleDB) NewBatch() corestore.Batch {
	return &pebbleDBBatch{
		db:    db,
		batch: db.storage.NewBatch(),
	}
}

func (db *PebbleDB) NewBatchWithSize(size int) corestore.Batch {
	return &pebbleDBBatch{
		db:    db,
		batch: db.storage.NewBatchWithSize(size),
	}
}

var _ corestore.Iterator = (*pebbleDBIterator)(nil)

type pebbleDBIterator struct {
	source  *pebble.Iterator
	start   []byte
	end     []byte
	valid   bool
	reverse bool
}

func newPebbleDBIterator(src *pebble.Iterator, start, end []byte, reverse bool) *pebbleDBIterator {
	// move the underlying PebbleDB cursor to the first key
	var valid bool
	if reverse {
		if end == nil {
			valid = src.Last()
		} else {
			valid = src.SeekLT(end)
		}
	} else {
		valid = src.First()
	}

	return &pebbleDBIterator{
		source:  src,
		start:   start,
		end:     end,
		valid:   valid,
		reverse: reverse,
	}
}

func (itr *pebbleDBIterator) Domain() (start, end []byte) {
	return itr.start, itr.end
}

func (itr *pebbleDBIterator) Valid() bool {
	// once invalid, forever invalid
	if !itr.valid || !itr.source.Valid() {
		itr.valid = false
		return itr.valid
	}

	// if source has error, consider it invalid
	if err := itr.source.Error(); err != nil {
		itr.valid = false
		return itr.valid
	}

	// if key is at the end or past it, consider it invalid
	if end := itr.end; end != nil {
		if bytes.Compare(end, itr.Key()) <= 0 {
			itr.valid = false
			return itr.valid
		}
	}

	return true
}

func (itr *pebbleDBIterator) Key() []byte {
	itr.assertIsValid()
	return slices.Clone(itr.source.Key())
}

func (itr *pebbleDBIterator) Value() []byte {
	itr.assertIsValid()
	return slices.Clone(itr.source.Value())
}

func (itr *pebbleDBIterator) Next() {
	itr.assertIsValid()

	if itr.reverse {
		itr.valid = itr.source.Prev()
	} else {
		itr.valid = itr.source.Next()
	}
}

func (itr *pebbleDBIterator) Error() error {
	return itr.source.Error()
}

func (itr *pebbleDBIterator) Close() error {
	if itr.source == nil {
		return nil
	}
	err := itr.source.Close()
	itr.source = nil
	itr.valid = false

	return err
}

func (itr *pebbleDBIterator) assertIsValid() {
	if !itr.valid {
		panic("pebbleDB iterator is invalid")
	}
}

var _ corestore.Batch = (*pebbleDBBatch)(nil)

type pebbleDBBatch struct {
	db    *PebbleDB
	batch *pebble.Batch
}

func (b *pebbleDBBatch) Set(key, value []byte) error {
	if len(key) == 0 {
		return storeerrors.ErrKeyEmpty
	}
	if value == nil {
		return storeerrors.ErrValueNil
	}
	if b.batch == nil {
		return storeerrors.ErrBatchClosed
	}

	return b.batch.Set(key, value, nil)
}

func (b *pebbleDBBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return storeerrors.ErrKeyEmpty
	}
	if b.batch == nil {
		return storeerrors.ErrBatchClosed
	}

	return b.batch.Delete(key, nil)
}

func (b *pebbleDBBatch) Write() error {
	err := b.batch.Commit(&pebble.WriteOptions{Sync: false})
	if err != nil {
		return fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}

	return nil
}

func (b *pebbleDBBatch) WriteSync() error {
	err := b.batch.Commit(&pebble.WriteOptions{Sync: true})
	if err != nil {
		return fmt.Errorf("failed to write PebbleDB batch: %w", err)
	}

	return nil
}

func (b *pebbleDBBatch) Close() error {
	return b.batch.Close()
}

func (b *pebbleDBBatch) GetByteSize() (int, error) {
	return b.batch.Len(), nil
}

type fatalLogger struct {
	pebble.Logger
}

func (*fatalLogger) Fatalf(format string, args ...interface{}) {
	pebble.DefaultLogger.Fatalf(format, args...)
}

func (*fatalLogger) Infof(format string, args ...interface{}) {}
