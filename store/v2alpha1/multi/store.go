package multi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"

	abci "github.com/tendermint/tendermint/abci/types"

	dbm "github.com/cosmos/cosmos-sdk/db"
	prefixdb "github.com/cosmos/cosmos-sdk/db/prefix"
	util "github.com/cosmos/cosmos-sdk/internal"
	dbutil "github.com/cosmos/cosmos-sdk/internal/db"
	"github.com/cosmos/cosmos-sdk/pruning"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	sdkmaps "github.com/cosmos/cosmos-sdk/store/internal/maps"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	types "github.com/cosmos/cosmos-sdk/store/v2alpha1"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/mem"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/smt"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/transient"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/kv"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

var (
	_ types.Queryable        = (*Store)(nil)
	_ types.CommitMultiStore = (*Store)(nil)
	_ types.CacheMultiStore  = (*cacheStore)(nil)
	_ types.MultiStore       = (*viewStore)(nil)
	_ types.KVStore          = (*substore)(nil)
)

var (
	ErrVersionDoesNotExist = errors.New("version does not exist")
	ErrMaximumHeight       = errors.New("maximum block height reached")

	// Root prefixes
	merkleRootKey = []byte{0} // Key for root hash of namespace tree
	schemaPrefix  = []byte{1} // Prefix for store keys (namespaces)
	contentPrefix = []byte{2} // Prefix for store contents

	// Per-substore prefixes
	substoreMerkleRootKey = []byte{0} // Key for root hashes of Merkle trees
	dataPrefix            = []byte{1} // Prefix for store data
	smtPrefix             = []byte{2} // Prefix for tree data
)

func ErrStoreNotFound(key string) error {
	return fmt.Errorf("store does not exist for key: %s", key)
}

// StoreParams is used to define a schema and other options and pass them to the MultiStore constructor.
type StoreParams struct {
	// Version pruning options for backing DBs.
	Pruning pruningtypes.PruningOptions
	// The minimum allowed version number.
	InitialVersion uint64
	// The optional backing DB to use for the state commitment Merkle tree data.
	// If nil, Merkle data is stored in the state storage DB under a separate prefix.
	StateCommitmentDB dbm.Connection
	// Contains the store schema and methods to modify it
	SchemaBuilder
	storeKeys
	// Inter-block persistent cache to use. TODO: not used/impl'd
	PersistentCache types.MultiStorePersistentCache
	// Any pending upgrades to apply on loading.
	Upgrades *types.StoreUpgrades
	// Contains The trace context and listeners that can also be set from store methods.
	*traceListenMixin
	substoreTraceListenMixins map[types.StoreKey]*traceListenMixin
}

// StoreSchema defineds a mapping of substore keys to store types
type StoreSchema map[string]types.StoreType
type StoreKeySchema map[types.StoreKey]types.StoreType

// storeKeys maps key names to StoreKey instances
type storeKeys map[string]types.StoreKey

// Store is the main persistent store type implementing CommitMultiStore.
// Substores consist of an SMT-based state commitment store and state storage.
// Substores must be reserved in the StoreParams or defined as part of a StoreUpgrade in order to be valid.
// Note:
// The state commitment data and proof are structured in the same basic pattern as the MultiStore, but use an SMT rather than IAVL tree:
// * The state commitment store of each substore consists of a independent SMT.
// * The state commitment of the root store consists of a Merkle map of all registered persistent substore names to the root hash of their corresponding SMTs
type Store struct {
	stateDB            dbm.Connection
	stateTxn           dbm.ReadWriter
	StateCommitmentDB  dbm.Connection
	stateCommitmentTxn dbm.ReadWriter

	schema StoreKeySchema

	mem  *mem.Store
	tran *transient.Store
	mtx  sync.RWMutex

	// Copied from StoreParams
	InitialVersion uint64
	*traceListenMixin

	pruningManager *pruning.Manager

	PersistentCache           types.MultiStorePersistentCache
	substoreCache             map[string]*substore
	substoreTraceListenMixins map[types.StoreKey]*traceListenMixin
}

type substore struct {
	root                 *Store
	name                 string
	dataBucket           dbm.ReadWriter
	stateCommitmentStore *smt.Store
	*traceListenMixin
}

