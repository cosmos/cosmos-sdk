package badgerdb

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	dbm "github.com/cosmos/cosmos-sdk/db"

	"github.com/dgraph-io/badger/v3"
)

var (
	versionsFilename string = "versions.csv"
)

type BadgerDB struct {
	db   *badger.DB
	vmgr *dbm.VersionManager
	mtx  sync.RWMutex
}

var _ dbm.DB = (*BadgerDB)(nil)
var _ dbm.DBReader = (*badgerTxn)(nil)
var _ dbm.DBReadWriter = (*badgerTxn)(nil)

type badgerTxn struct {
	txn *badger.Txn
}

// NewDB creates a Badger key-value store backed to the
// directory dir supplied. If dir does not exist, it will be created.
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

// NewDBWithOptions creates a BadgerDB key value store
// gives the flexibility of initializing a database with the
// respective options.
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
func readVersionsFile(path string) (*dbm.VersionManager, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := csv.NewReader(file)
	r.FieldsPerRecord = 1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	var versions []uint64
	for _, row := range rows {
		version, err := strconv.ParseUint(row[0], 10, 64)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return dbm.NewVersionManager(versions), nil
}

// Write version metadata to CSV file
func writeVersionsFile(vm *dbm.VersionManager, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	var rows [][]string
	for _, ver := range vm.Versions {
		rows = append(rows, []string{strconv.FormatUint(ver, 10)})
	}
	return w.WriteAll(rows)
}

func (b *BadgerDB) ReaderAt(version uint64) dbm.DBReader {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	if !b.vmgr.Valid(version) {
		return nil
	}
	return &badgerTxn{txn: b.db.NewTransactionAt(version, false)}
}

func (b *BadgerDB) ReadWriter() dbm.DBReadWriter {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	return &badgerTxn{txn: b.db.NewTransactionAt(b.vmgr.Current(), true)}
}

func (b *BadgerDB) Close() error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	writeVersionsFile(b.vmgr, filepath.Join(b.db.Opts().Dir, versionsFilename))
	return b.db.Close()
}

func (b *BadgerDB) CurrentVersion() uint64 {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	return b.vmgr.Current()
}

func (b *BadgerDB) Versions() []uint64 {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	return b.vmgr.Versions
}

func (b *BadgerDB) SaveVersion() uint64 {
	// TODO: wait on any pending txns
	b.mtx.Lock()
	defer b.mtx.Unlock()
	return b.vmgr.Save()
}

func (b *badgerTxn) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, dbm.ErrKeyEmpty
	}

	item, err := b.txn.Get(key)
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

func (b *badgerTxn) Has(key []byte) (bool, error) {
	if len(key) == 0 {
		return false, dbm.ErrKeyEmpty
	}

	_, err := b.txn.Get(key)
	if err != nil && err != badger.ErrKeyNotFound {
		return false, err
	}
	return (err != badger.ErrKeyNotFound), nil
}

func (b *badgerTxn) Set(key, value []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	if value == nil {
		return dbm.ErrValueNil
	}
	return b.txn.Set(key, value)
}

func (b *badgerTxn) Delete(key []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	return b.txn.Delete(key)
}

func (b *badgerTxn) Commit() error {
	// All commits write to the same (current) version until next SaveVersion() call
	return b.txn.CommitAt(b.txn.ReadTs(), nil)
}

func (b *badgerTxn) Discard() { b.txn.Discard() }

func (b *badgerTxn) iteratorOpts(start, end []byte, opts badger.IteratorOptions) (*badgerDBIterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, dbm.ErrKeyEmpty
	}
	iter := b.txn.NewIterator(opts)
	iter.Rewind()
	iter.Seek(start)
	if opts.Reverse && iter.Valid() && bytes.Equal(iter.Item().Key(), start) {
		// If we're going in reverse, our starting point was "end", which is exclusive.
		iter.Next()
	}
	return &badgerDBIterator{
		reverse: opts.Reverse,
		start:   start,
		end:     end,
		iter:    iter,
	}, nil
}

func (b *badgerTxn) Iterator(start, end []byte) (dbm.Iterator, error) {
	opts := badger.DefaultIteratorOptions
	return b.iteratorOpts(start, end, opts)
}

func (b *badgerTxn) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	opts := badger.DefaultIteratorOptions
	opts.Reverse = true
	return b.iteratorOpts(end, start, opts)
}

// TODO
func (b *BadgerDB) Stats() map[string]string { return nil }

type badgerDBIterator struct {
	reverse    bool
	start, end []byte

	iter *badger.Iterator

	lastErr error
}

func (i *badgerDBIterator) Close() error {
	i.iter.Close()
	return nil
}

func (i *badgerDBIterator) Domain() (start, end []byte) { return i.start, i.end }
func (i *badgerDBIterator) Error() error                { return i.lastErr }

func (i *badgerDBIterator) Next() {
	if !i.Valid() {
		panic("iterator is invalid")
	}
	i.iter.Next()
}

func (i *badgerDBIterator) Valid() bool {
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

func (i *badgerDBIterator) Key() []byte {
	if !i.Valid() {
		panic("iterator is invalid")
	}
	return i.iter.Item().KeyCopy(nil)
}

func (i *badgerDBIterator) Value() []byte {
	if !i.Valid() {
		panic("iterator is invalid")
	}
	val, err := i.iter.Item().ValueCopy(nil)
	if err != nil {
		i.lastErr = err
	}
	return val
}
