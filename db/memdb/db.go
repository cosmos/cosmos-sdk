package memdb

import (
	"bytes"
	"fmt"
	"sync"

	tmdb "github.com/cosmos/cosmos-sdk/db"
	"github.com/google/btree"
)

const (
	// The approximate number of items and children per B-tree node. Tuned with benchmarks.
	bTreeDegree = 32
)

// item is a btree.Item with byte slices as keys and values
type item struct {
	key   []byte
	value []byte
}

// Less implements btree.Item.
func (i *item) Less(other btree.Item) bool {
	// this considers nil == []byte{}, but that's ok since we handle nil endpoints
	// in iterators specially anyway
	return bytes.Compare(i.key, other.(*item).key) == -1
}

// newKey creates a new key item.
func newKey(key []byte) *item {
	return &item{key: key}
}

// newPair creates a new pair item.
func newPair(key, value []byte) *item {
	return &item{key: key, value: value}
}

// MemDB is an in-memory database backend using a B-tree for storage.
//
// For performance reasons, all given and returned keys and values are pointers to the in-memory
// database, so modifying them will cause the stored values to be modified as well. All DB methods
// already specify that keys and values should be considered read-only, but this is especially
// important with MemDB.
type MemDB struct {
	dbVersion
	saved []*btree.BTree
}

type dbVersion struct {
	mtx   sync.RWMutex
	btree *btree.BTree
}

var _ tmdb.DB = (*MemDB)(nil)
var _ tmdb.DBReader = (*dbVersion)(nil)
var _ tmdb.DBReadWriter = (*dbVersion)(nil)

// NewDB creates a new in-memory database.
func NewDB() *MemDB {
	return &MemDB{dbVersion: dbVersion{btree: btree.New(bTreeDegree)}}
}

// Close implements DB.
func (db *MemDB) Close() error {
	// Close is a noop since for an in-memory database, we don't have a destination to flush
	// contents to nor do we want any data loss on invoking Close().
	// See the discussion in https://github.com/tendermint/tendermint/libs/pull/56
	return nil
}

func (db *MemDB) InitialVersion() uint64 {
	return 1
}

func (db *MemDB) Versions() []uint64 {
	var ret []uint64
	for i := range db.saved {
		ret = append(ret, uint64(i+1))
	}
	return ret
}

func (db *MemDB) CurrentVersion() uint64 {
	return uint64(len(db.saved)) + db.InitialVersion()
}

func (db *MemDB) ReaderAt(version uint64) tmdb.DBReader {
	// allows AtVersion(current), desired? todo
	if version == db.CurrentVersion() {
		return &db.dbVersion
	}
	version -= db.InitialVersion()
	if version >= uint64(len(db.saved)) {
		return nil
	}
	return &dbVersion{btree: db.saved[version]}
}

func (db *MemDB) ReadWriter() tmdb.DBReadWriter {
	return &db.dbVersion
}

func (db *MemDB) SaveVersion() uint64 {
	db.dbVersion.mtx.Lock()
	defer db.dbVersion.mtx.Unlock()
	id := db.CurrentVersion()
	db.saved = append(db.saved, db.btree)
	// BTree's Clone() makes a CoW cloned ref of the current data
	db.dbVersion.btree = db.btree.Clone()
	return id
}

// Get implements DB.
func (db *dbVersion) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, tmdb.ErrKeyEmpty
	}
	db.mtx.RLock()
	defer db.mtx.RUnlock()

	i := db.btree.Get(newKey(key))
	if i != nil {
		return i.(*item).value, nil
	}
	return nil, nil
}

// Has implements DB.
func (db *dbVersion) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, tmdb.ErrKeyEmpty
	}
	db.mtx.RLock()
	defer db.mtx.RUnlock()

	return db.btree.Has(newKey(key)), nil
}

// Set implements DB.
func (db *dbVersion) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return tmdb.ErrKeyEmpty
	}
	if value == nil {
		return tmdb.ErrValueNil
	}
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.set(key, value)
	return nil
}

// set sets a value without locking the mutex.
func (db *dbVersion) set(key []byte, value []byte) {
	db.btree.ReplaceOrInsert(newPair(key, value))
}

// Delete implements DB.
func (db *dbVersion) Delete(key []byte) error {
	if len(key) == 0 {
		return tmdb.ErrKeyEmpty
	}
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.delete(key)
	return nil
}

// delete deletes a key without locking the mutex.
func (db *dbVersion) delete(key []byte) {
	db.btree.Delete(newKey(key))
}

// Iterator implements DB.
// Takes out a read-lock on the database until the iterator is closed.
func (db *dbVersion) Iterator(start, end []byte) (tmdb.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, tmdb.ErrKeyEmpty
	}
	return newMemDBIterator(db, start, end, false), nil
}

// ReverseIterator implements DB.
// Takes out a read-lock on the database until the iterator is closed.
func (db *dbVersion) ReverseIterator(start, end []byte) (tmdb.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, tmdb.ErrKeyEmpty
	}
	return newMemDBIterator(db, start, end, true), nil
}

func (db *dbVersion) Commit() error {
	// no-op, like Close()
	return nil
}
func (db *dbVersion) Discard() {}

// Print implements DB.
func (db *MemDB) Print() error {
	db.mtx.RLock()
	defer db.mtx.RUnlock()

	db.btree.Ascend(func(i btree.Item) bool {
		item := i.(*item)
		fmt.Printf("[%X]:\t[%X]\n", item.key, item.value)
		return true
	})
	return nil
}

// Stats implements DB.
func (db *MemDB) Stats() map[string]string {
	db.mtx.RLock()
	defer db.mtx.RUnlock()

	stats := make(map[string]string)
	stats["database.type"] = "memDB"
	stats["database.size"] = fmt.Sprintf("%d", db.btree.Len())
	return stats
}
