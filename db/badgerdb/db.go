package badgerdb

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/cosmos/cosmos-sdk/db"
	dbutil "github.com/cosmos/cosmos-sdk/db/internal"

	"github.com/dgraph-io/badger/v3"
	bpb "github.com/dgraph-io/badger/v3/pb"
	"github.com/dgraph-io/ristretto/z"
)

var versionsFilename = "versions.csv"

var (
	_ db.DBConnection = (*BadgerDB)(nil)
	_ db.DBReader     = (*badgerTxn)(nil)
	_ db.DBWriter     = (*badgerWriter)(nil)
	_ db.DBReadWriter = (*badgerWriter)(nil)
)

// BadgerDB is a connection to a BadgerDB key-value database.
type BadgerDB struct {
	db          *badger.DB
	vmgr        *versionManager
	mtx         sync.RWMutex
	openWriters int32
}

type badgerTxn struct {
	txn *badger.Txn
	db  *BadgerDB
}

type badgerWriter struct {
	badgerTxn
	discarded bool
}

type badgerIterator struct {
	reverse    bool
	start, end []byte
	iter       *badger.Iterator
	lastErr    error
	// Whether iterator has been advanced to the first element (is fully initialized)
	primed bool
}

// Map our versions to Badger timestamps.
//
// A badger Txn's commit timestamp must be strictly greater than a record's "last-read"
// timestamp in order to detect conflicts, and a Txn must be read at a timestamp after last
// commit to see current state. So we must use commit increments that are more
// granular than our version interval, and map versions to the corresponding timestamp.
type versionManager struct {
	*db.VersionManager
	vmap   map[uint64]uint64
	lastTs uint64
}

// NewDB creates or loads a BadgerDB key-value database inside the given directory.
// If dir does not exist, it will be created.
func NewDB(dir string) (*BadgerDB, error) {
	// Since Badger doesn't support database names, we join both to obtain
	// the final directory to use for the database.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	opts := badger.DefaultOptions(dir)
	opts.SyncWrites = false // note that we have Sync methods
	opts.Logger = nil       // badger is too chatty by default
	return NewDBWithOptions(opts)
}

// NewDBWithOptions creates a BadgerDB key-value database with the specified Options
// (https://pkg.go.dev/github.com/dgraph-io/badger/v3#Options)
func NewDBWithOptions(opts badger.Options) (*BadgerDB, error) {
	d, err := badger.OpenManaged(opts)
	if err != nil {
		return nil, err
	}
	vmgr, err := readVersionsFile(filepath.Join(opts.Dir, versionsFilename))
	if err != nil {
		return nil, err
	}
	return &BadgerDB{
		db:   d,
		vmgr: vmgr,
	}, nil
}

// Load metadata CSV file containing valid versions
func readVersionsFile(path string) (*versionManager, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0o644)
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
	var (
		versions []uint64
		lastTs   uint64
	)
	vmap := map[uint64]uint64{}
	for _, row := range rows {
		version, err := strconv.ParseUint(row[0], 10, 64)
		if err != nil {
			return nil, err
		}
		ts, err := strconv.ParseUint(row[1], 10, 64)
		if err != nil {
			return nil, err
		}
		if version == 0 { // 0 maps to the latest timestamp
			lastTs = ts
		}
		versions = append(versions, version)
		vmap[version] = ts
	}
	vmgr := db.NewVersionManager(versions)
	return &versionManager{
		VersionManager: vmgr,
		vmap:           vmap,
		lastTs:         lastTs,
	}, nil
}

// Write version metadata to CSV file
func writeVersionsFile(vm *versionManager, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	rows := [][]string{
		{"0", strconv.FormatUint(vm.lastTs, 10)},
	}
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

func (b *BadgerDB) Reader() db.DBReader {
	b.mtx.RLock()
	ts := b.vmgr.lastTs
	b.mtx.RUnlock()
	return &badgerTxn{txn: b.db.NewTransactionAt(ts, false), db: b}
}

func (b *BadgerDB) ReaderAt(version uint64) (db.DBReader, error) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	ts, has := b.vmgr.versionTs(version)
	if !has {
		return nil, db.ErrVersionDoesNotExist
	}
	return &badgerTxn{txn: b.db.NewTransactionAt(ts, false), db: b}, nil
}

func (b *BadgerDB) ReadWriter() db.DBReadWriter {
	atomic.AddInt32(&b.openWriters, 1)
	b.mtx.RLock()
	ts := b.vmgr.lastTs
	b.mtx.RUnlock()
	return &badgerWriter{badgerTxn{txn: b.db.NewTransactionAt(ts, true), db: b}, false}
}

func (b *BadgerDB) Writer() db.DBWriter {
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
func (b *BadgerDB) Versions() (db.VersionSet, error) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	return b.vmgr, nil
}

func (b *BadgerDB) save(target uint64) (uint64, error) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if b.openWriters > 0 {
		return 0, db.ErrOpenTransactions
	}
	b.vmgr = b.vmgr.Copy()
	return b.vmgr.Save(target)
}

// SaveNextVersion implements DBConnection.
func (b *BadgerDB) SaveNextVersion() (uint64, error) {
	return b.save(0)
}

// SaveVersion implements DBConnection.
func (b *BadgerDB) SaveVersion(target uint64) error {
	if target == 0 {
		return db.ErrInvalidVersion
	}
	_, err := b.save(target)
	return err
}

func (b *BadgerDB) DeleteVersion(target uint64) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if !b.vmgr.Exists(target) {
		return db.ErrVersionDoesNotExist
	}
	b.vmgr = b.vmgr.Copy()
	b.vmgr.Delete(target)
	return nil
}

