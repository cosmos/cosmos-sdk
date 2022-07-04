package memdb

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"

	dbutil "github.com/cosmos/cosmos-sdk/db/internal"
	"github.com/cosmos/cosmos-sdk/db/types"
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
// Note: Currently, transactions do not detect write conflicts, so multiple writers cannot be
// safely committed to overlapping domains. Because of this, the number of open writers is
// limited to 1.
type MemDB struct {
	btree       *btree.BTree            // Main contents
	mtx         sync.RWMutex            // Guards version history
	saved       map[uint64]*btree.BTree // Past versions
	vmgr        *types.VersionManager   // Mirrors version keys
	openWriters int32                   // Open writers
}

type dbTxn struct {
	btree *btree.BTree
	db    *MemDB
}
type dbWriter struct{ dbTxn }

var (
	_ types.Connection = (*MemDB)(nil)
	_ types.Reader     = (*dbTxn)(nil)
	_ types.Writer     = (*dbWriter)(nil)
	_ types.ReadWriter = (*dbWriter)(nil)
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
		vmgr:  types.NewVersionManager(nil),
	}
}

func (dbm *MemDB) newTxn(tree *btree.BTree) dbTxn {
	return dbTxn{tree, dbm}
}

// Close implements Connection.
// Close is a noop since for an in-memory database, we don't have a destination to flush
// contents to nor do we want any data loss on invoking Close().
// See the discussion in https://github.com/tendermint/tendermint/libs/pull/56
func (dbm *MemDB) Close() error {
	return nil
}

// Versions implements Connection.
func (dbm *MemDB) Versions() (types.VersionSet, error) {
	dbm.mtx.RLock()
	defer dbm.mtx.RUnlock()
	return dbm.vmgr, nil
}

// Reader implements Connection.
func (dbm *MemDB) Reader() types.Reader {
	dbm.mtx.RLock()
	defer dbm.mtx.RUnlock()
	ret := dbm.newTxn(dbm.btree)
	return &ret
}

// ReaderAt implements Connection.
func (dbm *MemDB) ReaderAt(version uint64) (types.Reader, error) {
	dbm.mtx.RLock()
	defer dbm.mtx.RUnlock()
	tree, ok := dbm.saved[version]
	if !ok {
		return nil, types.ErrVersionDoesNotExist
	}
	ret := dbm.newTxn(tree)
	return &ret, nil
}

// Writer implements Connection.
func (dbm *MemDB) Writer() types.Writer {
	return dbm.ReadWriter()
}

// ReadWriter implements Connection.
func (dbm *MemDB) ReadWriter() types.ReadWriter {
	dbm.mtx.RLock()
	defer dbm.mtx.RUnlock()
	atomic.AddInt32(&dbm.openWriters, 1)
	// Clone creates a copy-on-write extension of the current tree
	return &dbWriter{dbm.newTxn(dbm.btree.Clone())}
}

func (dbm *MemDB) save(target uint64) (uint64, error) {
	dbm.mtx.Lock()
	defer dbm.mtx.Unlock()
	if dbm.openWriters > 0 {
		return 0, types.ErrOpenTransactions
	}

	newVmgr := dbm.vmgr.Copy()
	target, err := newVmgr.Save(target)
	if err != nil {
		return 0, err
	}
	dbm.saved[target] = dbm.btree
	dbm.vmgr = newVmgr
	return target, nil
}

// SaveVersion implements Connection.
func (dbm *MemDB) SaveNextVersion() (uint64, error) {
	return dbm.save(0)
}

// SaveNextVersion implements Connection.
func (dbm *MemDB) SaveVersion(target uint64) error {
	if target == 0 {
		return types.ErrInvalidVersion
	}
	_, err := dbm.save(target)
	return err
}

// DeleteVersion implements Connection.
func (dbm *MemDB) DeleteVersion(target uint64) error {
	dbm.mtx.Lock()
	defer dbm.mtx.Unlock()
	if _, has := dbm.saved[target]; !has {
		return types.ErrVersionDoesNotExist
	}
	delete(dbm.saved, target)
	dbm.vmgr = dbm.vmgr.Copy()
	dbm.vmgr.Delete(target)
	return nil
}

func (dbm *MemDB) Revert() error {
	dbm.mtx.RLock()
	defer dbm.mtx.RUnlock()
	if dbm.openWriters > 0 {
		return types.ErrOpenTransactions
	}
	last := dbm.vmgr.Last()
	if last == 0 {
		dbm.btree = btree.New(bTreeDegree)
		return nil
	}
	return dbm.revert(last)
}