// Builder type used to create a valid schema with no prefix conflicts
type SchemaBuilder struct {
	StoreSchema
	reserved []string
}

// Mixin type that to compose trace & listen state into each root store variant type
type traceListenMixin struct {
	listeners         map[types.StoreKey][]types.WriteListener
	TraceWriter       io.Writer
	TraceContext      types.TraceContext
	traceContextMutex sync.RWMutex
}

func newTraceListenMixin() *traceListenMixin {
	return &traceListenMixin{listeners: map[types.StoreKey][]types.WriteListener{}}
}

func newSchemaBuilder() SchemaBuilder {
	return SchemaBuilder{StoreSchema: StoreSchema{}}
}

// Returns true for valid store types for a MultiStore schema
func validSubStoreType(sst types.StoreType) bool {
	switch sst {
	case types.StoreTypePersistent:
		return true
	case types.StoreTypeMemory:
		return true
	case types.StoreTypeTransient:
		return true
	default:
		return false
	}
}

// Returns true iff both schema maps match exactly (including mem/tran stores)
func (ss StoreSchema) equal(that StoreSchema) bool {
	if len(ss) != len(that) {
		return false
	}
	for key, val := range that {
		myval, has := ss[key]
		if !has {
			return false
		}
		if val != myval {
			return false
		}
	}
	return true
}

func (this StoreSchema) matches(that StoreKeySchema) bool {
	if len(this) != len(that) {
		return false
	}
	for key, val := range that {
		myval, has := this[key.Name()]
		if !has {
			return false
		}
		if val != myval {
			return false
		}
	}
	return true
}

// Parses a schema from the DB
func readSavedSchema(bucket dbm.Reader) (*SchemaBuilder, error) {
	ret := newSchemaBuilder()
	it, err := bucket.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}
	for it.Next() {
		value := it.Value()
		if len(value) != 1 || !validSubStoreType(types.StoreType(value[0])) {
			return nil, fmt.Errorf("invalid mapping for store key: %v => %v", it.Key(), value)
		}
		ret.StoreSchema[string(it.Key())] = types.StoreType(value[0])
		ret.reserved = append(ret.reserved, string(it.Key())) // assume iter yields keys sorted
	}
	if err = it.Close(); err != nil {
		return nil, err
	}
	return &ret, nil
}

