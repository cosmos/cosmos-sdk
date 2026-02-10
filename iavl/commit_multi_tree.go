package iavl

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	protoio "github.com/cosmos/gogoproto/io"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/store/mem"
	"cosmossdk.io/store/metrics"
	pruningtypes "cosmossdk.io/store/pruning/types"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	"cosmossdk.io/store/transient"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/iavl/internal"
	"github.com/cosmos/cosmos-sdk/iavl/internal/cachekv"
)

type CommitMultiTree struct {
	dir         string
	opts        Options
	stores      []*storeData                // always ordered by name
	iavlStores  []*storeData                // subset of stores that are IAVL, ordered by name
	otherStores []*storeData                // subset of stores that are not IAVL
	storesByKey map[storetypes.StoreKey]int // index of the stores by name

	commitMutex    sync.Mutex
	version        uint64
	lastCommitId   storetypes.CommitID
	lastCommitInfo *storetypes.CommitInfo

	pruningActive    atomic.Bool
	pruningOptions   pruningtypes.PruningOptions
	lastPruneVersion uint64
}

type storeData struct {
	key   storetypes.StoreKey
	typ   storetypes.StoreType
	store storetypes.CacheWrapper
}

func (db *CommitMultiTree) StartCommit(ctx context.Context, store storetypes.MultiStore, header cmtproto.Header) (storetypes.CommitFinalizer, error) {
	// TODO add mutex if needed
	multiTree, ok := store.(*internal.MultiTree)
	if !ok {
		return nil, fmt.Errorf("expected MultiTree, got %T", store)
	}

	if multiTree.LatestVersion() != int64(db.version) {
		return nil, fmt.Errorf("store version mismatch: expected %d, got %d", db.version, multiTree.LatestVersion())
	}

	numIavlStores := len(db.iavlStores)
	storeInfos := make([]storetypes.StoreInfo, numIavlStores)
	finalizers := make([]*commitTreeFinalizer, numIavlStores)
	commitInfo := &storetypes.CommitInfo{
		StoreInfos: storeInfos,
		Timestamp:  header.Time,
	}
	for i, si := range db.iavlStores {
		commitStore := si.store.(*CommitTree)
		cachedStore := multiTree.GetCacheWrapIfExists(si.key)
		var updates iter.Seq[KVUpdate]
		if cachedStore != nil {
			cacheKv, ok := cachedStore.(*cachekv.Store)
			if !ok {
				return nil, fmt.Errorf("expected cachekv.Store, got %T", cachedStore)
			}
			updates = cacheKv.Updates()
		}
		finalizer := commitStore.StartCommit(ctx, updates, 0)
		iavlFinalizer, ok := finalizer.(*commitTreeFinalizer)
		if !ok {
			return nil, fmt.Errorf("expected iavl commitTreeFinalizer, got %T", finalizer)
		}
		finalizers[i] = iavlFinalizer
		storeInfos[i].Name = si.key.Name()
	}
	ctx, cancel := context.WithCancel(ctx)
	finalizer := &multiTreeFinalizer{
		CommitMultiTree:    db,
		cacheMs:            multiTree,
		ctx:                ctx,
		cancel:             cancel,
		finalizers:         finalizers,
		workingCommitInfo:  commitInfo,
		done:               make(chan struct{}),
		hashReady:          make(chan struct{}),
		finalizeOrRollback: make(chan struct{}),
	}
	go func() {
		err := finalizer.commit(ctx)
		if err != nil {
			finalizer.err.Store(err)
		}
		close(finalizer.done)
		db.pruneIfNeeded() // start background pruning when needed
	}()
	return finalizer, nil
}

type multiTreeFinalizer struct {
	*CommitMultiTree
	ctx                context.Context
	cancel             context.CancelFunc
	cacheMs            *internal.MultiTree
	finalizers         []*commitTreeFinalizer
	workingCommitInfo  *storetypes.CommitInfo
	workingCommitId    storetypes.CommitID
	done               chan struct{}
	hashReady          chan struct{}
	finalizeOnce       sync.Once
	finalizeOrRollback chan struct{}
	err                atomic.Value
}

