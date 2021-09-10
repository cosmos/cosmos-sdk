package badgerdb

import (
	"bytes"
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"

	dbm "github.com/cosmos/cosmos-sdk/db"

	"github.com/dgraph-io/badger/v3"
)

var (
	versionsFilename string = "versions.csv"
)

var _ dbm.DBConnection = (*BadgerDB)(nil)
var _ dbm.DBReader = (*badgerTxn)(nil)
var _ dbm.DBWriter = (*badgerWriter)(nil)
var _ dbm.DBReadWriter = (*badgerWriter)(nil)

// BadgerDB is a connection to a BadgerDB key-value database.
type BadgerDB struct {
	db   *badger.DB
	vmgr *versionManager
	mtx  sync.RWMutex
	openWriters int32
}

type badgerTxn struct {
	txn *badger.Txn
	// vmgr *versionManager
	db *BadgerDB
}

type badgerWriter struct {
	badgerTxn
}

type badgerIterator struct {
	reverse    bool
	start, end []byte
	iter       *badger.Iterator
	lastErr    error
	primed     bool
}

// Map our versions to Badger timestamps.
//
// A badger Txn's commit TS must be strictly greater than a record's "last-read"
// TS in order to detect conflicts, and a Txn must be read at a TS after last
// commit to see current state. So we must use commit increments that are more
// granular than a version interval, mapping latter -> former.
type versionManager struct {
	*dbm.VersionManager
	vmap   versionTsMap
	lastTs uint64
}

type versionTsMap map[uint64]uint64

// NewDB creates or loads a BadgerDB key-value database inside the given directory.
// If dir does not exist, it will be created.
func NewDB(dir string) (*BadgerDB, error) {
	// Since Badger doesn't support database names, we join both to obtain
	// the final directory to use for the database.
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	opts := badger.DefaultOptions(dir)
	// todo: NumVersionsToKeep
	opts.SyncWrites = false // note that we have Sync methods
	opts.Logger = nil       // badger is too chatty by default
	return NewDBWithOptions(opts)
}

// NewDBWithOptions creates a BadgerDB key-value database with the specified Options
// (https://pkg.go.dev/github.com/dgraph-io/badger/v3#Options)
func NewDBWithOptions(opts badger.Options) (*BadgerDB, error) {
	db, err := badger.OpenManaged(opts)
	if err != nil {
		return nil, err
	}
	vmgr, err := readVersionsFile(filepath.Join(opts.Dir, versionsFilename))
	if err != nil {
		return nil, err
	}
	return &BadgerDB{
		db:   db,
		vmgr: vmgr,
	}, nil
}

// Load metadata CSV file containing valid versions
func readVersionsFile(path string) (*versionManager, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := csv.NewReader(file)
	r.FieldsPerRecord = 2
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	var versions []uint64
	var ts uint64
	vmap := make(versionTsMap)
	for _, row := range rows {
		version, err := strconv.ParseUint(row[0], 10, 64)
		if err != nil {
			return nil, err
		}
		ts, err = strconv.ParseUint(row[1], 10, 64)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
		vmap[version] = ts
	}
	if ts == 0 {
		ts = 1
	}
	return &versionManager{dbm.NewVersionManager(versions), vmap, ts}, nil
}

// Write version metadata to CSV file
func writeVersionsFile(vm *versionManager, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	var rows [][]string
	for it := vm.Iterator(); it.Next(); {
		version := it.Value()
		ts, ok := vm.vmap[version]
		if !ok {
			panic("version not mapped to ts")
		}
		rows = append(rows, []string{
			strconv.FormatUint(it.Value(), 10),
			strconv.FormatUint(ts, 10),
		})
	}
	return w.WriteAll(rows)
}

func (b *BadgerDB) Reader() dbm.DBReader {
	return &badgerTxn{txn: b.db.NewTransactionAt(math.MaxUint64, false), db: b}
}

func (b *BadgerDB) ReaderAt(version uint64) (dbm.DBReader, error) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	if !b.vmgr.Exists(version) {
		return nil, dbm.ErrVersionDoesNotExist
	}
	return &badgerTxn{txn: b.db.NewTransactionAt(b.vmgr.versionTs(version), false), db: b}, nil
}

func (b *BadgerDB) ReadWriter() dbm.DBReadWriter {
	atomic.AddInt32(&b.openWriters, 1)
	b.mtx.RLock()
	ts := b.vmgr.lastCommitTs()
	b.mtx.RUnlock()
	return &badgerWriter{badgerTxn{txn: b.db.NewTransactionAt(ts, true), db: b}}
}

func (b *BadgerDB) Writer() dbm.DBWriter {
	// Badger has a WriteBatch, but it doesn't support conflict detection
	return b.ReadWriter()
}

func (b *BadgerDB) Close() error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	writeVersionsFile(b.vmgr, filepath.Join(b.db.Opts().Dir, versionsFilename))
	return b.db.Close()
}

// Versions implements DBConnection.
// Returns a VersionSet that is valid until the next call to SaveVersion or DeleteVersion.
func (b *BadgerDB) Versions() (dbm.VersionSet, error) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	return b.vmgr, nil
}

func (b *BadgerDB) save(target uint64) (uint64, error) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if b.openWriters > 0 {
		return 0, dbm.ErrOpenTransactions
	}
	b.vmgr = b.vmgr.Copy()
	return b.vmgr.Save(target)
}