// NewStore constructs a MultiStore directly from a database.
// Creates a new store if no data exists; otherwise loads existing data.
func NewStore(db dbm.Connection, opts StoreParams) (ret *Store, err error) {
	pruningManager := pruning.NewManager()
	pruningManager.SetOptions(opts.Pruning)
	{ // load any pruned heights we missed from disk to be pruned on the next run
		r := db.Reader()
		defer r.Discard()
		tmdb := dbutil.ReadWriterAsTmdb(dbm.ReaderAsReadWriter(r))
		if err = pruningManager.LoadPruningHeights(tmdb); err != nil {
			return
		}
	}

	versions, err := db.Versions()
	if err != nil {
		return
	}
	// If the DB is not empty, attempt to load existing data
	if saved := versions.Count(); saved != 0 {
		if opts.InitialVersion != 0 && versions.Last() < opts.InitialVersion {
			return nil, fmt.Errorf("latest saved version is less than initial version: %v < %v",
				versions.Last(), opts.InitialVersion)
		}
	}
	// To abide by atomicity constraints, revert the DB to the last saved version, in case it contains
	// committed data in the "working" version.
	// This should only happen if Store.Commit previously failed.
	if err = db.Revert(); err != nil {
		return
	}
	stateTxn := db.ReadWriter()
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, stateTxn.Discard(), "stateTxn.Discard also failed")
		}
	}()
	stateCommitmentTxn := stateTxn
	if opts.StateCommitmentDB != nil {
		var scVersions dbm.VersionSet
		if scVersions, err = opts.StateCommitmentDB.Versions(); err != nil {
			return
		}
		// Version sets of each DB must match
		if !versions.Equal(scVersions) {
			err = fmt.Errorf("different version history between Storage and StateCommitment DB ")
			return
		}
		if err = opts.StateCommitmentDB.Revert(); err != nil {
			return
		}
		stateCommitmentTxn = opts.StateCommitmentDB.ReadWriter()
	}

	ret = &Store{
		stateDB:                   db,
		stateTxn:                  stateTxn,
		StateCommitmentDB:         opts.StateCommitmentDB,
		stateCommitmentTxn:        stateCommitmentTxn,
		mem:                       mem.NewStore(),
		tran:                      transient.NewStore(),
		substoreCache:             map[string]*substore{},
		traceListenMixin:          opts.traceListenMixin,
		PersistentCache:           opts.PersistentCache,
		pruningManager:            pruningManager,
		InitialVersion:            opts.InitialVersion,
		substoreTraceListenMixins: opts.substoreTraceListenMixins,
	}

	// Now load the substore schema
	schemaView := prefixdb.NewReader(ret.stateDB.Reader(), schemaPrefix)
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, schemaView.Discard(), "schemaView.Discard also failed")
			err = util.CombineErrors(err, ret.Close(), "base.Close also failed")
		}
	}()
	writeSchema := func(sch StoreSchema) {
		schemaWriter := prefixdb.NewWriter(ret.stateTxn, schemaPrefix)
		var it dbm.Iterator
		if it, err = schemaView.Iterator(nil, nil); err != nil {
			return
		}
		for it.Next() {
			err = schemaWriter.Delete(it.Key())
			if err != nil {
				return
			}
		}
		if err = it.Close(); err != nil {
			return
		}
		if err = schemaView.Discard(); err != nil {
			return
		}
		// NB. the migrated contents and schema are not committed until the next store.Commit
		for skey, typ := range sch {
			err = schemaWriter.Set([]byte(skey), []byte{byte(typ)})
			if err != nil {
				return
			}
		}
	}

	reg, err := readSavedSchema(schemaView)
	if err != nil {
		return
	}
	// If the loaded schema is empty (for new store), just copy the config schema;
	// Otherwise, migrate, then verify it is identical to the config schema
	if len(reg.StoreSchema) == 0 {
		writeSchema(opts.StoreSchema)
	} else {
		// Apply migrations to the schema
		if opts.Upgrades != nil {
			err = reg.migrateSchema(*opts.Upgrades)
			if err != nil {
				return
			}
		}
		if !reg.equal(opts.StoreSchema) {
			err = errors.New("loaded schema does not match configured schema")
			return
		}
		if opts.Upgrades != nil {
			err = migrateData(ret, *opts.Upgrades)
			if err != nil {
				return
			}
			writeSchema(opts.StoreSchema)
		}
	}
	ret.schema = StoreKeySchema{}
	for key, typ := range opts.StoreSchema {
		var skey types.StoreKey
		skey, err = opts.storeKey(key)
		if err != nil {
			return
		}
		ret.schema[skey] = typ
	}

	return
}

func (s *Store) Close() error {
	err := s.stateTxn.Discard()
	if s.StateCommitmentDB != nil {
		err = util.CombineErrors(err, s.stateCommitmentTxn.Discard(), "stateCommitmentTxn.Discard also failed")
	}
	return err
}