func (db *multiTreeFinalizer) commit(ctx context.Context) error {
	db.commitMutex.Lock()
	defer db.commitMutex.Unlock()

	stagedVersion := db.stagedVersion()
	ctx, span := tracer.Start(ctx, "CommitMultiTree.Commit",
		trace.WithAttributes(
			attribute.Int64("version", int64(stagedVersion)),
		),
	)
	defer span.End()

	// start writing commit info in background
	commitInfoSynced := make(chan error, 1)
	go func() {
		db.writeCommitInfo(commitInfoSynced)
	}()

	if err := db.prepareCommit(ctx); err != nil {
		db.startRollback()

		// do not use an errGroup here since, we want to rollback everything even if some rollbacks fail
		var wg sync.WaitGroup
		errs := make([]error, len(db.finalizers))
		for i, finalizer := range db.finalizers {
			i, finalizer := i, finalizer
			wg.Add(1)
			go func() {
				defer wg.Done()
				errs[i] = finalizer.Rollback()
			}()
		}

		wg.Wait()
		if err := errors.Join(errs...); err != nil {
			return fmt.Errorf("rollback failed: %w", err.(error))
		}

		return fmt.Errorf("%w; cause: %v", rolledbackErr, err)
	}

	var errGroup errgroup.Group
	// finalize IAVL stores
	for _, finalizer := range db.finalizers {
		finalizer := finalizer
		errGroup.Go(func() error {
			_, err := finalizer.Finalize()
			return err
		})
	}
	// commit non-IAVL stores
	for _, si := range db.otherStores {
		si := si
		errGroup.Go(func() error {
			cachedStore := db.cacheMs.GetCacheWrapIfExists(si.key)
			if cachedStore == nil {
				return nil
			}
			cachedStore.Write()
			committer, ok := si.store.(storetypes.Committer)
			if !ok {
				return nil
			}
			committer.Commit()
			return nil
		})
	}
	// wait for all stores to finalize
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("finalizing commit failed: %w", err)
	}

	// wait for commit info to be written
	if err := <-commitInfoSynced; err != nil {
		return fmt.Errorf("writing commit info failed: %w", err)
	}

	db.lastCommitId = db.workingCommitId
	db.lastCommitInfo = db.workingCommitInfo
	db.version++
	return nil
}

func (db *multiTreeFinalizer) writeCommitInfo(headerDone chan error) {
	// in order to not block on fsync until AFTER we have computed all hashes, which SHOULD be the slowest operation (WAL writing should complete before that)
	// we write and fsync the first part of the commit info (store names) as soon as we know finalization will happen,
	// and then append hashes at the end once they are ready, without fsyncing again since they aren't needed for durability

	file, err := db.writeCommitInfoHeader()
	headerDone <- err
	close(headerDone)
	if err != nil {
		return
	}

	// wait for hashes to be ready
	<-db.hashReady

	info := db.workingCommitInfo

	var scratchBuf [binary.MaxVarintLen64]byte

	// append each store hash to the file
	for _, storeInfo := range info.StoreInfos {
		// length-prefixed hash
		hashLen := uint64(len(storeInfo.CommitId.Hash))
		n := binary.PutUvarint(scratchBuf[:], hashLen)
		_, err := file.Write(scratchBuf[:n])
		if err != nil {
			logger.Error("failed to write commit info store info hash length", "error", err)
			return
		}

		_, err = file.Write(storeInfo.CommitId.Hash)
		if err != nil {
			logger.Error("failed to write commit info store info hash", "error", err)
			return
		}
	}

	err = file.Close()
	if err != nil {
		logger.Error("failed to close commit info file after writing hashes", "error", err)
		return
	}
}

