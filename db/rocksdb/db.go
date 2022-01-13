//go:build rocksdb_build

package rocksdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/cosmos/cosmos-sdk/db"
	dbutil "github.com/cosmos/cosmos-sdk/db/internal"
	"github.com/cosmos/gorocksdb"
)

var (
	currentDBFileName    string = "current.db"
	checkpointFileFormat string = "%020d.db"
)

var (
	_ db.DBConnection = (*RocksDB)(nil)
	_ db.DBReader     = (*dbTxn)(nil)
	_ db.DBWriter     = (*dbWriter)(nil)
	_ db.DBReadWriter = (*dbWriter)(nil)
)

// RocksDB is a connection to a RocksDB key-value database.
type RocksDB = dbManager

type dbManager struct {
	current *dbConnection
	dir     string
	opts    dbOptions
	vmgr    *db.VersionManager
	mtx     sync.RWMutex
	// Track open DBWriters
	openWriters int32
	cpCache     checkpointCache
}

type dbConnection = gorocksdb.OptimisticTransactionDB

type checkpointCache struct {
	cache map[uint64]*cpCacheEntry
	mtx   sync.RWMutex
}

type cpCacheEntry struct {
	cxn       *dbConnection
	openCount uint
}

type dbTxn struct {
	txn     *gorocksdb.Transaction
	mgr     *dbManager
	version uint64
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
	dbo.OptimizeLevelStyleCompaction(1<<30 + 512<<20)

	opts := dbOptions{
		dbo: dbo,
		txo: gorocksdb.NewDefaultOptimisticTransactionOptions(),
		ro:  gorocksdb.NewDefaultReadOptions(),
		wo:  gorocksdb.NewDefaultWriteOptions(),
	}
	mgr := &dbManager{
		dir:     dir,
		opts:    opts,
		cpCache: checkpointCache{cache: map[uint64]*cpCacheEntry{}},
	}

	err := os.MkdirAll(mgr.checkpointsDir(), 0755)
	if err != nil {
		return nil, err
	}
	if mgr.vmgr, err = readVersions(mgr.checkpointsDir()); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dir, currentDBFileName)
	// if the current db file is missing but there are checkpoints, restore it
	if mgr.vmgr.Count() > 0 {
		if _, err = os.Stat(dbPath); os.IsNotExist(err) {
			err = mgr.restoreFromCheckpoint(mgr.vmgr.Last(), dbPath)
			if err != nil {
				return nil, err
			}
		}
	}
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
func readVersions(dir string) (*db.VersionManager, error) {
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
	return db.NewVersionManager(versions), nil
}

func (mgr *dbManager) checkpointPath(version uint64) (string, error) {
	dbPath := filepath.Join(mgr.checkpointsDir(), fmt.Sprintf(checkpointFileFormat, version))
	if stat, err := os.Stat(dbPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = db.ErrVersionDoesNotExist
		}
		return "", err
	} else if !stat.IsDir() {
		return "", db.ErrVersionDoesNotExist
	}
	return dbPath, nil
}

func (mgr *dbManager) openCheckpoint(version uint64) (*dbConnection, error) {
	mgr.cpCache.mtx.Lock()
	defer mgr.cpCache.mtx.Unlock()
	cp, has := mgr.cpCache.cache[version]
	if has {
		cp.openCount += 1
		return cp.cxn, nil
	}
	dbPath, err := mgr.checkpointPath(version)
	if err != nil {
		return nil, err
	}
	db, err := gorocksdb.OpenOptimisticTransactionDb(mgr.opts.dbo, dbPath)
	if err != nil {
		return nil, err
	}
	mgr.cpCache.cache[version] = &cpCacheEntry{cxn: db, openCount: 1}
	return db, nil
}

func (mgr *dbManager) Reader() db.DBReader {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	return &dbTxn{
		// Note: oldTransaction could be passed here as a small optimization to
		// avoid allocating a new object.
		txn: mgr.current.TransactionBegin(mgr.opts.wo, mgr.opts.txo, nil),
		mgr: mgr,
	}
}

func (mgr *dbManager) ReaderAt(version uint64) (db.DBReader, error) {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	d, err := mgr.openCheckpoint(version)
	if err != nil {
		return nil, err
	}

	return &dbTxn{
		txn:     d.TransactionBegin(mgr.opts.wo, mgr.opts.txo, nil),
		mgr:     mgr,
		version: version,
	}, nil
}

func (mgr *dbManager) ReadWriter() db.DBReadWriter {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	atomic.AddInt32(&mgr.openWriters, 1)
	return &dbWriter{dbTxn{
		txn: mgr.current.TransactionBegin(mgr.opts.wo, mgr.opts.txo, nil),
		mgr: mgr,
	}}
}

func (mgr *dbManager) Writer() db.DBWriter {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	atomic.AddInt32(&mgr.openWriters, 1)
	return mgr.newRocksDBBatch()
}

func (mgr *dbManager) Versions() (db.VersionSet, error) {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	return mgr.vmgr, nil
}

// SaveNextVersion implements DBConnection.
func (mgr *dbManager) SaveNextVersion() (uint64, error) {
	return mgr.save(0)
}

// SaveVersion implements DBConnection.
func (mgr *dbManager) SaveVersion(target uint64) error {
	if target == 0 {
		return db.ErrInvalidVersion
	}
	_, err := mgr.save(target)
	return err
}

