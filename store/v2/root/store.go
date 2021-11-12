// RootStore supports a subset of the StoreType values: Persistent, Memory, and Transient

package root

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"

	abci "github.com/tendermint/tendermint/abci/types"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	prefixdb "github.com/cosmos/cosmos-sdk/db/prefix"
	util "github.com/cosmos/cosmos-sdk/internal"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	types "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/mem"
	"github.com/cosmos/cosmos-sdk/store/v2/smt"
	transkv "github.com/cosmos/cosmos-sdk/store/v2/transient"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

var (
	_ types.KVStore         = (*Store)(nil)
	_ types.Queryable       = (*Store)(nil)
	_ types.CommitRootStore = (*Store)(nil)
	_ types.CacheRootStore  = (*cacheStore)(nil)
	_ types.BasicRootStore  = (*viewStore)(nil)
)

var (
	merkleRootKey     = []byte{0} // Key for root hash of Merkle tree
	dataPrefix        = []byte{1} // Prefix for state mappings
	indexPrefix       = []byte{2} // Prefix for Store reverse index
	merkleNodePrefix  = []byte{3} // Prefix for Merkle tree nodes
	merkleValuePrefix = []byte{4} // Prefix for Merkle value mappings
	schemaPrefix      = []byte{5} // Prefix for store keys (namespaces)
)

var (
	ErrVersionDoesNotExist = errors.New("version does not exist")
	ErrMaximumHeight       = errors.New("maximum block height reached")
)

// StoreConfig is used to define a schema and pass options to the RootStore constructor.
type StoreConfig struct {
	// Version pruning options for backing DBs.
	Pruning        types.PruningOptions
	InitialVersion uint64
	// The backing DB to use for the state commitment Merkle tree data.
	// If nil, Merkle data is stored in the state storage DB under a separate prefix.
	StateCommitmentDB dbm.DBConnection

	prefixRegistry
	PersistentCache types.RootStorePersistentCache
	Upgrades        []types.StoreUpgrades

	*listenerMixin
	*traceMixin
}

// A loaded mapping of substore keys to store types
type StoreSchema map[string]types.StoreType

// Builder type used to create a valid schema with no prefix conflicts
type prefixRegistry struct {
	StoreSchema
	reserved []string
}

// Mixin types that will be composed into each distinct root store variant type
type listenerMixin struct {
	listeners map[types.StoreKey][]types.WriteListener
}

type traceMixin struct {
	TraceWriter  io.Writer
	TraceContext types.TraceContext
}

// Main persistent store type
type Store struct {
	stateDB            dbm.DBConnection
	stateTxn           dbm.DBReadWriter
	dataBucket         dbm.DBReadWriter
	indexBucket        dbm.DBReadWriter
	stateCommitmentTxn dbm.DBReadWriter
	// State commitment (SC) KV store for current version
	stateCommitmentStore *smt.Store

	Pruning           types.PruningOptions
	InitialVersion    uint64
	StateCommitmentDB dbm.DBConnection

	schema StoreSchema
	mem    *mem.Store
	tran   *transkv.Store
	*listenerMixin
	*traceMixin

	mtx sync.RWMutex

	PersistentCache types.RootStorePersistentCache
}

// Branched state
type cacheStore struct {
	types.CacheKVStore
	mem, tran types.CacheKVStore
	schema    StoreSchema
	*listenerMixin
	*traceMixin
}

// Read-only store for querying
type viewStore struct {
	stateView            dbm.DBReader
	dataBucket           dbm.DBReader
	indexBucket          dbm.DBReader
	stateCommitmentView  dbm.DBReader
	stateCommitmentStore *smt.Store

	schema StoreSchema
}

// Auxiliary type used only to avoid repetitive method implementations
type rootGeneric struct {
	schema             StoreSchema
	persist, mem, tran types.KVStore
}

// DefaultStoreConfig returns a RootStore config with an empty schema, a single backing DB,
// pruning with PruneDefault, no listeners and no tracer.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		Pruning: types.PruneDefault,
		prefixRegistry: prefixRegistry{
			StoreSchema: StoreSchema{},
		},
		listenerMixin: &listenerMixin{
			listeners: map[types.StoreKey][]types.WriteListener{},
		},
		traceMixin: &traceMixin{
			TraceWriter:  nil,
			TraceContext: nil,
		},
	}
}