func (db *multiTreeFinalizer) writeCommitInfoHeader() (*os.File, error) {
	var headerBuf bytes.Buffer
	// write version as litte-endian uint32
	stagedVersion := db.stagedVersion()
	var scratchBuf [binary.MaxVarintLen64]byte
	binary.LittleEndian.PutUint32(scratchBuf[:4], uint32(stagedVersion))
	_, err := headerBuf.Write(scratchBuf[:4])
	if err != nil {
		return nil, fmt.Errorf("failed to write commit info version: %w", err)
	}

	info := db.workingCommitInfo

	// write timestamp as unix nano int64
	binary.LittleEndian.PutUint64(scratchBuf[:8], uint64(info.Timestamp.UnixNano()))
	_, err = headerBuf.Write(scratchBuf[:8])
	if err != nil {
		return nil, fmt.Errorf("failed to write commit info timestamp: %w", err)
	}

	// write the number of store infos as little-endian uint32
	binary.LittleEndian.PutUint32(scratchBuf[:4], uint32(len(info.StoreInfos)))
	_, err = headerBuf.Write(scratchBuf[:4])
	if err != nil {
		return nil, fmt.Errorf("failed to write commit info store info count: %w", err)
	}

	// write each store name as a length-prefixed string
	for _, storeInfo := range info.StoreInfos {
		// varint length prefix
		nameLen := uint64(len(storeInfo.Name))
		n := binary.PutUvarint(scratchBuf[:], nameLen)
		_, err := headerBuf.Write(scratchBuf[:n])
		if err != nil {
			return nil, fmt.Errorf("failed to write commit info store info name length: %w", err)
		}
		_, err = headerBuf.Write([]byte(storeInfo.Name))
		if err != nil {
			return nil, fmt.Errorf("failed to write commit info store info name: %w", err)
		}
	}

	// wait for finalization signal
	<-db.finalizeOrRollback
	if db.ctx.Err() != nil {
		return nil, db.ctx.Err() // do not write commit info if rolling back
	}

	// write the header to disk
	commitInfoDir := filepath.Join(db.dir, commitInfoSubPath)
	err = os.MkdirAll(commitInfoDir, 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit info dir: %w", err)
	}

	commitInfoPath := filepath.Join(commitInfoDir, fmt.Sprintf("%d", stagedVersion))
	file, err := os.OpenFile(commitInfoPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open commit info file for version %d: %w", stagedVersion, err)
	}

	_, err = file.Write(headerBuf.Bytes())
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to write commit info header for version %d: %w", stagedVersion, err)
	}

	// fsync the file to ensure durability of store names
	err = file.Sync()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to sync commit info file for version %d: %w", stagedVersion, err)
	}

	// TODO optionally fsync the directory as well for extra durability guarantees

	return file, nil
}

func (db *multiTreeFinalizer) prepareCommit(ctx context.Context) error {
	var hashErrGroup errgroup.Group
	for i, finalizer := range db.finalizers {
		finalizer := finalizer
		hashErrGroup.Go(func() error {
			hash, err := finalizer.WaitForHash()
			if err != nil {
				return err
			}
			if hash.Version != 0 && hash.Version != int64(db.stagedVersion()) {
				return fmt.Errorf("store %s returned mismatched version in commit ID: expected %d, got %d", db.iavlStores[i].key.Name(), db.stagedVersion(), hash.Version)
			}
			db.workingCommitInfo.StoreInfos[i].CommitId = hash
			return nil
		})
	}
	if err := hashErrGroup.Wait(); err != nil {
		return err
	}

	db.workingCommitId = storetypes.CommitID{
		Version: int64(db.stagedVersion()),
		Hash:    db.workingCommitInfo.Hash(),
	}
	close(db.hashReady)

	<-db.finalizeOrRollback

	return ctx.Err()
}

func (db *multiTreeFinalizer) PrepareFinalize() (storetypes.CommitID, error) {
	if err := db.SignalFinalize(); err != nil {
		return storetypes.CommitID{}, err
	}
	select {
	case <-db.hashReady:
	case <-db.done:
	}
	err := db.err.Load()
	if err != nil {
		return storetypes.CommitID{}, err.(error)
	}
	return db.workingCommitId, nil
}

func (db *multiTreeFinalizer) SignalFinalize() error {
	db.finalizeOnce.Do(func() {
		close(db.finalizeOrRollback)
	})
	return nil
}

func (db *multiTreeFinalizer) Finalize() (storetypes.CommitID, error) {
	if err := db.SignalFinalize(); err != nil {
		return storetypes.CommitID{}, err
	}

	<-db.done
	err := db.err.Load()
	if err != nil {
		return storetypes.CommitID{}, err.(error)
	}
	return db.workingCommitId, nil
}