func (mgr *dbManager) save(target uint64) (uint64, error) {
	mgr.mtx.Lock()
	defer mgr.mtx.Unlock()
	if mgr.openWriters > 0 {
		return 0, db.ErrOpenTransactions
	}
	newVmgr := mgr.vmgr.Copy()
	target, err := newVmgr.Save(target)
	if err != nil {
		return 0, err
	}
	cp, err := mgr.current.NewCheckpoint()
	if err != nil {
		return 0, err
	}
	dir := filepath.Join(mgr.checkpointsDir(), fmt.Sprintf(checkpointFileFormat, target))
	if err := cp.CreateCheckpoint(dir, 0); err != nil {
		return 0, err
	}
	cp.Destroy()
	mgr.vmgr = newVmgr
	return target, nil
}

func (mgr *dbManager) DeleteVersion(ver uint64) error {
	if mgr.cpCache.has(ver) {
		return db.ErrOpenTransactions
	}
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

func (mgr *dbManager) Revert() (err error) {
	mgr.mtx.RLock()
	defer mgr.mtx.RUnlock()
	if mgr.openWriters > 0 {
		return db.ErrOpenTransactions
	}
	// Close current connection and replace it with a checkpoint (created from the last checkpoint)
	mgr.current.Close()
	dbPath := filepath.Join(mgr.dir, currentDBFileName)
	err = os.RemoveAll(dbPath)
	if err != nil {
		return
	}
	if last := mgr.vmgr.Last(); last != 0 {
		err = mgr.restoreFromCheckpoint(last, dbPath)
		if err != nil {
			return
		}
	}
	mgr.current, err = gorocksdb.OpenOptimisticTransactionDb(mgr.opts.dbo, dbPath)
	return
}

func (mgr *dbManager) restoreFromCheckpoint(version uint64, path string) error {
	cxn, err := mgr.openCheckpoint(version)
	if err != nil {
		return err
	}
	defer mgr.cpCache.decrement(version)
	cp, err := cxn.NewCheckpoint()
	if err != nil {
		return err
	}
	err = cp.CreateCheckpoint(path, 0)
	if err != nil {
		return err
	}
	cp.Destroy()
	return nil
}

// Close implements DBConnection.
func (mgr *dbManager) Close() error {
	mgr.current.Close()
	mgr.opts.destroy()
	return nil
}

// Stats implements DBConnection.
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
	if tx.txn == nil {
		return nil, db.ErrTransactionClosed
	}
	if len(key) == 0 {
		return nil, db.ErrKeyEmpty
	}
	res, err := tx.txn.Get(tx.mgr.opts.ro, key)
	if err != nil {
		return nil, err
	}
	return moveSliceToBytes(res), nil
}

// Get implements DBReader.
func (tx *dbWriter) Get(key []byte) ([]byte, error) {
	if tx.txn == nil {
		return nil, db.ErrTransactionClosed
	}
	if len(key) == 0 {
		return nil, db.ErrKeyEmpty
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
	if tx.txn == nil {
		return db.ErrTransactionClosed
	}
	if err := dbutil.ValidateKv(key, value); err != nil {
		return err
	}
	return tx.txn.Put(key, value)
}

// Delete implements DBWriter.
func (tx *dbWriter) Delete(key []byte) error {
	if tx.txn == nil {
		return db.ErrTransactionClosed
	}
	if len(key) == 0 {
		return db.ErrKeyEmpty
	}
	return tx.txn.Delete(key)
}

func (tx *dbWriter) Commit() (err error) {
	if tx.txn == nil {
		return db.ErrTransactionClosed
	}
	defer func() { err = dbutil.CombineErrors(err, tx.Discard(), "Discard also failed") }()
	err = tx.txn.Commit()
	return
}

func (tx *dbTxn) Discard() error {
	if tx.txn == nil {
		return nil // Discard() is idempotent
	}
	defer func() { tx.txn.Destroy(); tx.txn = nil }()
	if tx.version == 0 {
		return nil
	}
	if !tx.mgr.cpCache.decrement(tx.version) {
		return fmt.Errorf("transaction has no corresponding checkpoint cache entry: %v", tx.version)
	}
	return nil
}

func (tx *dbWriter) Discard() error {
	if tx.txn != nil {
		defer atomic.AddInt32(&tx.mgr.openWriters, -1)
	}
	return tx.dbTxn.Discard()
}

// Iterator implements DBReader.
func (tx *dbTxn) Iterator(start, end []byte) (db.Iterator, error) {
	if tx.txn == nil {
		return nil, db.ErrTransactionClosed
	}
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, db.ErrKeyEmpty
	}
	itr := tx.txn.NewIterator(tx.mgr.opts.ro)
	return newRocksDBIterator(itr, start, end, false), nil
}

// ReverseIterator implements DBReader.
func (tx *dbTxn) ReverseIterator(start, end []byte) (db.Iterator, error) {
	if tx.txn == nil {
		return nil, db.ErrTransactionClosed
	}
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, db.ErrKeyEmpty
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

func (cpc *checkpointCache) has(ver uint64) bool {
	cpc.mtx.RLock()
	defer cpc.mtx.RUnlock()
	_, has := cpc.cache[ver]
	return has
}

func (cpc *checkpointCache) decrement(ver uint64) bool {
	cpc.mtx.Lock()
	defer cpc.mtx.Unlock()
	cp, has := cpc.cache[ver]
	if !has {
		return false
	}
	cp.openCount -= 1
	if cp.openCount == 0 {
		cp.cxn.Close()
		delete(cpc.cache, ver)
	}
	return true
}
