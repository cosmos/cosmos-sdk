//go:build rocksdb
// +build rocksdb

package db

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/linxGnu/grocksdb"

	corestore "cosmossdk.io/core/store"
	storeerrors "cosmossdk.io/store/v2/errors"
)

var (
	_ corestore.KVStoreWithBatch = (*RocksDB)(nil)

	defaultReadOpts = grocksdb.NewDefaultReadOptions()
)

// RocksDB implements `corestore.KVStoreWithBatch`, using RocksDB as the underlying storage engine.
// It is only used for store v2 migration, since some clients use RocksDB as
// the IAVL v0/v1 backend.
type RocksDB struct {
	storage *grocksdb.DB
}

// defaultRocksdbOptions is good enough for most cases, including heavy workloads.
// 1GB table cache, 512MB write buffer(may use 50% more on heavy workloads).
// compression: snappy as default, use `-lsnappy` flag to enable.
func defaultRocksdbOptions() *grocksdb.Options {
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(1 << 30))
	bbto.SetFilterPolicy(grocksdb.NewBloomFilter(10))

	rocksdbOpts := grocksdb.NewDefaultOptions()
	rocksdbOpts.SetBlockBasedTableFactory(bbto)
	// SetMaxOpenFiles to 4096 seems to provide a reliable performance boost
	rocksdbOpts.SetMaxOpenFiles(4096)
	rocksdbOpts.SetCreateIfMissing(true)
	rocksdbOpts.IncreaseParallelism(runtime.NumCPU())
	// 1.5GB maximum memory use for writebuffer.
	rocksdbOpts.OptimizeLevelStyleCompaction(512 * 1024 * 1024)
	return rocksdbOpts
}

func NewRocksDB(name, dataDir string) (*RocksDB, error) {
	opts := defaultRocksdbOptions()
	opts.SetCreateIfMissing(true)

	return NewRocksDBWithOpts(name, dataDir, opts)
}

func NewRocksDBWithOpts(name, dataDir string, opts *grocksdb.Options) (*RocksDB, error) {
	dbPath := filepath.Join(dataDir, name+DBFileSuffix)
	storage, err := grocksdb.OpenDb(opts, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open RocksDB: %w", err)
	}

	return &RocksDB{
		storage: storage,
	}, nil
}

func (db *RocksDB) Close() error {
	db.storage.Close()
	db.storage = nil
	return nil
}

func (db *RocksDB) Get(key []byte) ([]byte, error) {
	bz, err := db.storage.GetBytes(defaultReadOpts, key)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

func (db *RocksDB) Has(key []byte) (bool, error) {
	bz, err := db.Get(key)
	if err != nil {
		return false, err
	}

	return bz != nil, nil
}

func (db *RocksDB) Set(key, value []byte) error {
	if len(key) == 0 {
		return storeerrors.ErrKeyEmpty
	}
	if value == nil {
		return storeerrors.ErrValueNil
	}

	return db.storage.Put(grocksdb.NewDefaultWriteOptions(), key, value)
}

func (db *RocksDB) Delete(key []byte) error {
	if len(key) == 0 {
		return storeerrors.ErrKeyEmpty
	}

	return db.storage.Delete(grocksdb.NewDefaultWriteOptions(), key)
}

func (db *RocksDB) Iterator(start, end []byte) (corestore.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, storeerrors.ErrKeyEmpty
	}

	itr := db.storage.NewIterator(defaultReadOpts)
	return newRocksDBIterator(itr, start, end, false), nil
}

func (db *RocksDB) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, storeerrors.ErrKeyEmpty
	}

	itr := db.storage.NewIterator(defaultReadOpts)
	return newRocksDBIterator(itr, start, end, true), nil
}

func (db *RocksDB) NewBatch() corestore.Batch {
	return &rocksDBBatch{
		db:    db,
		batch: grocksdb.NewWriteBatch(),
	}
}

func (db *RocksDB) NewBatchWithSize(_ int) corestore.Batch {
	return db.NewBatch()
}

var _ corestore.Iterator = (*rocksDBIterator)(nil)

type rocksDBIterator struct {
	source  *grocksdb.Iterator
	start   []byte
	end     []byte
	valid   bool
	reverse bool
}

