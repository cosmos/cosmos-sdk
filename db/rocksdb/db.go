package rocksdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/tecbot/gorocksdb"
)

var (
	checkpointFileFormat string = "%020d.db"
)

type dbManager struct {
	current dbCxn
	dir     string
	opts    dbOptions
	vmgr    *dbm.VersionManager
	mtx     *sync.RWMutex
}

type dbCxn = *gorocksdb.TransactionDB

type dbTxn struct {
	// db *dbManager
	txn      *gorocksdb.Transaction
	writable bool
	opts     *dbOptions
}

type dbOptions struct {
	dbo  *gorocksdb.Options
	tdbo *gorocksdb.TransactionDBOptions
	txo  *gorocksdb.TransactionOptions
	ro   *gorocksdb.ReadOptions
	wo   *gorocksdb.WriteOptions
}

var _ dbm.DB = (*dbManager)(nil)
var _ dbm.DBReader = (*dbTxn)(nil)
var _ dbm.DBReadWriter = (*dbTxn)(nil)

func NewDB(dir string) (*dbManager, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// default rocksdb option, good enough for most cases, including heavy workloads.
	// 1GB table cache, 512MB write buffer(may use 50% more on heavy workloads).
	// compression: snappy as default, need to -lsnappy to enable.
	bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(gorocksdb.NewLRUCache(1 << 30))
	bbto.SetFilterPolicy(gorocksdb.NewBloomFilter(10))
	dbo := gorocksdb.NewDefaultOptions()
	dbo.SetBlockBasedTableFactory(bbto)
	dbo.SetCreateIfMissing(true)
	dbo.IncreaseParallelism(runtime.NumCPU())
	// 1.5GB maximum memory use for writebuffer.
	dbo.OptimizeLevelStyleCompaction(512 * 1024 * 1024)
	tdbo := gorocksdb.NewDefaultTransactionDBOptions()

	opts := dbOptions{
		dbo:  dbo,
		tdbo: tdbo,
		txo:  gorocksdb.NewDefaultTransactionOptions(),
		ro:   gorocksdb.NewDefaultReadOptions(),
		wo:   gorocksdb.NewDefaultWriteOptions(),
	}
	mgr := &dbManager{
		dir:  dir,
		opts: opts,
		mtx:  &sync.RWMutex{},
	}

	err := os.MkdirAll(mgr.checkpointsDir(), 0755)
	if err != nil {
		return nil, err
	}
	if mgr.vmgr, err = readVersions(mgr.checkpointsDir()); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dir, "current.db")
	mgr.current, err = gorocksdb.OpenTransactionDb(dbo, tdbo, dbPath)
	if err != nil {
		return nil, err
	}
	return mgr, nil
}

func (mgr *dbManager) checkpointsDir() string {
	return filepath.Join(mgr.dir, "checkpoints")
}

// Reads directory for checkpoints files
func readVersions(dir string) (*dbm.VersionManager, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var versions []uint64
	for _, f := range files {
		var version uint64
		if _, err := fmt.Sscanf(f.Name(), checkpointFileFormat, &version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return dbm.NewVersionManager(versions), nil
}

func (mgr *dbManager) openCheckpoint(ver uint64) (dbCxn, error) {
	dbPath := filepath.Join(mgr.checkpointsDir(), fmt.Sprintf(checkpointFileFormat, ver))
	if stat, err := os.Stat(dbPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		return nil, err
	} else if !stat.IsDir() {
		return nil, nil
	}
	db, err := gorocksdb.OpenTransactionDb(mgr.opts.dbo, mgr.opts.tdbo, dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (mgr *dbManager) ReaderAt(ver uint64) dbm.DBReader {
	var db dbCxn
	if ver == mgr.vmgr.Current() {
		db = mgr.current
	} else {
		var err error
		if db, err = mgr.openCheckpoint(ver); db == nil {
			return nil
		}
		if err != nil {
			panic(err) // fixme
		}
	}
	return &dbTxn{
		// todo: meaning of oldtransaction?
		txn:      db.TransactionBegin(mgr.opts.wo, mgr.opts.txo, nil),
		opts:     &mgr.opts,
		writable: false,
	}
}

func (mgr *dbManager) ReadWriter() dbm.DBReadWriter {
	return &dbTxn{
		txn:      mgr.current.TransactionBegin(mgr.opts.wo, mgr.opts.txo, nil),
		opts:     &mgr.opts,
		writable: true,
	}
}

func (mgr *dbManager) CurrentVersion() uint64 {
	mgr.mtx.Lock()
	defer mgr.mtx.Unlock()
	return mgr.vmgr.Current()
}

func (mgr *dbManager) Versions() []uint64 {
	mgr.mtx.Lock()
	defer mgr.mtx.Unlock()
	return mgr.vmgr.Versions
}

func (mgr *dbManager) SaveVersion() uint64 {
	mgr.mtx.Lock()
	defer mgr.mtx.Unlock()
	ver := mgr.vmgr.Save()

	cp, err := mgr.current.NewCheckpoint()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(mgr.checkpointsDir(), fmt.Sprintf(checkpointFileFormat, ver))
	if err := cp.CreateCheckpoint(dir, 0); err != nil {
		panic(err)
	}
	cp.Destroy()

	return ver
}

// Close implements DB.
func (mgr *dbManager) Close() error {
	mgr.current.Close()
	mgr.opts.destroy()
	return nil
}

// TODO
func (mgr *dbManager) Stats() map[string]string { return nil }

// Get implements DBReader.
func (t *dbTxn) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, dbm.ErrKeyEmpty
	}
	var res *gorocksdb.Slice
	var err error
	if t.writable {
		res, err = t.txn.GetForUpdate(t.opts.ro, key)
	} else {
		res, err = t.txn.Get(t.opts.ro, key)
	}
	if err != nil {
		return nil, err
	}
	return moveSliceToBytes(res), nil
}

// Has implements DBReader.
func (t *dbTxn) Has(key []byte) (bool, error) {
	bytes, err := t.Get(key)
	if err != nil {
		return false, err
	}
	return bytes != nil, nil
}

// Set implements DBWriter.
func (t *dbTxn) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	if value == nil {
		return dbm.ErrValueNil
	}
	return t.txn.Put(key, value)
}

// Delete implements DBWriter.
func (t *dbTxn) Delete(key []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	return t.txn.Delete(key)
}

func (t *dbTxn) Commit() error {
	return t.txn.Commit()
}
func (t *dbTxn) Discard() {
	t.txn.Destroy()
}

// Iterator implements DBReader.
func (t *dbTxn) Iterator(start, end []byte) (dbm.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, dbm.ErrKeyEmpty
	}
	itr := t.txn.NewIterator(t.opts.ro)
	return newRocksDBIterator(itr, start, end, false), nil
}

// ReverseIterator implements DBReader.
func (t *dbTxn) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, dbm.ErrKeyEmpty
	}
	itr := t.txn.NewIterator(t.opts.ro)
	return newRocksDBIterator(itr, start, end, true), nil
}

func (o dbOptions) destroy() {
	o.ro.Destroy()
	o.wo.Destroy()
	o.txo.Destroy()
	o.dbo.Destroy()
}