func (b *BadgerDB) Revert() error {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	if b.openWriters > 0 {
		return db.ErrOpenTransactions
	}

	// Revert from latest commit timestamp to last "saved" timestamp
	// if no versions exist, use 0 as it precedes any possible commit timestamp
	var target uint64
	last := b.vmgr.Last()
	if last == 0 {
		target = 0
	} else {
		var has bool
		if target, has = b.vmgr.versionTs(last); !has {
			return errors.New("bad version history")
		}
	}
	lastTs := b.vmgr.lastTs
	if target == lastTs {
		return nil
	}

	// Badger provides no way to rollback committed data, so we undo all changes
	// since the target version using the Stream API
	stream := b.db.NewStreamAt(lastTs)
	// Skips unchanged keys
	stream.ChooseKey = func(item *badger.Item) bool { return item.Version() > target }
	// Scans for value at target version
	stream.KeyToList = func(key []byte, itr *badger.Iterator) (*bpb.KVList, error) {
		kv := bpb.KV{Key: key}
		// advance down to <= target version
		itr.Next() // we have at least one newer version
		for itr.Valid() && bytes.Equal(key, itr.Item().Key()) && itr.Item().Version() > target {
			itr.Next()
		}
		if itr.Valid() && bytes.Equal(key, itr.Item().Key()) && !itr.Item().IsDeletedOrExpired() {
			var err error
			kv.Value, err = itr.Item().ValueCopy(nil)
			if err != nil {
				return nil, err
			}
		}
		return &bpb.KVList{Kv: []*bpb.KV{&kv}}, nil
	}
	txn := b.db.NewTransactionAt(lastTs, true)
	defer txn.Discard()
	stream.Send = func(buf *z.Buffer) error {
		kvl, err := badger.BufferToKVList(buf)
		if err != nil {
			return err
		}
		// nil Value indicates a deleted entry
		for _, kv := range kvl.Kv {
			if kv.Value == nil {
				err = txn.Delete(kv.Key)
				if err != nil {
					return err
				}
			} else {
				err = txn.Set(kv.Key, kv.Value)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	err := stream.Orchestrate(context.Background())
	if err != nil {
		return err
	}
	return txn.CommitAt(lastTs, nil)
}

func (b *BadgerDB) Stats() map[string]string { return nil }

func (tx *badgerTxn) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, db.ErrKeyEmpty
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
		return false, db.ErrKeyEmpty
	}

	_, err := tx.txn.Get(key)
	if err != nil && err != badger.ErrKeyNotFound {
		return false, err
	}
	return (err != badger.ErrKeyNotFound), nil
}

func (tx *badgerWriter) Set(key, value []byte) error {
	if err := dbutil.ValidateKv(key, value); err != nil {
		return err
	}
	return tx.txn.Set(key, value)
}

func (tx *badgerWriter) Delete(key []byte) error {
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	return tx.txn.Delete(key)
}

func (tx *badgerWriter) Commit() (err error) {
	if tx.discarded {
		return errors.New("transaction has been discarded")
	}
	defer func() { err = dbutil.CombineErrors(err, tx.Discard(), "Discard also failed") }()
	// Commit to the current commit timestamp, after ensuring it is > ReadTs
	tx.db.mtx.RLock()
	tx.db.vmgr.updateCommitTs(tx.txn.ReadTs())
	ts := tx.db.vmgr.lastTs
	tx.db.mtx.RUnlock()
	err = tx.txn.CommitAt(ts, nil)
	return
}

func (tx *badgerTxn) Discard() error {
	tx.txn.Discard()
	return nil
}

func (tx *badgerWriter) Discard() error {
	if !tx.discarded {
		defer atomic.AddInt32(&tx.db.openWriters, -1)
		tx.discarded = true
	}
	return tx.badgerTxn.Discard()
}

func (tx *badgerTxn) iteratorOpts(start, end []byte, opts badger.IteratorOptions) (*badgerIterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, db.ErrKeyEmpty
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

func (tx *badgerTxn) Iterator(start, end []byte) (db.Iterator, error) {
	opts := badger.DefaultIteratorOptions
	return tx.iteratorOpts(start, end, opts)
}

func (tx *badgerTxn) ReverseIterator(start, end []byte) (db.Iterator, error) {
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

func (vm *versionManager) versionTs(ver uint64) (uint64, bool) {
	ts, has := vm.vmap[ver]
	return ts, has
}

// updateCommitTs increments the lastTs if equal to readts.
func (vm *versionManager) updateCommitTs(readts uint64) {
	if vm.lastTs == readts {
		vm.lastTs++
	}
}

// Atomically accesses the last commit timestamp used as a version marker.
func (vm *versionManager) lastCommitTs() uint64 {
	return atomic.LoadUint64(&vm.lastTs)
}

func (vm *versionManager) Copy() *versionManager {
	vmap := map[uint64]uint64{}
	for ver, ts := range vm.vmap {
		vmap[ver] = ts
	}
	return &versionManager{
		VersionManager: vm.VersionManager.Copy(),
		vmap:           vmap,
		lastTs:         vm.lastCommitTs(),
	}
}

func (vm *versionManager) Save(target uint64) (uint64, error) {
	id, err := vm.VersionManager.Save(target)
	if err != nil {
		return 0, err
	}
	vm.vmap[id] = vm.lastTs // non-atomic, already guarded by the vmgr mutex
	return id, nil
}

func (vm *versionManager) Delete(target uint64) {
	vm.VersionManager.Delete(target)
	delete(vm.vmap, target)
}
