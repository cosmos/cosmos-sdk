package memdb

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/google/btree"
)

const (
	// The approximate number of items and children per B-tree node. Tuned with benchmarks.
	bTreeDegree = 32
)

// MemDB is an in-memory database backend using a B-tree for storage.
//
// For performance reasons, all given and returned keys and values are pointers to the in-memory
// database, so modifying them will cause the stored values to be modified as well. All DB methods
// already specify that keys and values should be considered read-only, but this is especially
// important with MemDB.
//
// Versioning is implemented by maintaining references to copy-on-write clones of the backing btree.
//
// TODO: Currently transactions do not detect write conflicts, so writers cannot be used concurrently.
type MemDB struct {
	btree       *btree.BTree            // Main contents
	mtx         sync.RWMutex            // Guards version history
	saved       map[uint64]*btree.BTree // Past versions
	vmgr        *dbm.VersionManager     // Mirrors version keys
	openWriters int32                   // Open writers
}

type dbTxn struct {
	btree *btree.BTree
	db    *MemDB
}
type dbWriter struct{ dbTxn }

var (
	_ dbm.DBConnection = (*MemDB)(nil)
	_ dbm.DBReader     = (*dbTxn)(nil)
	_ dbm.DBWriter     = (*dbWriter)(nil)
	_ dbm.DBReadWriter = (*dbWriter)(nil)
)

// item is a btree.Item with byte slices as keys and values
type item struct {
	key   []byte
	value []byte
}

// NewDB creates a new in-memory database.
func NewDB() *MemDB {
	return &MemDB{
		btree: btree.New(bTreeDegree),
		saved: make(map[uint64]*btree.BTree),
		vmgr:  dbm.NewVersionManager(nil),
	}
}

func (db *MemDB) newTxn(tree *btree.BTree) dbTxn {
	return dbTxn{tree, db}
}

// Close implements DB.
// Close is a noop since for an in-memory database, we don't have a destination to flush
// contents to nor do we want any data loss on invoking Close().
// See the discussion in https://github.com/tendermint/tendermint/libs/pull/56
func (db *MemDB) Close() error {
	return nil
}

// Versions implements DBConnection.
func (db *MemDB) Versions() (dbm.VersionSet, error) {
	db.mtx.RLock()
	defer db.mtx.RUnlock()
	return db.vmgr, nil
}

// Reader implements DBConnection.
func (db *MemDB) Reader() dbm.DBReader {
	db.mtx.RLock()
	defer db.mtx.RUnlock()
	ret := db.newTxn(db.btree)
	return &ret
}

// ReaderAt implements DBConnection.
func (db *MemDB) ReaderAt(version uint64) (dbm.DBReader, error) {
	db.mtx.RLock()
	defer db.mtx.RUnlock()
	tree, ok := db.saved[version]
	if !ok {
		return nil, dbm.ErrVersionDoesNotExist
	}
	ret := db.newTxn(tree)
	return &ret, nil
}

// Writer implements DBConnection.
func (db *MemDB) Writer() dbm.DBWriter {
	return db.ReadWriter()
}

// ReadWriter implements DBConnection.
func (db *MemDB) ReadWriter() dbm.DBReadWriter {
	db.mtx.RLock()
	defer db.mtx.RUnlock()
	atomic.AddInt32(&db.openWriters, 1)
	// Clone creates a copy-on-write extension of the current tree
	return &dbWriter{db.newTxn(db.btree.Clone())}
}

func (db *MemDB) save(target uint64) (uint64, error) {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if db.openWriters > 0 {
		return 0, dbm.ErrOpenTransactions
	}

	newVmgr := db.vmgr.Copy()
	target, err := newVmgr.Save(target)
	if err != nil {
		return 0, err
	}
	db.saved[target] = db.btree
	db.vmgr = newVmgr
	return target, nil
}

// SaveVersion implements DBConnection.
func (db *MemDB) SaveNextVersion() (uint64, error) {
	return db.save(0)
}

// SaveNextVersion implements DBConnection.
func (db *MemDB) SaveVersion(target uint64) error {
	if target == 0 {
		return dbm.ErrInvalidVersion
	}
	_, err := db.save(target)
	return err
}

// DeleteVersion implements DBConnection.
func (db *MemDB) DeleteVersion(target uint64) error {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	if _, has := db.saved[target]; !has {
		return dbm.ErrVersionDoesNotExist
	}
	delete(db.saved, target)
	db.vmgr = db.vmgr.Copy()
	db.vmgr.Delete(target)
	return nil
}

// Get implements DBReader.
func (tx *dbTxn) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, dbm.ErrKeyEmpty
	}
	i := tx.btree.Get(newKey(key))
	if i != nil {
		return i.(*item).value, nil
	}
	return nil, nil
}

// Has implements DBReader.
func (tx *dbTxn) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, dbm.ErrKeyEmpty
	}
	return tx.btree.Has(newKey(key)), nil
}

// Set implements DBWriter.
func (tx *dbWriter) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	if value == nil {
		return dbm.ErrValueNil
	}
	tx.btree.ReplaceOrInsert(newPair(key, value))
	return nil
}

// Delete implements DBWriter.
func (tx *dbWriter) Delete(key []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	tx.btree.Delete(newKey(key))
	return nil
}

// Iterator implements DBReader.
// Takes out a read-lock on the database until the iterator is closed.
func (tx *dbTxn) Iterator(start, end []byte) (dbm.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, dbm.ErrKeyEmpty
	}
	return newMemDBIterator(tx, start, end, false), nil
}

// ReverseIterator implements DBReader.
// Takes out a read-lock on the database until the iterator is closed.
func (tx *dbTxn) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, dbm.ErrKeyEmpty
	}
	return newMemDBIterator(tx, start, end, true), nil
}

// Commit implements DBWriter.
func (tx *dbWriter) Commit() error {
	tx.db.mtx.Lock()
	defer tx.db.mtx.Unlock()
	defer tx.Discard()
	tx.db.btree = tx.btree
	return nil
}

// Discard implements DBReader and DBWriter.
func (tx *dbTxn) Discard() {}
func (tx *dbWriter) Discard() {
	atomic.AddInt32(&tx.db.openWriters, -1)
}

// Print prints the database contents.
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

// Stats implements DBConnection.
func (db *MemDB) Stats() map[string]string {
	db.mtx.RLock()
	defer db.mtx.RUnlock()

	stats := make(map[string]string)
	stats["database.type"] = "memDB"
	stats["database.size"] = fmt.Sprintf("%d", db.btree.Len())
	return stats
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