func (db *multiTreeFinalizer) Rollback() error {
	db.startRollback()
	<-db.done
	err := db.err.Load()
	if err == nil {
		return fmt.Errorf("rollback failed, commit succeeded")
	}
	if !errors.Is(err.(error), rolledbackErr) {
		return fmt.Errorf("rollback failed: %w", err.(error))
	}
	return nil
}

func (db *multiTreeFinalizer) startRollback() {
	// we must propagate cancellation to any background operations
	db.cancel()
	db.finalizeOnce.Do(func() {
		close(db.finalizeOrRollback)
	})
}

func (db *CommitMultiTree) GetCommitInfo(ver int64) (*storetypes.CommitInfo, error) {
	return loadCommitInfo(db.dir, uint64(ver))
}

func (db *CommitMultiTree) Commit() storetypes.CommitID {
	panic("cannot call Commit on uncached CommitMultiTree directly; use StartCommit")
}

func (db *CommitMultiTree) WorkingHash() []byte {
	panic("cannot call PrepareFinalize on uncached CommitMultiTree directly; use StartCommit")
}

func (db *CommitMultiTree) LastCommitID() storetypes.CommitID {
	return db.lastCommitId
}

const commitInfoSubPath = "commit_info"

// loadLatestCommitInfo loads the highest version number commit info file from the commit_info directory
// if any exist, returning the version and CommitInfo.
func loadLatestCommitInfo(dir string) (uint64, *storetypes.CommitInfo, error) {
	commitInfoDir := filepath.Join(dir, commitInfoSubPath)
	err := os.MkdirAll(commitInfoDir, 0o700)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create commit info dir: %w", err)
	}

	entries, err := os.ReadDir(commitInfoDir)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read commit info dir: %w", err)
	}

	// find the latest version by looking for the highest numbered file
	var latestVersion uint64
	for _, entry := range entries {
		var version uint64
		_, err := fmt.Sscanf(entry.Name(), "%d", &version)
		if err != nil {
			// skip non-numeric files
			continue
		}
		if version > latestVersion {
			latestVersion = version
		}
	}

	if latestVersion == 0 {
		// no versions found, no commit info to load
		return 0, nil, nil
	}

	commitInfo, err := loadCommitInfo(dir, latestVersion)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to load commit info for version %d: %w", latestVersion, err)
	}

	return latestVersion, commitInfo, nil
}

func loadCommitInfo(dir string, version uint64) (*storetypes.CommitInfo, error) {
	commitInfoDir := filepath.Join(dir, commitInfoSubPath)
	err := os.MkdirAll(commitInfoDir, 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit info dir: %w", err)
	}

	commitInfoPath := filepath.Join(commitInfoDir, fmt.Sprintf("%d", version))
	bz, err := os.ReadFile(commitInfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit info file for version %d: %w", version, err)
	}

	rdr := bytes.NewReader(bz)

	// read version
	var storedVersion uint32
	err = binary.Read(rdr, binary.LittleEndian, &storedVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit info version for version %d: %w", version, err)
	}
	if uint64(storedVersion) != version {
		return nil, fmt.Errorf("commit info version mismatch: expected %d, got %d", version, storedVersion)
	}

	// read timestamp
	var timestampNano uint64
	err = binary.Read(rdr, binary.LittleEndian, &timestampNano)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit info timestamp for version %d: %w", version, err)
	}

	// read store count
	var storeCount uint32
	err = binary.Read(rdr, binary.LittleEndian, &storeCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit info store count for version %d: %w", version, err)
	}

	commitInfo := &storetypes.CommitInfo{
		StoreInfos: make([]storetypes.StoreInfo, storeCount),
		Timestamp:  time.Unix(0, int64(timestampNano)),
		Version:    int64(version),
	}

	// read each store info
	for i := uint32(0); i < storeCount; i++ {
		// read name length
		nameLen, err := binary.ReadUvarint(rdr)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit info store info name length for version %d: %w", version, err)
		}
		nameBytes := make([]byte, nameLen)
		_, err = io.ReadFull(rdr, nameBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit info store info name for version %d: %w", version, err)
		}
		commitInfo.StoreInfos[i].Name = string(nameBytes)
	}

	// TODO handle cases where we have no hashes (pre-hash commit infos), for now we just error

	// read each store hash
	for i := uint32(0); i < storeCount; i++ {
		// read hash length
		hashLen, err := binary.ReadUvarint(rdr)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit info store info hash length for version %d: %w", version, err)
		}

		hashBytes := make([]byte, hashLen)
		_, err = io.ReadFull(rdr, hashBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read commit info store info hash for version %d: %w", version, err)
		}

		commitInfo.StoreInfos[i].CommitId = storetypes.CommitID{
			Version: int64(version),
			Hash:    hashBytes,
		}
	}

	return commitInfo, nil
}

