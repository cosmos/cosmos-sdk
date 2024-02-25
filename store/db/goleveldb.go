package db

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/spf13/cast"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

// GoLevelDB implements RawDB using github.com/syndtr/goleveldb/leveldb.
// It is used for only store v2 migration, since some clients use goleveldb as the iavl v0/v1 backend.
type GoLevelDB struct {
	db *leveldb.DB
}

var _ store.RawDB = (*GoLevelDB)(nil)

func NewGoLevelDB(name, dir string, opts store.DBOptions) (*GoLevelDB, error) {
	defaultOpts := &opt.Options{
		Filter: filter.NewBloomFilter(10), // by default, goleveldb doesn't use a bloom filter.
	}
	if opts != nil {
		files := cast.ToInt(opts.Get("maxopenfiles"))
		if files > 0 {
			defaultOpts.OpenFilesCacheCapacity = files
		}
	}

	return NewGoLevelDBWithOpts(name, dir, defaultOpts)
}

func NewGoLevelDBWithOpts(name, dir string, o *opt.Options) (*GoLevelDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	db, err := leveldb.OpenFile(dbPath, o)
	if err != nil {
		return nil, err
	}
	database := &GoLevelDB{
		db: db,
	}
	return database, nil
}

// Get implements RawDB.
func (db *GoLevelDB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, store.ErrKeyEmpty
	}
	res, err := db.db.Get(key, nil)
	if err != nil {
		if err == errors.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return res, nil
}

// Has implements RawDB.
func (db *GoLevelDB) Has(key []byte) (bool, error) {
	bytes, err := db.Get(key)
	if err != nil {
		return false, err
	}
	return bytes != nil, nil
}

// Set implements RawDB.
func (db *GoLevelDB) Set(key, value []byte) error {
	if len(key) == 0 {
		return store.ErrKeyEmpty
	}
	if value == nil {
		return store.ErrValueNil
	}
	if err := db.db.Put(key, value, nil); err != nil {
		return err
	}
	return nil
}

// SetSync implements RawDB.
func (db *GoLevelDB) SetSync(key, value []byte) error {
	if len(key) == 0 {
		return store.ErrKeyEmpty
	}
	if value == nil {
		return store.ErrValueNil
	}
	if err := db.db.Put(key, value, &opt.WriteOptions{Sync: true}); err != nil {
		return err
	}
	return nil
}

// Delete implements RawDB.
func (db *GoLevelDB) Delete(key []byte) error {
	if len(key) == 0 {
		return store.ErrKeyEmpty
	}
	if err := db.db.Delete(key, nil); err != nil {
		return err
	}
	return nil
}

// DeleteSync implements RawDB.
func (db *GoLevelDB) DeleteSync(key []byte) error {
	if len(key) == 0 {
		return store.ErrKeyEmpty
	}
	err := db.db.Delete(key, &opt.WriteOptions{Sync: true})
	if err != nil {
		return err
	}
	return nil
}

func (db *GoLevelDB) RawDB() *leveldb.DB {
	return db.db
}

// Close implements RawDB.
func (db *GoLevelDB) Close() error {
	if err := db.db.Close(); err != nil {
		return err
	}
	return nil
}

// Print implements RawDB.
func (db *GoLevelDB) Print() error {
	str, err := db.db.GetProperty("leveldb.stats")
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", str)

	itr := db.db.NewIterator(nil, nil)
	for itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}
	return nil
}

// Stats implements RawDB.
func (db *GoLevelDB) Stats() map[string]string {
	keys := []string{
		"leveldb.num-files-at-level{n}",
		"leveldb.stats",
		"leveldb.sstables",
		"leveldb.blockpool",
		"leveldb.cachedblock",
		"leveldb.openedtables",
		"leveldb.alivesnaps",
		"leveldb.aliveiters",
	}

	stats := make(map[string]string)
	for _, key := range keys {
		str, err := db.db.GetProperty(key)
		if err == nil {
			stats[key] = str
		}
	}
	return stats
}

func (db *GoLevelDB) ForceCompact(start, limit []byte) error {
	return db.db.CompactRange(util.Range{Start: start, Limit: limit})
}

// NewBatch implements RawDB.
func (db *GoLevelDB) NewBatch() store.RawBatch {
	return newGoLevelDBBatch(db)
}

// NewBatchWithSize implements RawDB.
func (db *GoLevelDB) NewBatchWithSize(size int) store.RawBatch {
	return newGoLevelDBBatchWithSize(db, size)
}

// Iterator implements RawDB.
func (db *GoLevelDB) Iterator(start, end []byte) (corestore.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}
	itr := db.db.NewIterator(&util.Range{Start: start, Limit: end}, nil)
	return newGoLevelDBIterator(itr, start, end, false), nil
}

// ReverseIterator implements RawDB.
func (db *GoLevelDB) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, store.ErrKeyEmpty
	}
	itr := db.db.NewIterator(&util.Range{Start: start, Limit: end}, nil)
	return newGoLevelDBIterator(itr, start, end, true), nil
}