// Applies store upgrades to the DB contents.
func migrateData(store *Store, upgrades types.StoreUpgrades) error {
	// Get a view of current state to allow mutation while iterating
	reader := store.stateDB.Reader()
	scReader := reader
	if store.StateCommitmentDB != nil {
		scReader = store.StateCommitmentDB.Reader()
	}

	for _, key := range upgrades.Deleted {
		pfx := prefixSubstore(key)
		subReader := prefixdb.NewReader(reader, pfx)
		it, err := subReader.Iterator(nil, nil)
		if err != nil {
			return err
		}
		for it.Next() {
			if err = store.stateTxn.Delete(it.Key()); err != nil {
				return err
			}
		}
		it.Close()
		if store.StateCommitmentDB != nil {
			subReader = prefixdb.NewReader(scReader, pfx)
			it, err = subReader.Iterator(nil, nil)
			if err != nil {
				return err
			}
			for it.Next() {
				store.stateCommitmentTxn.Delete(it.Key())
			}
			if err = it.Close(); err != nil {
				return err
			}
		}
	}
	for _, rename := range upgrades.Renamed {
		oldPrefix := prefixSubstore(rename.OldKey)
		newPrefix := prefixSubstore(rename.NewKey)
		subReader := prefixdb.NewReader(reader, oldPrefix)
		subWriter := prefixdb.NewWriter(store.stateTxn, newPrefix)
		it, err := subReader.Iterator(nil, nil)
		if err != nil {
			return err
		}
		for it.Next() {
			subWriter.Set(it.Key(), it.Value())
		}
		if it.Close(); err != nil {
			return err
		}
		if store.StateCommitmentDB != nil {
			subReader = prefixdb.NewReader(scReader, oldPrefix)
			subWriter = prefixdb.NewWriter(store.stateCommitmentTxn, newPrefix)
			it, err = subReader.Iterator(nil, nil)
			if err != nil {
				return err
			}
			for it.Next() {
				subWriter.Set(it.Key(), it.Value())
			}
			if it.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

// encode key length as varint
func varintLen(l int) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, uint64(l))
	return buf[:n]
}

func prefixSubstore(key string) []byte {
	lv := varintLen(len(key))
	ret := append(lv, key...)
	return append(contentPrefix, ret...)
}

func prefixNonpersistent(key string) []byte {
	lv := varintLen(len(key))
	return append(lv, key...)
}

// GetKVStore implements MultiStore.
func (s *Store) GetKVStore(skey types.StoreKey) types.KVStore {
	key := skey.Name()
	var parent types.KVStore

	typ, has := s.schema[skey]
	if !has {
		panic(ErrStoreNotFound(key))
	}
	switch typ {
	case types.StoreTypeMemory:
		parent = s.mem
	case types.StoreTypeTransient:
		parent = s.tran
	case types.StoreTypePersistent:
	default:
		panic(fmt.Errorf("StoreType not supported: %v", typ)) // should never happen
	}

	var ret types.KVStore
	if parent != nil { // store is non-persistent
		ret = prefix.NewStore(parent, prefixNonpersistent(key))
	} else { // store is persistent
		sub, err := s.getSubstore(key)
		if err != nil {
			panic(err)
		}
		s.substoreCache[key] = sub
		ret = sub
	}
	// Wrap with trace/listen if needed. Note: we don't cache this, so users must get a new substore after
	// modifying tracers/listeners.
	return s.wrapTraceListen(ret, skey)
}

// HasKVStore implements MultiStore.
func (rs *Store) HasKVStore(skey types.StoreKey) bool {
	_, has := rs.schema[skey]
	return has
}

// Gets a persistent substore. This reads, but does not update the substore cache.
// Use it in cases where we need to access a store internally (e.g. read/write Merkle keys, queries)
func (s *Store) getSubstore(key string) (*substore, error) {
	if cached, has := s.substoreCache[key]; has {
		return cached, nil
	}
	pfx := prefixSubstore(key)
	stateRW := prefixdb.NewReadWriter(s.stateTxn, pfx)
	stateCommitmentRW := prefixdb.NewReadWriter(s.stateCommitmentTxn, pfx)
	var stateCommitmentStore *smt.Store

	rootHash, err := stateRW.Get(substoreMerkleRootKey)
	if err != nil {
		return nil, err
	}
	if rootHash != nil {
		stateCommitmentStore = loadSMT(stateCommitmentRW, rootHash)
	} else {
		smtdb := prefixdb.NewReadWriter(stateCommitmentRW, smtPrefix)
		stateCommitmentStore = smt.NewStore(smtdb)
	}

	return &substore{
		root:                 s,
		name:                 key,
		dataBucket:           prefixdb.NewReadWriter(stateRW, dataPrefix),
		stateCommitmentStore: stateCommitmentStore,
	}, nil
}

func (s *Store) SetSubstoreKVPair(skey types.StoreKey, kv, val []byte) {
	sub := s.GetKVStore(skey)
	sub.Set(kv, val)
}

// Resets a substore's state after commit (because root stateTxn has been discarded)
func (s *substore) refresh(rootHash []byte) {
	pfx := prefixSubstore(s.name)
	stateRW := prefixdb.NewReadWriter(s.root.stateTxn, pfx)
	stateCommitmentRW := prefixdb.NewReadWriter(s.root.stateCommitmentTxn, pfx)
	s.dataBucket = prefixdb.NewReadWriter(stateRW, dataPrefix)
	s.stateCommitmentStore = loadSMT(stateCommitmentRW, rootHash)
}

// Commit implements Committer.
func (s *Store) Commit() types.CommitID {
	// Substores read-lock this mutex; lock to prevent racey invalidation of underlying txns
	s.mtx.Lock()
	defer s.mtx.Unlock()
	// Determine the target version
	versions, err := s.stateDB.Versions()
	if err != nil {
		panic(err)
	}
	target := versions.Last() + 1
	if target > math.MaxInt64 {
		panic(ErrMaximumHeight)
	}
	// Fast forward to initial version if needed
	if s.InitialVersion != 0 && target < s.InitialVersion {
		target = s.InitialVersion
	}
	cid, err := s.commit(target)
	if err != nil {
		panic(err)
	}

	if err = s.handlePruning(cid.Version); err != nil {
		panic(err)
	}

	s.tran.Commit()
	return *cid
}

func (rs *Store) handlePruning(current int64) error {
	// Pass DB txn to pruning manager via adapter; running txns must be refreshed after this.
	// This is hacky but needed in order to restrict to a single txn (for memdb compatibility)
	// since the manager calls SetSync internally.
	rs.stateTxn.Discard()
	defer rs.refreshTransactions(true)
	db := rs.stateDB.ReadWriter()
	rs.pruningManager.HandleHeight(current-1, dbutil.ReadWriterAsTmdb(db)) // we should never prune the current version.
	db.Discard()
	if !rs.pruningManager.ShouldPruneAtHeight(current) {
		return nil
	}
	db = rs.stateDB.ReadWriter()
	defer db.Discard()
	pruningHeights, err := rs.pruningManager.GetFlushAndResetPruningHeights(dbutil.ReadWriterAsTmdb(db))
	if err != nil {
		return err
	}
	return pruneVersions(pruningHeights, func(ver int64) error {
		if err := rs.stateDB.DeleteVersion(uint64(ver)); err != nil {
			return fmt.Errorf("error pruning StateDB: %w", err)
		}

		if rs.StateCommitmentDB != nil {
			if err := rs.StateCommitmentDB.DeleteVersion(uint64(ver)); err != nil {
				return fmt.Errorf("error pruning StateCommitmentDB: %w", err)
			}
		}
		return nil
	})
}

// Performs necessary pruning via callback
func pruneVersions(heights []int64, prune func(int64) error) error {
	for _, height := range heights {
		if err := prune(height); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) getMerkleRoots() (ret map[string][]byte, err error) {
	ret = map[string][]byte{}
	for key := range s.schema {
		sub, has := s.substoreCache[key.Name()]
		if !has {
			sub, err = s.getSubstore(key.Name())
			if err != nil {
				return
			}
		}
		ret[key.Name()] = sub.stateCommitmentStore.Root()
	}
	return
}

func (s *Store) GetSubStoreProof(storeKeyName string) (*tmcrypto.ProofOp, []byte, error) {
	storeHashes, err := s.getMerkleRoots()
	if err != nil {
		return nil, nil, err
	}
	proofOp, err := types.ProofOpFromMap(storeHashes, storeKeyName)
	return &proofOp, storeHashes[storeKeyName], err
}

// Calculates root hashes and commits to DB. Does not verify target version or perform pruning.
func (s *Store) commit(target uint64) (id *types.CommitID, err error) {
	storeHashes, err := s.getMerkleRoots()
	if err != nil {
		return
	}
	// Update substore Merkle roots
	for key, storeHash := range storeHashes {
		w := prefixdb.NewReadWriter(s.stateTxn, prefixSubstore(key))
		if err = w.Set(substoreMerkleRootKey, storeHash); err != nil {
			return
		}
	}
	rootHash := sdkmaps.HashFromMap(storeHashes)
	if err = s.stateTxn.Set(merkleRootKey, rootHash); err != nil {
		return
	}
	if err = s.stateTxn.Commit(); err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, s.stateDB.Revert(), "stateDB.Revert also failed")
		}
	}()
	if err = s.stateDB.SaveVersion(target); err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = util.CombineErrors(err, s.stateTxn.Discard(), "stateTxn.Discard also failed")
		}
	}()
	// If DBs are not separate, StateCommitment state has been committed & snapshotted
	if s.StateCommitmentDB != nil {
		// if any error is encountered henceforth, we must revert the state and SC dbs
		defer func() {
			if err != nil {
				if delerr := s.stateDB.DeleteVersion(target); delerr != nil {
					err = fmt.Errorf("%w: commit rollback failed: %v", err, delerr)
				}
			}
		}()

		err = s.stateCommitmentTxn.Commit()
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				err = util.CombineErrors(err, s.StateCommitmentDB.Revert(), "stateCommitmentDB.Revert also failed")
			}
		}()

		if err = s.StateCommitmentDB.SaveVersion(target); err != nil {
			return
		}
	}

	// flush is complete, refresh our DB read/writers
	if err = s.refreshTransactions(false); err != nil {
		return
	}

	return &types.CommitID{Version: int64(target), Hash: rootHash}, nil
}