func (db *CommitMultiTree) SetPruning(opts pruningtypes.PruningOptions) {
	db.pruningOptions = opts
}

func (db *CommitMultiTree) GetPruning() pruningtypes.PruningOptions {
	return db.pruningOptions
}

func (db *CommitMultiTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

func (db *CommitMultiTree) GetStore(storetypes.StoreKey) storetypes.Store {
	panic("cannot call GetStore on uncached CommitMultiTree directly; use CacheMultiStore first")
}

func (db *CommitMultiTree) GetKVStore(storetypes.StoreKey) storetypes.KVStore {
	panic("cannot call GetKVStore on uncached CommitMultiTree directly; use CacheMultiStore first")
}

// GetObjKVStore returns a mounted ObjKVStore for a given StoreKey.
func (db *CommitMultiTree) GetObjKVStore(storetypes.StoreKey) storetypes.ObjKVStore {
	panic("cannot call GetObjKVStore on uncached CommitMultiTree directly; use CacheMultiStore first")
}

func (db *CommitMultiTree) GetCommitStore(storetypes.StoreKey) storetypes.CommitStore {
	panic("cannot call GetCommitStore on uncached CommitMultiTree directly; use CacheMultiStore first")
}

func (db *CommitMultiTree) GetCommitKVStore(storetypes.StoreKey) storetypes.CommitKVStore {
	panic("cannot call GetCommitKVStore on uncached CommitMultiTree directly; use CacheMultiStore first")
}

func (db *CommitMultiTree) SetTracer(io.Writer) storetypes.MultiStore {
	logger.Warn("SetTracer is not implemented for CommitMultiTree")
	return db
}

func (db *CommitMultiTree) SetTracingContext(storetypes.TraceContext) storetypes.MultiStore {
	logger.Warn("SetTracingContext is not implemented for CommitMultiTree")
	return db
}

func (db *CommitMultiTree) TracingEnabled() bool {
	return false
}

func (db *CommitMultiTree) Snapshot(height uint64, protoWriter protoio.Writer) error {
	return fmt.Errorf("snapshotting has not been implemented yet")
}

func (db *CommitMultiTree) PruneSnapshotHeight(height int64) {
	logger.Warn("PruneSnapshotHeight is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) SetSnapshotInterval(snapshotInterval uint64) {
	logger.Warn("SetSnapshotInterval is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) Restore(height uint64, format uint32, protoReader protoio.Reader) (snapshottypes.SnapshotItem, error) {
	return snapshottypes.SnapshotItem{}, fmt.Errorf("restoring from snapshot has not been implemented yet")
}

func (db *CommitMultiTree) MountStoreWithDB(key storetypes.StoreKey, typ storetypes.StoreType, _ dbm.DB) {
	if _, exists := db.storesByKey[key]; exists {
		panic(fmt.Sprintf("store with key %s already mounted", key.Name()))
	}
	db.storesByKey[key] = -1 // we assign actual index when loading
	db.stores = append(db.stores, &storeData{
		key: key,
		typ: typ,
	})
}

func (db *CommitMultiTree) LoadLatestVersion() error {
	// sort treeKeys to ensure deterministic order
	// we assume that MountStoreWithDB has been called for all stores before this
	// so treeKeys and storeTypes are aligned by index
	slices.SortFunc(db.stores, func(a, b *storeData) int {
		return bytes.Compare([]byte(a.key.Name()), []byte(b.key.Name()))
	})

	version, ci, err := loadLatestCommitInfo(db.dir)
	if err != nil {
		return fmt.Errorf("failed to load latest commit info: %w", err)
	}

	for i, si := range db.stores {
		key := si.key
		storeType := si.typ
		store, err := db.loadStore(si.key, storeType, version)
		if err != nil {
			return fmt.Errorf("failed to load store %s: %w", key.Name(), err)
		}
		si.store = store
		db.storesByKey[key] = i
		if _, ok := store.(*CommitTree); ok {
			db.iavlStores = append(db.iavlStores, si)
		} else {
			db.otherStores = append(db.otherStores, si)
		}
	}

	db.version = version
	if ci != nil {
		// should be nil on initial creation
		db.lastCommitId = storetypes.CommitID{
			Version: int64(version),
			Hash:    ci.Hash(),
		}
		db.lastCommitInfo = ci
	}

	return nil
}

func (db *CommitMultiTree) loadStore(key storetypes.StoreKey, typ storetypes.StoreType, expectedVersion uint64) (storetypes.CacheWrapper, error) {
	switch typ {
	case storetypes.StoreTypeIAVL, storetypes.StoreTypeDB:
		dir := filepath.Join(db.dir, "stores", fmt.Sprintf(key.Name(), ".iavl"))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0o755)
			if err != nil {
				return nil, fmt.Errorf("failed to create store dir %s: %w", dir, err)
			}
		}
		ct, err := NewCommitTree(dir, db.opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load CommitTree for store %s: %w", key.Name(), err)
		}
		if uint64(ct.LatestVersion()) != expectedVersion {
			return nil, fmt.Errorf("store %s version mismatch: expected %d, got %d", key.Name(), expectedVersion, ct.LatestVersion())
		}
		return ct, nil
	case storetypes.StoreTypeTransient:
		_, ok := key.(*storetypes.TransientStoreKey)
		if !ok {
			return nil, fmt.Errorf("invalid StoreKey for StoreTypeTransient: %s", key.String())
		}

		return transient.NewStore(), nil
	case storetypes.StoreTypeMemory:
		if _, ok := key.(*storetypes.MemoryStoreKey); !ok {
			return nil, fmt.Errorf("unexpected key type for a MemoryStoreKey; got: %s", key.String())
		}

		return mem.NewStore(), nil
	case storetypes.StoreTypeObject:
		if _, ok := key.(*storetypes.ObjectStoreKey); !ok {
			return nil, fmt.Errorf("unexpected key type for a ObjectStoreKey: %s", key.String())
		}
		return transient.NewObjStore(), nil
	default:
		return nil, fmt.Errorf("unsupported store type: %s", typ.String())
	}
}

func (db *CommitMultiTree) LoadLatestVersionAndUpgrade(upgrades *storetypes.StoreUpgrades) error {
	return fmt.Errorf("LoadLatestVersionAndUpgrade has not been implemented yet")
}

func (db *CommitMultiTree) LoadVersionAndUpgrade(ver int64, upgrades *storetypes.StoreUpgrades) error {
	return fmt.Errorf("LoadVersionAndUpgrade has not been implemented yet")
}

func (db *CommitMultiTree) LoadVersion(ver int64) error {
	return fmt.Errorf("LoadVersion has not been implemented yet")
}

func (db *CommitMultiTree) SetInterBlockCache(cache storetypes.MultiStorePersistentCache) {
	logger.Warn("SetInterBlockCache is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) SetInitialVersion(version int64) error {
	return fmt.Errorf("SetInitialVersion has not been implemented yet")
}

func (db *CommitMultiTree) SetIAVLCacheSize(size int) {}

func (db *CommitMultiTree) SetIAVLDisableFastNode(disable bool) {}

func (db *CommitMultiTree) SetIAVLSyncPruning(sync bool) {}

func (db *CommitMultiTree) RollbackToVersion(version int64) error {
	return fmt.Errorf("RollbackToVersion has not been implemented yet")
}

func (db *CommitMultiTree) ListeningEnabled(key storetypes.StoreKey) bool {
	logger.Warn("ListeningEnabled is not implemented for CommitMultiTree")
	return false
}

func (db *CommitMultiTree) AddListeners(keys []storetypes.StoreKey) {
	logger.Warn("AddListeners is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) PopStateCache() []*storetypes.StoreKVPair {
	// TODO implement me
	panic("implement me")
}

func (db *CommitMultiTree) SetMetrics(metrics metrics.StoreMetrics) {
	logger.Warn("SetMetrics is not implemented for CommitMultiTree")
}

func LoadCommitMultiTree(path string, opts Options) (*CommitMultiTree, error) {
	db := &CommitMultiTree{
		dir:         path,
		opts:        opts,
		storesByKey: make(map[storetypes.StoreKey]int),
	}
	return db, nil
}

func (db *CommitMultiTree) stagedVersion() uint64 {
	return db.version + 1
}

func (db *CommitMultiTree) LatestVersion() int64 {
	return int64(db.version)
}

func (db *CommitMultiTree) Close() error {
	for _, si := range db.stores {
		if closer, ok := si.store.(io.Closer); ok {
			err := closer.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *CommitMultiTree) CacheWrap() storetypes.CacheWrap {
	return db.CacheMultiStore()
}

func (db *CommitMultiTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	logger.Warn("CacheWrapWithTrace is not implemented for CommitMultiTree; falling back to CacheWrap")
	return db.CacheWrap()
}

func (db *CommitMultiTree) CacheMultiStore() storetypes.CacheMultiStore {
	// TODO we need to make sure each cached tree has the correct version!!
	// as is they will always get the latest version no matter how long the CacheMultiStore lives
	// which is incorrect if this outlives a commit and violates concurrency safety
	version := int64(db.version)
	return internal.NewMultiTree(version, func(key storetypes.StoreKey) storetypes.CacheWrap {
		idx, ok := db.storesByKey[key]
		if !ok {
			panic(fmt.Sprintf("store with key %s not mounted", key.Name()))
		}
		tree := db.stores[idx].store
		return tree.CacheWrap()
	})
}

func (db *CommitMultiTree) CacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	if version == 0 {
		// use latest version
		return db.CacheMultiStore(), nil
	}

	mt := internal.NewMultiTree(version, func(key storetypes.StoreKey) storetypes.CacheWrap {
		idx, ok := db.storesByKey[key]
		if !ok {
			panic(fmt.Sprintf("store with key %s not mounted", key.Name()))
		}
		tree := db.stores[idx].store
		switch tree := tree.(type) {
		case *CommitTree:
			// it's not really ideal to panic here, but the MultiTree interface doesn't allow for error returns
			// alternatively we can check out all historical roots aggressively, but that may be unnecessary when
			// historical queries will often touch a single store
			// a better approach may be to keep some global tracking of historical roots in the CommitMultiTree itself
			// by pruning old commit_info files when pruning is implemented
			t, err := tree.GetVersion(version)
			if err != nil {
				panic(fmt.Sprintf("failed to get version %d for store %s: %v", version, key.Name(), err))
			}
			return t.CacheWrap()
		default:
			return tree.CacheWrap()
		}
	})

	return mt, nil
}

func (db *CommitMultiTree) pruneIfNeeded() {
	if db.pruningActive.Load() {
		// already pruning, skip
		return
	}
	intervalSinceLastPrune := db.version - db.lastPruneVersion
	if db.pruningOptions.Interval > intervalSinceLastPrune {
		// not time to prune yet
		return
	}
	db.pruningActive.Store(true)
	go func() {
		ctx, span := tracer.Start(context.Background(), "CommitMultiTree.Prune")
		defer span.End()
		defer db.pruningActive.Store(false)
		for _, si := range db.iavlStores {
			ctx, span := tracer.Start(ctx, "PruneStore",
				trace.WithAttributes(
					attribute.String("store", si.key.Name()),
				),
			)
			ct, ok := si.store.(*CommitTree)
			if !ok {
				logger.Error("expected CommitTree store, got %T", si.store)
				continue
			}
			err := ct.Prune(ctx, db.pruningOptions)
			if err != nil {
				logger.Error("failed to prune store %s: %v", si.key.Name(), err)
				continue
			}
			span.End()
		}
	}()
}

var _ storetypes.CommitMultiStore2 = &CommitMultiTree{}