func (dbm *MemDB) RevertTo(target uint64) error {
	dbm.mtx.RLock()
	defer dbm.mtx.RUnlock()
	if dbm.openWriters > 0 {
		return types.ErrOpenTransactions
	}
	if !dbm.vmgr.Exists(target) {
		return types.ErrVersionDoesNotExist
	}
	err := dbm.revert(target)
	if err != nil {
		dbm.vmgr.DeleteAbove(target)
	}
	return err
}

func (dbm *MemDB) revert(target uint64) error {
	var has bool
	dbm.btree, has = dbm.saved[target]
	if !has {
		return fmt.Errorf("bad version history: version %v not saved", target)
	}
	for ver, _ := range dbm.saved {
		if ver > target {
			delete(dbm.saved, ver)
		}
	}
	return nil
}

// Get implements Reader.
func (tx *dbTxn) Get(key []byte) ([]byte, error) {
	if tx.btree == nil {
		return nil, types.ErrTransactionClosed
	}
	if len(key) == 0 {
		return nil, types.ErrKeyEmpty
	}
	i := tx.btree.Get(newKey(key))
	if i != nil {
		return i.(*item).value, nil
	}
	return nil, nil
}

// Has implements Reader.
func (tx *dbTxn) Has(key []byte) (bool, error) {
	if tx.btree == nil {
		return false, types.ErrTransactionClosed
	}
	if len(key) == 0 {
		return false, types.ErrKeyEmpty
	}
	return tx.btree.Has(newKey(key)), nil
}

// Set implements Writer.
func (tx *dbWriter) Set(key []byte, value []byte) error {
	if tx.btree == nil {
		return types.ErrTransactionClosed
	}
	if err := dbutil.ValidateKv(key, value); err != nil {
		return err
	}
	tx.btree.ReplaceOrInsert(newPair(key, value))
	return nil
}

// Delete implements Writer.
func (tx *dbWriter) Delete(key []byte) error {
	if tx.btree == nil {
		return types.ErrTransactionClosed
	}
	if len(key) == 0 {
		return types.ErrKeyEmpty
	}
	tx.btree.Delete(newKey(key))
	return nil
}

// Iterator implements Reader.
// Takes out a read-lock on the database until the iterator is closed.
func (tx *dbTxn) Iterator(start, end []byte) (types.Iterator, error) {
	if tx.btree == nil {
		return nil, types.ErrTransactionClosed
	}
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, types.ErrKeyEmpty
	}
	return newMemDBIterator(tx, start, end, false), nil
}

// ReverseIterator implements Reader.
// Takes out a read-lock on the database until the iterator is closed.
func (tx *dbTxn) ReverseIterator(start, end []byte) (types.Iterator, error) {
	if tx.btree == nil {
		return nil, types.ErrTransactionClosed
	}
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, types.ErrKeyEmpty
	}
	return newMemDBIterator(tx, start, end, true), nil
}

// Commit implements Writer.
func (tx *dbWriter) Commit() error {
	if tx.btree == nil {
		return types.ErrTransactionClosed
	}
	tx.db.mtx.Lock()
	defer tx.db.mtx.Unlock()
	tx.db.btree = tx.btree
	return tx.Discard()
}

// Discard implements Reader.
func (tx *dbTxn) Discard() error {
	if tx.btree != nil {
		tx.btree = nil
	}
	return nil
}

// Discard implements Writer.
func (tx *dbWriter) Discard() error {
	if tx.btree != nil {
		defer atomic.AddInt32(&tx.db.openWriters, -1)
	}
	return tx.dbTxn.Discard()
}

// Print prints the database contents.
func (dbm *MemDB) Print() error {
	dbm.mtx.RLock()
	defer dbm.mtx.RUnlock()

	dbm.btree.Ascend(func(i btree.Item) bool {
		item := i.(*item)
		fmt.Printf("[%X]:\t[%X]\n", item.key, item.value)
		return true
	})
	return nil
}

// Stats implements Connection.
func (dbm *MemDB) Stats() map[string]string {
	dbm.mtx.RLock()
	defer dbm.mtx.RUnlock()

	stats := make(map[string]string)
	stats["database.type"] = "memDB"
	stats["database.size"] = fmt.Sprintf("%d", dbm.btree.Len())
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