// SaveVersion implements DBConnection.
func (b *BadgerDB) SaveNextVersion() (uint64, error) {
	return b.save(0)
}

// SaveNextVersion implements DBConnection.
func (b *BadgerDB) SaveVersion(target uint64) error {
	if target == 0 {
		return dbm.ErrInvalidVersion
	}
	_, err := b.save(target)
	return err
}

func (b *BadgerDB) DeleteVersion(target uint64) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if !b.vmgr.Exists(target) {
		return dbm.ErrVersionDoesNotExist
	}
	b.vmgr = b.vmgr.Copy()
	b.vmgr.Delete(target)
	return nil
}

func (b *BadgerDB) Stats() map[string]string { return nil }

func (tx *badgerTxn) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, dbm.ErrKeyEmpty
	}

	item, err := tx.txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	val, err := item.ValueCopy(nil)
	if err == nil && val == nil {
		val = []byte{}
	}
	return val, err
}

func (tx *badgerTxn) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, dbm.ErrKeyEmpty
	}

	_, err := tx.txn.Get(key)
	if err != nil && err != badger.ErrKeyNotFound {
		return false, err
	}
	return (err != badger.ErrKeyNotFound), nil
}

func (tx *badgerWriter) Set(key, value []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	if value == nil {
		return dbm.ErrValueNil
	}
	return tx.txn.Set(key, value)
}

func (tx *badgerWriter) Delete(key []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	return tx.txn.Delete(key)
}

func (tx *badgerWriter) Commit() error {
	// Commit to the current commit TS, after ensuring it is > ReadTs
	tx.db.vmgr.updateCommitTs(tx.txn.ReadTs())
	defer tx.Discard()
	return tx.txn.CommitAt(tx.db.vmgr.lastCommitTs(), nil)
}

func (tx *badgerTxn) Discard() {
	tx.txn.Discard()
}

func (tx *badgerWriter) Discard() {
	defer atomic.AddInt32(&tx.db.openWriters, -1)
	tx.badgerTxn.Discard()
}

func (tx *badgerTxn) iteratorOpts(start, end []byte, opts badger.IteratorOptions) (*badgerIterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, dbm.ErrKeyEmpty
	}
	iter := tx.txn.NewIterator(opts)
	iter.Rewind()
	iter.Seek(start)
	if opts.Reverse && iter.Valid() && bytes.Equal(iter.Item().Key(), start) {
		// If we're going in reverse, our starting point was "end", which is exclusive.
		iter.Next()
	}
	return &badgerIterator{
		reverse: opts.Reverse,
		start:   start,
		end:     end,
		iter:    iter,
		primed:  false,
	}, nil
}

func (tx *badgerTxn) Iterator(start, end []byte) (dbm.Iterator, error) {
	opts := badger.DefaultIteratorOptions
	return tx.iteratorOpts(start, end, opts)
}

func (tx *badgerTxn) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	opts := badger.DefaultIteratorOptions
	opts.Reverse = true
	return tx.iteratorOpts(end, start, opts)
}

func (i *badgerIterator) Close() error {
	i.iter.Close()
	return nil
}

func (i *badgerIterator) Domain() (start, end []byte) { return i.start, i.end }
func (i *badgerIterator) Error() error                { return i.lastErr }

func (i *badgerIterator) Next() bool {
	if !i.primed {
		i.primed = true
	} else {
		i.iter.Next()
	}
	return i.Valid()
}

func (i *badgerIterator) Valid() bool {
	if !i.iter.Valid() {
		return false
	}
	if len(i.end) > 0 {
		key := i.iter.Item().Key()
		if c := bytes.Compare(key, i.end); (!i.reverse && c >= 0) || (i.reverse && c < 0) {
			// We're at the end key, or past the end.
			return false
		}
	}
	return true
}

func (i *badgerIterator) Key() []byte {
	if !i.Valid() {
		panic("iterator is invalid")
	}
	return i.iter.Item().KeyCopy(nil)
}

func (i *badgerIterator) Value() []byte {
	if !i.Valid() {
		panic("iterator is invalid")
	}
	val, err := i.iter.Item().ValueCopy(nil)
	if err != nil {
		i.lastErr = err
	}
	return val
}

func (vm *versionManager) versionTs(ver uint64) uint64 {
	return vm.vmap[ver]
}
func (vm *versionManager) lastCommitTs() uint64 {
	return atomic.LoadUint64(&vm.lastTs)
}
func (vm *versionManager) Copy() *versionManager {
	return &versionManager{
		VersionManager: vm.VersionManager.Copy(),
		vmap:           vm.vmap,
		lastTs:         vm.lastTs,
	}
}

// updateCommitTs atomically increments the lastTs if equal to readts.
// Returns the new value.
func (vm *versionManager) updateCommitTs(readts uint64) {
	atomic.CompareAndSwapUint64(&vm.lastTs, readts, readts+1)
}
func (vm *versionManager) Save(target uint64) (uint64, error) {
	id, err := vm.VersionManager.Save(target)
	if err != nil {
		return 0, err
	}
	vm.vmap[id] = atomic.LoadUint64(&vm.lastTs)
	return id, nil
}