type goLevelDBIterator struct {
	source    iterator.Iterator
	start     []byte
	end       []byte
	isReverse bool
	isInvalid bool
}

var _ corestore.Iterator = (*goLevelDBIterator)(nil)

func newGoLevelDBIterator(source iterator.Iterator, start, end []byte, isReverse bool) *goLevelDBIterator {
	if isReverse {
		if end == nil {
			source.Last()
		} else {
			valid := source.Seek(end)
			if valid {
				eoakey := source.Key() // end or after key
				if bytes.Compare(end, eoakey) <= 0 {
					source.Prev()
				}
			} else {
				source.Last()
			}
		}
	} else {
		if start == nil {
			source.First()
		} else {
			source.Seek(start)
		}
	}
	return &goLevelDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
	}
}

// Domain implements Iterator.
func (itr *goLevelDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Valid implements Iterator.
func (itr *goLevelDBIterator) Valid() bool {
	// Once invalid, forever invalid.
	if itr.isInvalid {
		return false
	}

	// If source errors, invalid.
	if err := itr.Error(); err != nil {
		itr.isInvalid = true
		return false
	}

	// If source is invalid, invalid.
	if !itr.source.Valid() {
		itr.isInvalid = true
		return false
	}

	// If key is end or past it, invalid.
	start := itr.start
	end := itr.end
	key := itr.source.Key()

	if itr.isReverse {
		if start != nil && bytes.Compare(key, start) < 0 {
			itr.isInvalid = true
			return false
		}
	} else {
		if end != nil && bytes.Compare(end, key) <= 0 {
			itr.isInvalid = true
			return false
		}
	}

	// Valid
	return true
}

// Key implements Iterator.
func (itr *goLevelDBIterator) Key() []byte {
	// Key returns a copy of the current key.
	// See https://github.com/syndtr/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	itr.assertIsValid()
	return cp(itr.source.Key())
}

// Value implements Iterator.
func (itr *goLevelDBIterator) Value() []byte {
	// Value returns a copy of the current value.
	// See https://github.com/syndtr/goleveldb/blob/52c212e6c196a1404ea59592d3f1c227c9f034b2/leveldb/iterator/iter.go#L88
	itr.assertIsValid()
	return cp(itr.source.Value())
}

// Next implements Iterator.
func (itr *goLevelDBIterator) Next() {
	itr.assertIsValid()
	if itr.isReverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

// Error implements Iterator.
func (itr *goLevelDBIterator) Error() error {
	return itr.source.Error()
}

// Close implements Iterator.
func (itr *goLevelDBIterator) Close() error {
	itr.source.Release()
	return nil
}

func (itr goLevelDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("iterator is invalid")
	}
}

type goLevelDBBatch struct {
	db    *GoLevelDB
	batch *leveldb.Batch
}

var _ store.RawBatch = (*goLevelDBBatch)(nil)

func newGoLevelDBBatch(db *GoLevelDB) *goLevelDBBatch {
	return &goLevelDBBatch{
		db:    db,
		batch: new(leveldb.Batch),
	}
}

func newGoLevelDBBatchWithSize(db *GoLevelDB, size int) *goLevelDBBatch {
	return &goLevelDBBatch{
		db:    db,
		batch: leveldb.MakeBatch(size),
	}
}

// Set implements RawBatch.
func (b *goLevelDBBatch) Set(key, value []byte) error {
	if len(key) == 0 {
		return store.ErrKeyEmpty
	}
	if value == nil {
		return store.ErrValueNil
	}
	if b.batch == nil {
		return store.ErrBatchClosed
	}
	b.batch.Put(key, value)
	return nil
}

// Delete implements RawBatch.
func (b *goLevelDBBatch) Delete(key []byte) error {
	if len(key) == 0 {
		return store.ErrKeyEmpty
	}
	if b.batch == nil {
		return store.ErrBatchClosed
	}
	b.batch.Delete(key)
	return nil
}

// Write implements RawBatch.
func (b *goLevelDBBatch) Write() error {
	return b.write(false)
}

// WriteSync implements RawBatch.
func (b *goLevelDBBatch) WriteSync() error {
	return b.write(true)
}

func (b *goLevelDBBatch) write(sync bool) error {
	if b.batch == nil {
		return store.ErrBatchClosed
	}
	err := b.db.db.Write(b.batch, &opt.WriteOptions{Sync: sync})
	if err != nil {
		return err
	}
	// Make sure batch cannot be used afterwards. Callers should still call Close(), for errors.
	return b.Close()
}

// Close implements RawBatch.
func (b *goLevelDBBatch) Close() error {
	if b.batch != nil {
		b.batch.Reset()
		b.batch = nil
	}
	return nil
}

// GetByteSize implements RawBatch
func (b *goLevelDBBatch) GetByteSize() (int, error) {
	if b.batch == nil {
		return 0, store.ErrBatchClosed
	}
	return len(b.batch.Dump()), nil
}