// Resets the txn objects in the store (does not discard current txns), then propagates
// them to cached substores.
// justState indicates we only need to refresh the stateDB txn.
func (s *Store) refreshTransactions(onlyState bool) error {
	s.stateTxn = s.stateDB.ReadWriter()
	if s.StateCommitmentDB != nil {
		if !onlyState {
			s.stateCommitmentTxn = s.StateCommitmentDB.ReadWriter()
		}
	} else {
		s.stateCommitmentTxn = s.stateTxn
	}

	storeHashes, err := s.getMerkleRoots()
	if err != nil {
		return err
	}
	// the state of all live substores must be refreshed
	for key, sub := range s.substoreCache {
		sub.refresh(storeHashes[key])
	}
	return nil
}

// LastCommitID implements Committer.
func (s *Store) LastCommitID() types.CommitID {
	versions, err := s.stateDB.Versions()
	if err != nil {
		panic(err)
	}
	last := versions.Last()
	if last == 0 {
		return types.CommitID{}
	}
	// Latest Merkle root is the one currently stored
	hash, err := s.stateTxn.Get(merkleRootKey)
	if err != nil {
		panic(err)
	}
	return types.CommitID{Version: int64(last), Hash: hash}
}

// SetInitialVersion implements CommitMultiStore.
func (s *Store) SetInitialVersion(version uint64) error {
	s.InitialVersion = version
	return nil
}

