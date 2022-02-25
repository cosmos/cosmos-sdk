package multi

import (
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
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	sdkmaps "github.com/cosmos/cosmos-sdk/store/internal/maps"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	types "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/mem"
	"github.com/cosmos/cosmos-sdk/store/v2/smt"
	"github.com/cosmos/cosmos-sdk/store/v2/transient"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/kv"
	protoio "github.com/gogo/protobuf/io"
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
	dataPrefix            = []byte{1} // Prefix for state mappings
	indexPrefix           = []byte{2} // Prefix for Store reverse index
	smtPrefix             = []byte{3} // Prefix for SMT data
)

func ErrStoreNotFound(key string) error {
	return fmt.Errorf("store does not exist for key: %s", key)
}

// StoreParams is used to define a schema and other options and pass them to the MultiStore constructor.
type StoreParams struct {
	// Version pruning options for backing DBs.
	Pruning types.PruningOptions
	// The minimum allowed version number.
	InitialVersion uint64
	// The optional backing DB to use for the state commitment Merkle tree data.
	// If nil, Merkle data is stored in the state storage DB under a separate prefix.
	StateCommitmentDB dbm.DBConnection
	// Contains the store schema and methods to modify it
	SchemaBuilder
	storeKeys
	// Inter-block persistent cache to use. TODO: not implemented
	PersistentCache types.MultiStorePersistentCache
	// Any pending upgrades to apply on loading.
	Upgrades *types.StoreUpgrades
	// Contains The trace context and listeners that can also be set from store methods.
	*traceListenMixin
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
	stateDB            dbm.DBConnection
	stateTxn           dbm.DBReadWriter
	StateCommitmentDB  dbm.DBConnection
	stateCommitmentTxn dbm.DBReadWriter

	schema StoreKeySchema

	mem  *mem.Store
	tran *transient.Store
	mtx  sync.RWMutex

	// Copied from StoreParams
	Pruning        types.PruningOptions
	InitialVersion uint64
	*traceListenMixin

	PersistentCache types.MultiStorePersistentCache
	substoreCache   map[string]*substore
}