// Returns true for valid store types for a RootStore schema
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
func readSavedSchema(bucket dbm.DBReader) (*prefixRegistry, error) {
	ret := prefixRegistry{StoreSchema: StoreSchema{}}
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
	it.Close()
	return &ret, nil
}

// NewStore constructs a RootStore directly from a DB connection and options.
func NewStore(db dbm.DBConnection, opts StoreConfig) (ret *Store, err error) {
	versions, err := db.Versions()
	if err != nil {
		return
	}
	loadExisting := false
	// If the DB is not empty, attempt to load existing data
	if saved := versions.Count(); saved != 0 {
		if opts.InitialVersion != 0 && versions.Last() < opts.InitialVersion {
			return nil, fmt.Errorf("latest saved version is less than initial version: %v < %v",
				versions.Last(), opts.InitialVersion)
		}
		loadExisting = true
	}
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
		var mversions dbm.VersionSet
		mversions, err = opts.StateCommitmentDB.Versions()
		if err != nil {
			return
		}
		// Version sets of each DB must match
		if !versions.Equal(mversions) {
			err = fmt.Errorf("Storage and StateCommitment DB have different version history") //nolint:stylecheck
			return
		}
		err = opts.StateCommitmentDB.Revert()
		if err != nil {
			return
		}
		stateCommitmentTxn = opts.StateCommitmentDB.ReadWriter()
	}

	var stateCommitmentStore *smt.Store
	if loadExisting {
		var root []byte
		root, err = stateTxn.Get(merkleRootKey)
		if err != nil {
			return
		}
		if root == nil {
			err = fmt.Errorf("could not get root of SMT")
			return
		}
		stateCommitmentStore = loadSMT(stateCommitmentTxn, root)
	} else {
		merkleNodes := prefixdb.NewPrefixReadWriter(stateCommitmentTxn, merkleNodePrefix)
		merkleValues := prefixdb.NewPrefixReadWriter(stateCommitmentTxn, merkleValuePrefix)
		stateCommitmentStore = smt.NewStore(merkleNodes, merkleValues)
	}
	ret = &Store{
		stateDB:              db,
		stateTxn:             stateTxn,
		dataBucket:           prefixdb.NewPrefixReadWriter(stateTxn, dataPrefix),
		indexBucket:          prefixdb.NewPrefixReadWriter(stateTxn, indexPrefix),
		stateCommitmentTxn:   stateCommitmentTxn,
		stateCommitmentStore: stateCommitmentStore,

		Pruning:           opts.Pruning,
		InitialVersion:    opts.InitialVersion,
		StateCommitmentDB: opts.StateCommitmentDB,
		PersistentCache:   opts.PersistentCache,
		listenerMixin:     opts.listenerMixin,
		traceMixin:        opts.traceMixin,
	}

	// Now load the substore schema
	schemaView := prefixdb.NewPrefixReader(ret.stateDB.Reader(), schemaPrefix)
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, schemaView.Discard(), "schemaView.Discard also failed")
			err = util.CombineErrors(err, ret.Close(), "base.Close also failed")
		}
	}()
	reg, err := readSavedSchema(schemaView)
	if err != nil {
		return
	}
	// If the loaded schema is empty, just copy the config schema;
	// Otherwise, verify it is identical to the config schema
	if len(reg.StoreSchema) == 0 {
		for k, v := range opts.StoreSchema {
			reg.StoreSchema[k] = v
		}
		reg.reserved = make([]string, len(opts.reserved))
		copy(reg.reserved, opts.reserved)
	} else {
		if !reg.equal(opts.StoreSchema) {
			err = errors.New("loaded schema does not match configured schema")
			return
		}
	}
	// Apply migrations, then clear old schema and write the new one
	for _, upgrades := range opts.Upgrades {
		err = reg.migrate(ret, upgrades)
		if err != nil {
			return
		}
	}
	schemaWriter := prefixdb.NewPrefixWriter(ret.stateTxn, schemaPrefix)
	it, err := schemaView.Iterator(nil, nil)
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
	for skey, typ := range reg.StoreSchema {
		err = schemaWriter.Set([]byte(skey), []byte{byte(typ)})
		if err != nil {
			return
		}
	}
	// The migrated contents and schema are not committed until the next store.Commit
	ret.mem = mem.NewStore(memdb.NewDB())
	ret.tran = transkv.NewStore(memdb.NewDB())
	ret.schema = reg.StoreSchema
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
func (pr *prefixRegistry) migrate(store *Store, upgrades types.StoreUpgrades) error {
	// branch state to allow mutation while iterating
	branch := cachekv.NewStore(store)

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

		sub := prefix.NewStore(store, []byte(key))
		subbranch := prefix.NewStore(branch, []byte(key))
		it := sub.Iterator(nil, nil)
		for ; it.Valid(); it.Next() {
			subbranch.Delete(it.Key())
		}
		it.Close()
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
		err = pr.ReservePrefix(rename.NewKey, types.StoreTypePersistent)
		if err != nil {
			return err
		}

		sub := prefix.NewStore(store, []byte(rename.OldKey))
		subbranch := prefix.NewStore(branch, []byte(rename.NewKey))
		it := sub.Iterator(nil, nil)
		for ; it.Valid(); it.Next() {
			subbranch.Set(it.Key(), it.Value())
		}
		it.Close()
	}
	branch.Write()

	for _, key := range upgrades.Added {
		err := pr.ReservePrefix(key, types.StoreTypePersistent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rs *Store) GetKVStore(key types.StoreKey) types.KVStore {
	return rs.generic().getStore(key.Name())
}

// Commit implements Committer.
func (s *Store) Commit() types.CommitID {
	versions, err := s.stateDB.Versions()
	if err != nil {
		panic(err)
	}
	target := versions.Last() + 1
	if target > math.MaxInt64 {
		panic(ErrMaximumHeight)
	}
	// Fast forward to initialversion if needed
	if s.InitialVersion != 0 && target < s.InitialVersion {
		target = s.InitialVersion
	}
	cid, err := s.commit(target)
	if err != nil {
		panic(err)
	}

	previous := cid.Version - 1
	if s.Pruning.KeepEvery != 1 && s.Pruning.Interval != 0 && cid.Version%int64(s.Pruning.Interval) == 0 {
		// The range of newly prunable versions
		lastPrunable := previous - int64(s.Pruning.KeepRecent)
		firstPrunable := lastPrunable - int64(s.Pruning.Interval)
		for version := firstPrunable; version <= lastPrunable; version++ {
			if s.Pruning.KeepEvery == 0 || version%int64(s.Pruning.KeepEvery) != 0 {
				s.stateDB.DeleteVersion(uint64(version))
				if s.StateCommitmentDB != nil {
					s.StateCommitmentDB.DeleteVersion(uint64(version))
				}
			}
		}
	}

	s.tran.Commit()

	return *cid
}

func (s *Store) commit(target uint64) (id *types.CommitID, err error) {
	root := s.stateCommitmentStore.Root()
	err = s.stateTxn.Set(merkleRootKey, root)
	if err != nil {
		return
	}
	err = s.stateTxn.Commit()
	if err != nil {
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

	// If DBs are not separate, StateCommitment state has been commmitted & snapshotted
	if s.StateCommitmentDB != nil {
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
	s.dataBucket = prefixdb.NewPrefixReadWriter(stateTxn, dataPrefix)
	s.indexBucket = prefixdb.NewPrefixReadWriter(stateTxn, indexPrefix)
	s.stateCommitmentTxn = stateCommitmentTxn
	s.stateCommitmentStore = loadSMT(stateCommitmentTxn, root)

	return &types.CommitID{Version: int64(target), Hash: root}, nil
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

func (rs *Store) SetInitialVersion(version uint64) error {
	rs.InitialVersion = uint64(version)
	return nil
}

func (rs *Store) GetVersion(version int64) (types.BasicRootStore, error) {
	return rs.getView(version)
}

func (rs *Store) CacheRootStore() types.CacheRootStore {
	return &cacheStore{
		CacheKVStore:  cachekv.NewStore(rs),
		mem:           cachekv.NewStore(rs.mem),
		tran:          cachekv.NewStore(rs.tran),
		schema:        rs.schema,
		listenerMixin: &listenerMixin{},
		traceMixin:    &traceMixin{},
	}
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
		return sdkerrors.QueryResult(err, false)
	}
	view, err := rs.getView(height)
	if err != nil {
		if errors.Is(err, dbm.ErrVersionDoesNotExist) {
			err = sdkerrors.ErrInvalidHeight
		}
		return sdkerrors.QueryResult(err, false)
	}

	substore := view.generic().getStore(storeName)
	if substore == nil {
		return sdkerrors.QueryResult(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "no such store: %s", storeName), false)
	}

	switch subpath {
	case "/key":
		var err error
		res.Key = req.Data // data holds the key bytes
		res.Value = substore.Get(res.Key)
		if !req.Prove {
			break
		}
		// res.ProofOps, err = view.prove(storeName, res.Key)
		res.ProofOps, err = view.stateCommitmentStore.GetProof([]byte(storeName + string(res.Key)))
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

func (rs *Store) generic() rootGeneric { return rootGeneric{rs.schema, rs, rs.mem, rs.tran} }

func (store rootGeneric) getStore(key string) types.KVStore {
	var sub types.KVStore
	if typ, ok := store.schema[key]; ok {
		switch typ {
		case types.StoreTypePersistent:
			sub = store.persist
		case types.StoreTypeMemory:
			sub = store.mem
		case types.StoreTypeTransient:
			sub = store.tran
		}
	}
	if sub == nil {
		panic(fmt.Errorf("store does not exist for key: %s", key))
	}
	return prefix.NewStore(sub, []byte(key))
}

func (rs *cacheStore) GetKVStore(key types.StoreKey) types.KVStore {
	return rs.generic().getStore(key.Name())
}

func (rs *cacheStore) Write() {
	rs.CacheKVStore.Write()
	rs.mem.Write()
	rs.tran.Write()
}

// Recursively wraps the CacheRootStore in another cache store.
func (rs *cacheStore) CacheRootStore() types.CacheRootStore {
	return &cacheStore{
		CacheKVStore:  cachekv.NewStore(rs),
		mem:           cachekv.NewStore(rs.mem),
		tran:          cachekv.NewStore(rs.tran),
		schema:        rs.schema,
		listenerMixin: &listenerMixin{},
		traceMixin:    &traceMixin{},
	}
}

func (rs *cacheStore) generic() rootGeneric { return rootGeneric{rs.schema, rs, rs.mem, rs.tran} }

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

func (pr *prefixRegistry) storeInfo(key string) (sst types.StoreType, ix int, err error) {
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

func (pr *prefixRegistry) ReservePrefix(key string, typ types.StoreType) error {
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

func (lreg *listenerMixin) AddListeners(key types.StoreKey, listeners []types.WriteListener) {
	if ls, ok := lreg.listeners[key]; ok {
		lreg.listeners[key] = append(ls, listeners...)
	} else {
		lreg.listeners[key] = listeners
	}
}

// ListeningEnabled returns if listening is enabled for a specific KVStore
func (lreg *listenerMixin) ListeningEnabled(key types.StoreKey) bool {
	if ls, ok := lreg.listeners[key]; ok {
		return len(ls) != 0
	}
	return false
}

func (treg *traceMixin) TracingEnabled() bool {
	return treg.TraceWriter != nil
}
func (treg *traceMixin) SetTracer(w io.Writer) {
	treg.TraceWriter = w
}
func (treg *traceMixin) SetTraceContext(tc types.TraceContext) {
	treg.TraceContext = tc
}

func (rs *Store) Restore(height uint64, format uint32, chunks <-chan io.ReadCloser, ready chan<- struct{}) error {
	return nil
}
func (rs *Store) Snapshot(height uint64, format uint32) (<-chan io.ReadCloser, error) {
	return nil, nil
}