func newRocksDBIterator(src *grocksdb.Iterator, start, end []byte, reverse bool) *rocksDBIterator {
	if reverse {
		if end == nil {
			src.SeekToLast()
		} else {
			src.Seek(end)

			if src.Valid() {
				eoaKey := readOnlySlice(src.Key()) // end or after key
				if bytes.Compare(end, eoaKey) <= 0 {
					src.Prev()
				}
			} else {
				src.SeekToLast()
			}
		}
	} else {
		if start == nil {
			src.SeekToFirst()
		} else {
			src.Seek(start)
		}
	}

	return &rocksDBIterator{
		source:  src,
		start:   start,
		end:     end,
		reverse: reverse,
		valid:   src.Valid(),
	}
}

func (itr *rocksDBIterator) Domain() (start, end []byte) {
	return itr.start, itr.end
}

func (itr *rocksDBIterator) Valid() bool {
	// once invalid, forever invalid
	if !itr.valid {
		return false
	}

	// if source has error, consider it invalid
	if err := itr.source.Err(); err != nil {
		itr.valid = false
		return false
	}

	// if source is invalid, consider it invalid
	if !itr.source.Valid() {
		itr.valid = false
		return false
	}

	// if key is at the end or past it, consider it invalid
	start := itr.start
	end := itr.end
	key := readOnlySlice(itr.source.Key())

	if itr.reverse {
		if start != nil && bytes.Compare(key, start) < 0 {
			itr.valid = false
			return false
		}
	} else {
		if end != nil && bytes.Compare(end, key) <= 0 {
			itr.valid = false
			return false
		}
	}

	return true
}

func (itr *rocksDBIterator) Key() []byte {
	itr.assertIsValid()
	return copyAndFreeSlice(itr.source.Key())
}

func (itr *rocksDBIterator) Value() []byte {
	itr.assertIsValid()
	return copyAndFreeSlice(itr.source.Value())
}

func (itr *rocksDBIterator) Next() {
	if !itr.valid {
		return
	}

	if itr.reverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

func (itr *rocksDBIterator) Error() error {
	return itr.source.Err()
}

func (itr *rocksDBIterator) Close() error {
	itr.source.Close()
	itr.source = nil
	itr.valid = false

	return nil
}

func (itr *rocksDBIterator) assertIsValid() {
	if !itr.valid {
		panic("rocksDB iterator is invalid")
	}
}

type rocksDBBatch struct {
	db    *RocksDB
	batch *grocksdb.WriteBatch
}

func (b *rocksDBBatch) Set(key, value []byte) error {
	if len(key) == 0 {
		return storeerrors.ErrKeyEmpty
	}
	if value == nil {
		return storeerrors.ErrValueNil
	}
	if b.batch == nil {
		return storeerrors.ErrBatchClosed
	}

	b.batch.Put(key, value)
	return nil
}

func (b *rocksDBBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return storeerrors.ErrKeyEmpty
	}
	if b.batch == nil {
		return storeerrors.ErrBatchClosed
	}

	b.batch.Delete(key)
	return nil
}

func (b *rocksDBBatch) Write() error {
	writeOpts := grocksdb.NewDefaultWriteOptions()
	writeOpts.SetSync(false)

	if err := b.db.storage.Write(writeOpts, b.batch); err != nil {
		return fmt.Errorf("failed to write RocksDB batch: %w", err)
	}

	return nil
}

func (b *rocksDBBatch) WriteSync() error {
	writeOpts := grocksdb.NewDefaultWriteOptions()
	writeOpts.SetSync(true)

	if err := b.db.storage.Write(writeOpts, b.batch); err != nil {
		return fmt.Errorf("failed to write RocksDB batch: %w", err)
	}

	return nil
}

func (b *rocksDBBatch) Close() error {
	b.batch.Destroy()
	return nil
}

func (b *rocksDBBatch) GetByteSize() (int, error) {
	return len(b.batch.Data()), nil
}

func readOnlySlice(s *grocksdb.Slice) []byte {
	if !s.Exists() {
		return nil
	}

	return s.Data()
}

// copyAndFreeSlice will copy a given RocksDB slice and free it. If the slice
// does not exist, <nil> will be returned.
func copyAndFreeSlice(s *grocksdb.Slice) []byte {
	defer s.Free()

	if !s.Exists() {
		return nil
	}

	return slices.Clone(s.Data())
}