// GetVersion implements CommitMultiStore.
func (s *Store) GetVersion(version int64) (types.MultiStore, error) {
	return s.getView(version)
}

// CacheWrap implements MultiStore.
func (s *Store) CacheWrap() types.CacheMultiStore {
	return newCacheStore(s)
}

// GetAllVersions returns all available versions.
// https://github.com/cosmos/cosmos-sdk/pull/11124
func (s *Store) GetAllVersions() []uint64 {
	vs, err := s.stateDB.Versions()
	if err != nil {
		panic(err)
	}
	var ret []uint64
	for it := vs.Iterator(); it.Next(); {
		ret = append(ret, it.Value())
	}
	return ret
}

// PruneSnapshotHeight prunes the given height according to the prune strategy.
// If PruneNothing, this is a no-op.
// If other strategy, this height is persisted until it is
// less than <current height> - KeepRecent and <current height> % Interval == 0
func (rs *Store) PruneSnapshotHeight(height int64) {
	rs.stateTxn.Discard()
	defer rs.refreshTransactions(true)
	db := rs.stateDB.ReadWriter()
	defer db.Discard()
	rs.pruningManager.HandleHeightSnapshot(height, dbutil.ReadWriterAsTmdb(db))
}

// SetSnapshotInterval sets the interval at which the snapshots are taken.
// It is used by the store to determine which heights to retain until after the snapshot is complete.
func (rs *Store) SetSnapshotInterval(snapshotInterval uint64) {
	rs.pruningManager.SetSnapshotInterval(snapshotInterval)
}

// parsePath expects a format like /<storeName>[/<subpath>]
// Must start with /, subpath may be empty
// Returns error if it doesn't start with /
func parsePath(path string) (storeName string, subpath string, err error) {
	if !strings.HasPrefix(path, "/") {
		return storeName, subpath, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid path: %s", path)
	}

	paths := strings.SplitN(path[1:], "/", 2)
	storeName = paths[0]

	if len(paths) == 2 {
		subpath = "/" + paths[1]
	}

	return storeName, subpath, nil
}

