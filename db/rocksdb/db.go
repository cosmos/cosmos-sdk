package rocksdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/tecbot/gorocksdb"
)

var (
	checkpointFileFormat string = "%020d.db"
)

var (
	_ dbm.DBConnection = (*RocksDB)(nil)
	_ dbm.DBReader     = (*dbTxn)(nil)
	_ dbm.DBWriter     = (*dbWriter)(nil)
	_ dbm.DBReadWriter = (*dbWriter)(nil)
)

// RocksDB is a connection to a RocksDB key-value database.
type RocksDB = dbManager

type dbManager struct {
	current *dbConnection
	dir     string
	opts    dbOptions
	vmgr    *dbm.VersionManager
	mtx     *sync.RWMutex
	// Track open DBWriters
	openWriters int32
}

type dbConnection = gorocksdb.OptimisticTransactionDB

type dbTxn struct {
	txn *gorocksdb.Transaction
	mgr *dbManager
}
type dbWriter struct{ dbTxn }

type dbOptions struct {
	dbo *gorocksdb.Options
	txo *gorocksdb.OptimisticTransactionOptions
	ro  *gorocksdb.ReadOptions
	wo  *gorocksdb.WriteOptions
}

// NewDB creates a new RocksDB key-value database with inside the given directory.
// If dir does not exist, it will be created.
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

	opts := dbOptions{
		dbo: dbo,
		txo: gorocksdb.NewDefaultOptimisticTransactionOptions(),
		ro:  gorocksdb.NewDefaultReadOptions(),
		wo:  gorocksdb.NewDefaultWriteOptions(),
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
	mgr.current, err = gorocksdb.OpenOptimisticTransactionDb(dbo, dbPath)
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

func (mgr *dbManager) checkpointPath(ver uint64) (string, error) {
	dbPath := filepath.Join(mgr.checkpointsDir(), fmt.Sprintf(checkpointFileFormat, ver))
	if stat, err := os.Stat(dbPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = dbm.ErrVersionDoesNotExist
		}
		return "", err
	} else if !stat.IsDir() {
		return "", dbm.ErrVersionDoesNotExist
	}
	return dbPath, nil
}