type substore struct {
	root                 *Store
	name                 string
	dataBucket           dbm.DBReadWriter
	indexBucket          dbm.DBReadWriter
	stateCommitmentStore *smt.Store
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

// DefaultStoreParams returns a MultiStore config with an empty schema, a single backing DB,
// pruning with PruneDefault, no listeners and no tracer.
func DefaultStoreParams() StoreParams {
	return StoreParams{
		Pruning:          types.PruneDefault,
		SchemaBuilder:    newSchemaBuilder(),
		storeKeys:        storeKeys{},
		traceListenMixin: newTraceListenMixin(),
	}
}

// func (pr *SchemaBuilder) registerName(key string, typ types.StoreType) error {
func (par *StoreParams) RegisterSubstore(skey types.StoreKey, typ types.StoreType) error {
	if err := par.registerName(skey.Name(), typ); err != nil {
		return err
	}
	par.storeKeys[skey.Name()] = skey
	return nil
}

func (par *StoreParams) storeKey(key string) (types.StoreKey, error) {
	skey, ok := par.storeKeys[key]
	if !ok {
		return nil, fmt.Errorf("StoreKey instance not mapped: %s", key)
	}
	return skey, nil
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
func (this StoreSchema) equal(that StoreSchema) bool {
	if len(this) != len(that) {
		return false
	}
	for key, val := range that {
		myval, has := this[key]
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
func readSavedSchema(bucket dbm.DBReader) (*SchemaBuilder, error) {
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
func NewStore(db dbm.DBConnection, opts StoreParams) (ret *Store, err error) {
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
	err = db.Revert()
	if err != nil {
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
		scVersions, err = opts.StateCommitmentDB.Versions()
		if err != nil {
			return
		}
		// Version sets of each DB must match
		if !versions.Equal(scVersions) {
			err = fmt.Errorf("Storage and StateCommitment DB have different version history") //nolint:stylecheck
			return
		}
		err = opts.StateCommitmentDB.Revert()
		if err != nil {
			return
		}
		stateCommitmentTxn = opts.StateCommitmentDB.ReadWriter()
	}

	ret = &Store{
		stateDB:            db,
		stateTxn:           stateTxn,
		StateCommitmentDB:  opts.StateCommitmentDB,
		stateCommitmentTxn: stateCommitmentTxn,
		mem:                mem.NewStore(),
		tran:               transient.NewStore(),

		substoreCache: map[string]*substore{},

		traceListenMixin: opts.traceListenMixin,
		PersistentCache:  opts.PersistentCache,

		Pruning:        opts.Pruning,
		InitialVersion: opts.InitialVersion,
	}

	// Now load the substore schema
	schemaView := prefixdb.NewPrefixReader(ret.stateDB.Reader(), schemaPrefix)
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, schemaView.Discard(), "schemaView.Discard also failed")
			err = util.CombineErrors(err, ret.Close(), "base.Close also failed")
		}
	}()
	writeSchema := func(sch StoreSchema) {
		schemaWriter := prefixdb.NewPrefixWriter(ret.stateTxn, schemaPrefix)
		var it dbm.Iterator
		it, err = schemaView.Iterator(nil, nil)
		if err != nil {
			return
		}
		for it.Next() {
			err = schemaWriter.Delete(it.Key())
			if err != nil {
				return
			}
		}
		err = it.Close()
		if err != nil {
			return
		}
		err = schemaView.Discard()
		if err != nil {
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
		pfx := substorePrefix(key)
		subReader := prefixdb.NewPrefixReader(reader, pfx)
		it, err := subReader.Iterator(nil, nil)
		if err != nil {
			return err
		}
		for it.Next() {
			store.stateTxn.Delete(it.Key())
		}
		it.Close()
		if store.StateCommitmentDB != nil {
			subReader = prefixdb.NewPrefixReader(scReader, pfx)
			it, err = subReader.Iterator(nil, nil)
			if err != nil {
				return err
			}
			for it.Next() {
				store.stateCommitmentTxn.Delete(it.Key())
			}
			it.Close()
		}
	}
	for _, rename := range upgrades.Renamed {
		oldPrefix := substorePrefix(rename.OldKey)
		newPrefix := substorePrefix(rename.NewKey)
		subReader := prefixdb.NewPrefixReader(reader, oldPrefix)
		subWriter := prefixdb.NewPrefixWriter(store.stateTxn, newPrefix)
		it, err := subReader.Iterator(nil, nil)
		if err != nil {
			return err
		}
		for it.Next() {
			subWriter.Set(it.Key(), it.Value())
		}
		it.Close()
		if store.StateCommitmentDB != nil {
			subReader = prefixdb.NewPrefixReader(scReader, oldPrefix)
			subWriter = prefixdb.NewPrefixWriter(store.stateCommitmentTxn, newPrefix)
			it, err = subReader.Iterator(nil, nil)
			if err != nil {
				return err
			}
			for it.Next() {
				subWriter.Set(it.Key(), it.Value())
			}
			it.Close()
		}
	}
	return nil
}

func substorePrefix(key string) []byte {
	return append(contentPrefix, key...)
}

// GetKVStore implements MultiStore.
func (rs *Store) GetKVStore(skey types.StoreKey) types.KVStore {
	key := skey.Name()
	var parent types.KVStore
	typ, has := rs.schema[skey]
	if !has {
		panic(ErrStoreNotFound(key))
	}
	switch typ {
	case types.StoreTypeMemory:
		parent = rs.mem
	case types.StoreTypeTransient:
		parent = rs.tran
	case types.StoreTypePersistent:
	default:
		panic(fmt.Errorf("StoreType not supported: %v", typ)) // should never happen
	}
	var ret types.KVStore
	if parent != nil { // store is non-persistent
		ret = prefix.NewStore(parent, []byte(key))
	} else { // store is persistent
		sub, err := rs.getSubstore(key)
		if err != nil {
			panic(err)
		}
		rs.substoreCache[key] = sub
		ret = sub
	}
	// Wrap with trace/listen if needed. Note: we don't cache this, so users must get a new substore after
	// modifying tracers/listeners.
	return rs.wrapTraceListen(ret, skey)
}

// Gets a persistent substore. This reads, but does not update the substore cache.
// Use it in cases where we need to access a store internally (e.g. read/write Merkle keys, queries)
func (rs *Store) getSubstore(key string) (*substore, error) {
	if cached, has := rs.substoreCache[key]; has {
		return cached, nil
	}
	pfx := substorePrefix(key)
	stateRW := prefixdb.NewPrefixReadWriter(rs.stateTxn, pfx)
	stateCommitmentRW := prefixdb.NewPrefixReadWriter(rs.stateCommitmentTxn, pfx)
	var stateCommitmentStore *smt.Store

	rootHash, err := stateRW.Get(substoreMerkleRootKey)
	if err != nil {
		return nil, err
	}
	if rootHash != nil {
		stateCommitmentStore = loadSMT(stateCommitmentRW, rootHash)
	} else {
		smtdb := prefixdb.NewPrefixReadWriter(stateCommitmentRW, smtPrefix)
		stateCommitmentStore = smt.NewStore(smtdb)
	}

	return &substore{
		root:                 rs,
		name:                 key,
		dataBucket:           prefixdb.NewPrefixReadWriter(stateRW, dataPrefix),
		indexBucket:          prefixdb.NewPrefixReadWriter(stateRW, indexPrefix),
		stateCommitmentStore: stateCommitmentStore,
	}, nil
}

// Resets a substore's state after commit (because root stateTxn has been discarded)
func (s *substore) refresh(rootHash []byte) {
	pfx := substorePrefix(s.name)
	stateRW := prefixdb.NewPrefixReadWriter(s.root.stateTxn, pfx)
	stateCommitmentRW := prefixdb.NewPrefixReadWriter(s.root.stateCommitmentTxn, pfx)
	s.dataBucket = prefixdb.NewPrefixReadWriter(stateRW, dataPrefix)
	s.indexBucket = prefixdb.NewPrefixReadWriter(stateRW, indexPrefix)
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

	// Prune if necessary
	previous := cid.Version - 1
	if s.Pruning.Interval != 0 && cid.Version%int64(s.Pruning.Interval) == 0 {
		// The range of newly prunable versions
		lastPrunable := previous - int64(s.Pruning.KeepRecent)
		firstPrunable := lastPrunable - int64(s.Pruning.Interval)

		for version := firstPrunable; version <= lastPrunable; version++ {
			s.stateDB.DeleteVersion(uint64(version))

			if s.StateCommitmentDB != nil {
				s.StateCommitmentDB.DeleteVersion(uint64(version))
			}
		}
	}

	s.tran.Commit()
	return *cid
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

// Calculates root hashes and commits to DB. Does not verify target version or perform pruning.
func (s *Store) commit(target uint64) (id *types.CommitID, err error) {
	storeHashes, err := s.getMerkleRoots()
	if err != nil {
		return
	}
	// Update substore Merkle roots
	for key, storeHash := range storeHashes {
		pfx := substorePrefix(key)
		stateW := prefixdb.NewPrefixReadWriter(s.stateTxn, pfx)
		if err = stateW.Set(substoreMerkleRootKey, storeHash); err != nil {
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
	err = s.stateDB.SaveVersion(target)
	if err != nil {
		return
	}

	stateTxn := s.stateDB.ReadWriter()
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, stateTxn.Discard(), "stateTxn.Discard also failed")
		}
	}()
	stateCommitmentTxn := stateTxn

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

		err = s.StateCommitmentDB.SaveVersion(target)
		if err != nil {
			return
		}
		stateCommitmentTxn = s.StateCommitmentDB.ReadWriter()
	}

	s.stateTxn = stateTxn
	s.stateCommitmentTxn = stateCommitmentTxn
	// the state of all live substores must be refreshed
	for key, sub := range s.substoreCache {
		sub.refresh(storeHashes[key])
	}

	return &types.CommitID{Version: int64(target), Hash: rootHash}, nil
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
func (rs *Store) SetInitialVersion(version uint64) error {
	rs.InitialVersion = uint64(version)
	return nil
}

// GetVersion implements CommitMultiStore.
func (rs *Store) GetVersion(version int64) (types.MultiStore, error) {
	return rs.getView(version)
}

// CacheWrap implements MultiStore.
func (rs *Store) CacheWrap() types.CacheMultiStore {
	return newCacheStore(rs)
}

// GetAllVersions implements CommitMultiStore.
// https://github.com/cosmos/cosmos-sdk/pull/11124
func (rs *Store) GetAllVersions() []int {
	vs, err := rs.stateDB.Versions()
	if err != nil {
		panic(err)
	}
	var ret []int
	for it := vs.Iterator(); it.Next(); {
		ret = append(ret, int(it.Value()))
	}
	return ret
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
func (rs *Store) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
	if len(req.Data) == 0 {
		return sdkerrors.QueryResult(sdkerrors.Wrap(sdkerrors.ErrTxDecode, "query cannot be zero length"), false)
	}

	// if height is 0, use the latest height
	height := req.Height
	if height == 0 {
		versions, err := rs.stateDB.Versions()
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
	view, err := rs.getView(height)
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
			return sdkerrors.QueryResult(fmt.Errorf("Merkle proof creation failed for key: %v", res.Key), false) //nolint: stylecheck // proper name
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

func loadSMT(stateCommitmentTxn dbm.DBReadWriter, root []byte) *smt.Store {
	smtdb := prefixdb.NewPrefixReadWriter(stateCommitmentTxn, smtPrefix)
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

func (pr *SchemaBuilder) storeInfo(key string) (sst types.StoreType, ix int, err error) {
	ix, has := binarySearch(pr.reserved, key)
	if !has {
		err = fmt.Errorf("prefix does not exist: %v", key)
		return
	}
	sst, has = pr.StoreSchema[key]
	if !has {
		err = fmt.Errorf("prefix is registered but not in schema: %v", key)
	}

	return
}

// registerName registers a store key by name only
func (pr *SchemaBuilder) registerName(key string, typ types.StoreType) error {
	if !validSubStoreType(typ) {
		return fmt.Errorf("StoreType not supported: %v", typ)
	}

	// Find the neighboring reserved prefix, and check for duplicates and conflicts
	i, has := binarySearch(pr.reserved, key)
	if has {
		return fmt.Errorf("prefix already exists: %v", key)
	}
	if i > 0 && strings.HasPrefix(key, pr.reserved[i-1]) {
		return fmt.Errorf("prefix conflict: '%v' exists, cannot add '%v'", pr.reserved[i-1], key)
	}
	if i < len(pr.reserved) && strings.HasPrefix(pr.reserved[i], key) {
		return fmt.Errorf("prefix conflict: '%v' exists, cannot add '%v'", pr.reserved[i], key)
	}
	reserved := pr.reserved[:i]
	reserved = append(reserved, key)
	pr.reserved = append(reserved, pr.reserved[i:]...)
	pr.StoreSchema[key] = typ
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

func (tlm *traceListenMixin) wrapTraceListen(store types.KVStore, skey types.StoreKey) types.KVStore {
	if tlm.TracingEnabled() {
		store = tracekv.NewStore(store, tlm.TraceWriter, tlm.getTracingContext())
	}
	if tlm.ListeningEnabled(skey) {
		store = listenkv.NewStore(store, skey, tlm.listeners[skey])
	}
	return store
}

func (s *Store) GetPruning() types.PruningOptions   { return s.Pruning }
func (s *Store) SetPruning(po types.PruningOptions) { s.Pruning = po }

func (rs *Store) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshottypes.SnapshotItem, error) {
	return snapshottypes.SnapshotItem{}, nil
}
func (rs *Store) Snapshot(height uint64, protoWriter protoio.Writer) error {
	return nil
}