// Query implements ABCI interface, allows queries.
//
// by default we will return from (latest height -1),
// as we will have merkle proofs immediately (header height = data height + 1)
// If latest-1 is not present, use latest (which must be present)
// if you care to have the latest data to see a tx results, you must
// explicitly set the height you want to see
func (s *Store) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	if len(req.Data) == 0 {
		return sdkerrors.QueryResult(sdkerrors.Wrap(sdkerrors.ErrTxDecode, "query cannot be zero length"), false)
	}

	// if height is 0, use the latest height
	height := req.Height
	if height == 0 {
		versions, err := s.stateDB.Versions()
		if err != nil {
			return sdkerrors.QueryResult(errors.New("failed to get version info"), false)
		}
		latest := versions.Last()
		if versions.Exists(latest - 1) {
			height = int64(latest - 1)
		} else {
			height = int64(latest)
		}
	}
	if height < 0 {
		return sdkerrors.QueryResult(fmt.Errorf("height overflow: %v", height), false)
	}
	res.Height = height

	storeName, subpath, err := parsePath(req.Path)
	if err != nil {
		return sdkerrors.QueryResult(sdkerrors.Wrapf(err, "failed to parse path"), false)
	}
	view, err := s.getView(height)
	if err != nil {
		if errors.Is(err, dbm.ErrVersionDoesNotExist) {
			err = sdkerrors.ErrInvalidHeight
		}
		return sdkerrors.QueryResult(sdkerrors.Wrapf(err, "failed to access height"), false)
	}

	substore, err := view.getSubstore(storeName)
	if err != nil {
		return sdkerrors.QueryResult(sdkerrors.Wrapf(err, "failed to access store: %s", storeName), false)
	}

	switch subpath {
	case "/key":
		var err error
		res.Key = req.Data // data holds the key bytes
		res.Value = substore.Get(res.Key)
		if !req.Prove {
			break
		}
		// TODO: actual IBC compatible proof. This is a placeholder so unit tests can pass
		res.ProofOps, err = substore.GetProof(res.Key)
		if err != nil {
			return sdkerrors.QueryResult(fmt.Errorf("merkle proof creation failed for key: %v", res.Key), false)
		}

	case "/subspace":
		res.Key = req.Data // data holds the subspace prefix

		pairs := kv.Pairs{
			Pairs: make([]kv.Pair, 0),
		}

		res.Key = req.Data // data holds the subspace prefix

		iterator := substore.Iterator(res.Key, types.PrefixEndBytes(res.Key))
		for ; iterator.Valid(); iterator.Next() {
			pairs.Pairs = append(pairs.Pairs, kv.Pair{Key: iterator.Key(), Value: iterator.Value()})
		}
		iterator.Close()

		bz, err := pairs.Marshal()
		if err != nil {
			panic(fmt.Errorf("failed to marshal KV pairs: %w", err))
		}

		res.Value = bz

	default:
		return sdkerrors.QueryResult(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unexpected query path: %v", req.Path), false)
	}

	return res
}

func loadSMT(stateCommitmentTxn dbm.ReadWriter, root []byte) *smt.Store {
	smtdb := prefixdb.NewReadWriter(stateCommitmentTxn, smtPrefix)
	return smt.LoadStore(smtdb, root)
}

// Returns closest index and whether it's a match
func binarySearch(hay []string, ndl string) (int, bool) {
	var mid int
	from, to := 0, len(hay)-1
	for from <= to {
		mid = (from + to) / 2
		switch strings.Compare(hay[mid], ndl) {
		case -1:
			from = mid + 1
		case 1:
			to = mid - 1
		default:
			return mid, true
		}
	}
	return from, false
}