func (mgr *dbManager) openCheckpoint(ver uint64) (*dbConnection, error) {
	dbPath, err := mgr.checkpointPath(ver)
	if err != nil {
		return nil, err
	}
	db, err := gorocksdb.OpenOptimisticTransactionDb(mgr.opts.dbo, dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (mgr *dbManager) Reader() dbm.DBReader {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	return &dbTxn{
		txn: mgr.current.TransactionBegin(mgr.opts.wo, mgr.opts.txo, nil),
		mgr: mgr,
	}
}

func (mgr *dbManager) ReaderAt(ver uint64) (dbm.DBReader, error) {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	// TODO: cache opened checkpoints
	db, err := mgr.openCheckpoint(ver)
	if err != nil {
		return nil, err
	}

	return &dbTxn{
		// todo: meaning of oldtransaction?
		txn: db.TransactionBegin(mgr.opts.wo, mgr.opts.txo, nil),
		mgr: mgr,
	}, nil
}

func (mgr *dbManager) ReadWriter() dbm.DBReadWriter {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	atomic.AddInt32(&mgr.openWriters, 1)
	return &dbWriter{dbTxn{
		txn: mgr.current.TransactionBegin(mgr.opts.wo, mgr.opts.txo, nil),
		mgr: mgr,
	}}
}

func (mgr *dbManager) Writer() dbm.DBWriter {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	atomic.AddInt32(&mgr.openWriters, 1)
	return mgr.newRocksDBBatch()
}

func (mgr *dbManager) Versions() (dbm.VersionSet, error) {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	return mgr.vmgr, nil
}

// SaveVersion implements DBConnection.
func (mgr *dbManager) SaveNextVersion() (uint64, error) {
	return mgr.save(0)
}

// SaveNextVersion implements DBConnection.
func (mgr *dbManager) SaveVersion(target uint64) error {
	if target == 0 {
		return dbm.ErrInvalidVersion
	}
	_, err := mgr.save(target)
	return err
}

func (mgr *dbManager) save(target uint64) (uint64, error) {
	mgr.mtx.Lock()
	defer mgr.mtx.Unlock()
	if mgr.openWriters > 0 {
		return 0, dbm.ErrOpenTransactions
	}
	newVmgr := mgr.vmgr.Copy()
	ver, err := newVmgr.Save(target)
	if err != nil {
		return 0, err
	}
	cp, err := mgr.current.NewCheckpoint()
	if err != nil {
		return 0, err
	}
	dir := filepath.Join(mgr.checkpointsDir(), fmt.Sprintf(checkpointFileFormat, ver))
	if err := cp.CreateCheckpoint(dir, 0); err != nil {
		panic(err)
	}
	cp.Destroy()
	mgr.vmgr = newVmgr
	return ver, nil
}

func (mgr *dbManager) DeleteVersion(ver uint64) error {
	mgr.mtx.Lock()
	defer mgr.mtx.Unlock()

	dbPath, err := mgr.checkpointPath(ver)
	if err != nil {
		return err
	}
	mgr.vmgr = mgr.vmgr.Copy()
	mgr.vmgr.Delete(ver)
	return os.RemoveAll(dbPath)
}

// Close implements DBConnection.
func (mgr *dbManager) Close() error {
	mgr.current.Close()
	mgr.opts.destroy()
	return nil
}

// Close implements DBConnection.
func (mgr *dbManager) Stats() map[string]string {
	keys := []string{"rocksdb.stats"}
	stats := make(map[string]string, len(keys))
	for _, key := range keys {
		stats[key] = mgr.current.GetProperty(key)
	}
	return stats
}

// Get implements DBReader.
func (tx *dbTxn) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, dbm.ErrKeyEmpty
	}
	res, err := tx.txn.Get(tx.mgr.opts.ro, key)
	if err != nil {
		return nil, err
	}
	return moveSliceToBytes(res), nil
}

// Get implements DBReader.
func (tx *dbWriter) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, dbm.ErrKeyEmpty
	}
	res, err := tx.txn.GetForUpdate(tx.mgr.opts.ro, key)
	if err != nil {
		return nil, err
	}
	return moveSliceToBytes(res), nil
}

// Has implements DBReader.
func (tx *dbTxn) Has(key []byte) (bool, error) {
	bytes, err := tx.Get(key)
	if err != nil {
		return false, err
	}
	return bytes != nil, nil
}

// Set implements DBWriter.
func (tx *dbWriter) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	if value == nil {
		return dbm.ErrValueNil
	}
	return tx.txn.Put(key, value)
}

// Delete implements DBWriter.
func (tx *dbWriter) Delete(key []byte) error {
	if len(key) == 0 {
		return dbm.ErrKeyEmpty
	}
	return tx.txn.Delete(key)
}

func (tx *dbWriter) Commit() error {
	defer tx.Discard()
	return tx.txn.Commit()
}

func (tx *dbTxn) Discard() {
	tx.txn.Destroy()
}

func (tx *dbWriter) Discard() {
	defer atomic.AddInt32(&tx.mgr.openWriters, -1)
	tx.dbTxn.Discard()
}

// Iterator implements DBReader.
func (tx *dbTxn) Iterator(start, end []byte) (dbm.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, dbm.ErrKeyEmpty
	}
	itr := tx.txn.NewIterator(tx.mgr.opts.ro)
	return newRocksDBIterator(itr, start, end, false), nil
}

// ReverseIterator implements DBReader.
func (tx *dbTxn) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, dbm.ErrKeyEmpty
	}
	itr := tx.txn.NewIterator(tx.mgr.opts.ro)
	return newRocksDBIterator(itr, start, end, true), nil
}

func (o dbOptions) destroy() {
	o.ro.Destroy()
	o.wo.Destroy()
	o.txo.Destroy()
	o.dbo.Destroy()
}