// Migrates the state of the registry based on the upgrades
func (pr *SchemaBuilder) migrateSchema(upgrades types.StoreUpgrades) error {
	for _, key := range upgrades.Deleted {
		sst, ix, err := pr.storeInfo(key)
		if err != nil {
			return err
		}
		if sst != types.StoreTypePersistent {
			return fmt.Errorf("prefix is for non-persistent substore: %v (%v)", key, sst)
		}
		pr.reserved = append(pr.reserved[:ix], pr.reserved[ix+1:]...)
		delete(pr.StoreSchema, key)
	}
	for _, rename := range upgrades.Renamed {
		sst, ix, err := pr.storeInfo(rename.OldKey)
		if err != nil {
			return err
		}
		if sst != types.StoreTypePersistent {
			return fmt.Errorf("prefix is for non-persistent substore: %v (%v)", rename.OldKey, sst)
		}
		pr.reserved = append(pr.reserved[:ix], pr.reserved[ix+1:]...)
		delete(pr.StoreSchema, rename.OldKey)
		err = pr.registerName(rename.NewKey, types.StoreTypePersistent)
		if err != nil {
			return err
		}
	}
	for _, key := range upgrades.Added {
		err := pr.registerName(key, types.StoreTypePersistent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (reg *SchemaBuilder) storeInfo(key string) (sst types.StoreType, ix int, err error) {
	ix, has := binarySearch(reg.reserved, key)
	if !has {
		err = fmt.Errorf("name does not exist: %v", key)
		return
	}
	sst, has = reg.StoreSchema[key]
	if !has {
		err = fmt.Errorf("name is registered but not in schema: %v", key)
	}

	return
}

// registerName registers a store key by name only
func (reg *SchemaBuilder) registerName(key string, typ types.StoreType) error {
	// Find the neighboring reserved prefix, and check for duplicates and conflicts
	i, has := binarySearch(reg.reserved, key)
	if has {
		return fmt.Errorf("name already exists: %v", key)
	}
	// TODO auth vs authz ?
	// if i > 0 && strings.HasPrefix(key, reg.reserved[i-1]) {
	// 	return fmt.Errorf("name conflict: '%v' exists, cannot add '%v'", reg.reserved[i-1], key)
	// }
	// if i < len(reg.reserved) && strings.HasPrefix(reg.reserved[i], key) {
	// 	return fmt.Errorf("name conflict: '%v' exists, cannot add '%v'", reg.reserved[i], key)
	// }
	reserved := reg.reserved[:i]
	reserved = append(reserved, key)
	reg.reserved = append(reserved, reg.reserved[i:]...)
	reg.StoreSchema[key] = typ
	return nil
}

func (tlm *traceListenMixin) AddListeners(skey types.StoreKey, listeners []types.WriteListener) {
	tlm.listeners[skey] = append(tlm.listeners[skey], listeners...)
}

// ListeningEnabled returns if listening is enabled for a specific KVStore
func (tlm *traceListenMixin) ListeningEnabled(key types.StoreKey) bool {
	if ls, has := tlm.listeners[key]; has {
		return len(ls) != 0
	}
	return false
}

func (tlm *traceListenMixin) TracingEnabled() bool {
	return tlm.TraceWriter != nil
}

func (tlm *traceListenMixin) SetTracer(w io.Writer) {
	tlm.TraceWriter = w
}
func (tlm *traceListenMixin) SetTracingContext(tc types.TraceContext) {
	tlm.traceContextMutex.Lock()
	defer tlm.traceContextMutex.Unlock()
	if tlm.TraceContext != nil {
		for k, v := range tc {
			tlm.TraceContext[k] = v
		}
	} else {
		tlm.TraceContext = tc
	}
}

func (tlm *traceListenMixin) getTracingContext() types.TraceContext {
	tlm.traceContextMutex.Lock()
	defer tlm.traceContextMutex.Unlock()

	if tlm.TraceContext == nil {
		return nil
	}

	ctx := types.TraceContext{}
	for k, v := range tlm.TraceContext {
		ctx[k] = v
	}
	return ctx
}

func (s *Store) wrapTraceListen(store types.KVStore, skey types.StoreKey) types.KVStore {
	if s.TracingEnabled() {
		subStoreTlm := s.substoreTraceListenMixins[skey]
		store = tracekv.NewStore(store, subStoreTlm.TraceWriter, s.getTracingContext())
	}
	if s.ListeningEnabled(skey) {
		subStoreTlm := s.substoreTraceListenMixins[skey]
		store = listenkv.NewStore(store, skey, subStoreTlm.listeners[skey])
	}
	return store
}

func (s *Store) GetPruning() pruningtypes.PruningOptions {
	return s.pruningManager.GetOptions()
}

func (s *Store) SetPruning(po pruningtypes.PruningOptions) {
	s.pruningManager.SetOptions(po)
}

func (s *Store) GetSubstoreSMT(key string) *smt.Store {
	sub, err := s.getSubstore(key)
	if err != nil {
		panic(err)
	}
	return sub.stateCommitmentStore
}
