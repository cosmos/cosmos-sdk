package flat

import (
	"errors"
	"fmt"
	"io"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	prefixdb "github.com/cosmos/cosmos-sdk/db/prefix"
	util "github.com/cosmos/cosmos-sdk/internal"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	types "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/mem"
	transkv "github.com/cosmos/cosmos-sdk/store/v2/transient"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

var (
	_ types.CommitRootStore = (*rootStore)(nil)
	_ types.CacheRootStore  = (*rootCache)(nil)
)

var (
	schemaPrefix = []byte{5} // Prefix for store keys (prefixes)
)

// RootStoreConfig is used to define a schema and pass options to the RootStore constructor.
type RootStoreConfig struct {
	StoreConfig
	PersistentCache types.RootStorePersistentCache
	Upgrades        []types.StoreUpgrades
	prefixRegistry
	*listenerMixin
	*traceMixin
}

// Represents the valid store types for a RootStore schema, a subset of the StoreType values
type subStoreType byte

const (
	subStorePersistent subStoreType = iota
	subStoreMemory
	subStoreTransient
)

// A loaded mapping of store names to types
type storeSchema map[string]subStoreType

// Builder type used to create a valid schema with no prefix conflicts
type prefixRegistry struct {
	storeSchema
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

type storeMixin struct {
	schema storeSchema
	*listenerMixin
	*traceMixin
}

// Main persistent store type
type rootStore struct {
	*Store
	mem  *mem.Store
	tran *transkv.Store
	storeMixin
}

// Branched state
type rootCache struct {
	types.CacheKVStore
	mem, tran types.CacheKVStore
	storeMixin
}

// Read-only store for querying
type rootView struct {
	*storeView
	schema storeSchema
	// storeMixin //?
}

// Auxiliary type used only to avoid repetitive method implementations
type rootGeneric struct {
	schema             storeSchema
	persist, mem, tran types.KVStore
}

// DefaultRootStoreConfig returns a RootStore config with an empty schema, a single backing DB,
// pruning with PruneDefault, no listeners and no tracer.
func DefaultRootStoreConfig() RootStoreConfig {
	return RootStoreConfig{
		StoreConfig: StoreConfig{Pruning: types.PruneDefault},
		prefixRegistry: prefixRegistry{
			storeSchema: storeSchema{},
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

func validSubStoreType(sst subStoreType) bool {
	return byte(sst) <= byte(subStoreTransient)
}

// Returns true iff both schema maps match exactly (including mem/tran stores)
func (this storeSchema) equal(that storeSchema) bool {
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
func readSchema(bucket dbm.DBReader) (*prefixRegistry, error) {
	ret := prefixRegistry{storeSchema: storeSchema{}}
	it, err := bucket.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}
	for it.Next() {
		value := it.Value()
		if len(value) != 1 || !validSubStoreType(subStoreType(value[0])) {
			return nil, fmt.Errorf("invalid mapping for store key: %v => %v", it.Key(), value)
		}
		ret.storeSchema[string(it.Key())] = subStoreType(value[0])
		ret.reserved = append(ret.reserved, string(it.Key())) // assume iter yields keys sorted
	}
	it.Close()
	return &ret, nil
}

// NewRootStore constructs a RootStore directly from a DB connection and options.
func NewRootStore(db dbm.DBConnection, opts RootStoreConfig) (*rootStore, error) {
	base, err := NewStore(db, opts.StoreConfig)
	if err != nil {
		return nil, err
	}
	return makeRootStore(base, opts)
}

// TODO:
// should config contain the pre- or post-migration schema? - currently pre
func makeRootStore(base *Store, opts RootStoreConfig) (ret *rootStore, err error) {
	schemaView := prefixdb.NewPrefixReader(base.stateDB.Reader(), schemaPrefix)
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, schemaView.Discard(), "schemaView.Discard also failed")
			err = util.CombineErrors(err, base.Close(), "base.Close also failed")
		}
	}()
	pr, err := readSchema(schemaView)
	if err != nil {
		return
	}
	// If the loaded schema is empty, just copy the config schema;
	// Otherwise, verify it is identical to the config schema
	if len(pr.storeSchema) == 0 {
		for k, v := range opts.storeSchema {
			pr.storeSchema[k] = v
		}
		pr.reserved = make([]string, len(opts.reserved))
		copy(pr.reserved, opts.reserved)
	} else {
		if !pr.equal(opts.storeSchema) {
			err = errors.New("loaded schema does not match configured schema")
			return
		}
	}
	// Apply migrations, then clear old schema and write the new one
	for _, upgrades := range opts.Upgrades {
		err = pr.migrate(base, upgrades)
		if err != nil {
			return
		}
	}
	schemaWriter := prefixdb.NewPrefixWriter(base.stateTxn, schemaPrefix)
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
	for skey, typ := range pr.storeSchema {
		err = schemaWriter.Set([]byte(skey), []byte{byte(typ)})
		if err != nil {
			return
		}
	}
	// The migrated contents and schema are not committed until the next store.Commit
	ret = &rootStore{
		Store: base,
		mem:   mem.NewStore(memdb.NewDB()),
		tran:  transkv.NewStore(memdb.NewDB()),
		storeMixin: storeMixin{
			schema:        pr.storeSchema,
			listenerMixin: opts.listenerMixin,
			traceMixin:    opts.traceMixin,
		},
	}
	return
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
		if sst != subStorePersistent {
			return fmt.Errorf("prefix is for non-persistent substore: %v (%v)", key, sst)
		}
		pr.reserved = append(pr.reserved[:ix], pr.reserved[ix+1:]...)
		delete(pr.storeSchema, key)

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
		if sst != subStorePersistent {
			return fmt.Errorf("prefix is for non-persistent substore: %v (%v)", rename.OldKey, sst)
		}
		pr.reserved = append(pr.reserved[:ix], pr.reserved[ix+1:]...)
		delete(pr.storeSchema, rename.OldKey)
		err = pr.ReservePrefix(types.NewKVStoreKey(rename.NewKey), types.StoreTypePersistent)
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
		err := pr.ReservePrefix(types.NewKVStoreKey(key), types.StoreTypePersistent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rs *rootStore) GetKVStore(key types.StoreKey) types.KVStore {
	return rs.generic().getStore(key.Name())
}

func (rs *rootStore) Commit() types.CommitID {
	id := rs.Store.Commit()
	rs.mem.Commit()
	rs.tran.Commit()
	return id
}

func (rs *rootStore) Close() error { return rs.Store.Close() }

func (rs *Store) SetInitialVersion(version uint64) error {
	rs.opts.InitialVersion = uint64(version)
	return nil
}

func (rs *rootStore) GetVersion(version int64) (types.BasicRootStore, error) {
	return rs.getView(version)
}

func (rs *rootStore) getView(version int64) (*rootView, error) {
	view, err := rs.Store.GetVersion(version)
	if err != nil {
		return nil, err
	}
	return rs.makeRootView(view)
}

func (rs *rootStore) makeRootView(view *storeView) (ret *rootView, err error) {
	schemaView := prefixdb.NewPrefixReader(view.stateView, schemaPrefix)
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, schemaView.Discard(), "schemaView.Discard also failed")
		}
	}()
	pr, err := readSchema(schemaView)
	if err != nil {
		return
	}
	// The migrated contents and schema are not committed until the next store.Commit
	return &rootView{
		storeView: view,
		schema:    pr.storeSchema,
	}, nil
}

// if the schema indicates a mem/tran store, it's ignored
func (rv *rootView) generic() rootGeneric { return rootGeneric{rv.schema, rv, nil, nil} }

func (rv *rootView) GetKVStore(key types.StoreKey) types.KVStore {
	return rv.generic().getStore(key.Name())
}

// Copies only the schema
func newStoreMixin(schema storeSchema) storeMixin {
	return storeMixin{
		schema:        schema,
		listenerMixin: &listenerMixin{},
		traceMixin:    &traceMixin{},
	}
}

func (rv *rootView) CacheRootStore() types.CacheRootStore {
	return &rootCache{
		CacheKVStore: cachekv.NewStore(rv),
		mem:          cachekv.NewStore(mem.NewStore(memdb.NewDB())),
		tran:         cachekv.NewStore(transkv.NewStore(memdb.NewDB())),
		storeMixin:   newStoreMixin(rv.schema),
	}
}

func (rs *rootStore) CacheRootStore() types.CacheRootStore {
	return &rootCache{
		CacheKVStore: cachekv.NewStore(rs),
		mem:          cachekv.NewStore(rs.mem),
		tran:         cachekv.NewStore(rs.tran),
		storeMixin:   newStoreMixin(rs.schema),
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
func (rs *rootStore) Query(req abci.RequestQuery) (res abci.ResponseQuery) {
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

	// trim the path and make the query
	// req.Path = subpath
	// res := rs.Store.Query(req)

	switch subpath {
	case "/key":
		var err error
		res.Key = req.Data // data holds the key bytes

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
		res.Value = substore.Get(res.Key)
		if !req.Prove {
			break
		}
		// res.ProofOps, err = view.prove(storeName, res.Key)
		fullkey := storeName + string(res.Key)
		res.ProofOps, err = view.merkleStore.GetProof([]byte(fullkey))
		if err != nil {
			return sdkerrors.QueryResult(fmt.Errorf("Merkle proof creation failed for key: %v", res.Key), false)
		}

	case "/subspace":
		pairs := kv.Pairs{
			Pairs: make([]kv.Pair, 0),
		}

		subspace := req.Data
		res.Key = subspace

		iterator := rs.Iterator(subspace, types.PrefixEndBytes(subspace))
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

func (rs *rootStore) generic() rootGeneric { return rootGeneric{rs.schema, rs, rs.mem, rs.tran} }

func (store rootGeneric) getStore(key string) types.KVStore {
	var sub types.KVStore
	if typ, ok := store.schema[key]; ok {
		switch typ {
		case subStorePersistent:
			sub = store.persist
		case subStoreMemory:
			sub = store.mem
		case subStoreTransient:
			sub = store.tran
		}
	}
	if sub == nil {
		panic(fmt.Sprintf("store does not exist for key: %s", key))
	}
	return prefix.NewStore(sub, []byte(key))
}

func (rs *rootCache) GetKVStore(key types.StoreKey) types.KVStore {
	return rs.generic().getStore(key.Name())
}

func (rs *rootCache) Write() {
	rs.CacheKVStore.Write()
	rs.mem.Write()
	rs.tran.Write()
}

// Recursively wraps the CacheRootStore in another cache store.
func (rs *rootCache) CacheRootStore() types.CacheRootStore {
	return &rootCache{
		CacheKVStore: cachekv.NewStore(rs),
		mem:          cachekv.NewStore(rs.mem),
		tran:         cachekv.NewStore(rs.tran),
		storeMixin:   newStoreMixin(rs.schema),
	}
}

func (rs *rootCache) generic() rootGeneric { return rootGeneric{rs.schema, rs, rs.mem, rs.tran} }

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

func (pr *prefixRegistry) storeInfo(key string) (sst subStoreType, ix int, err error) {
	ix, has := binarySearch(pr.reserved, key)
	if !has {
		err = fmt.Errorf("prefix does not exist: %v", key)
		return
	}
	sst, has = pr.storeSchema[key]
	if !has {
		err = fmt.Errorf("prefix is registered but not in schema: %v", key)
	}

	return
}

func (pr *prefixRegistry) ReservePrefix(key types.StoreKey, typ types.StoreType) error {
	// Find the neighboring reserved prefix, and check for duplicates and conflicts
	i, has := binarySearch(pr.reserved, key.Name())
	if has {
		return fmt.Errorf("prefix already exists: %v", key)
	}
	if i > 0 && strings.HasPrefix(key.Name(), pr.reserved[i-1]) {
		return fmt.Errorf("prefix conflict: '%v' exists, cannot add '%v'", pr.reserved[i-1], key.Name())
	}
	if i < len(pr.reserved) && strings.HasPrefix(pr.reserved[i], key.Name()) {
		return fmt.Errorf("prefix conflict: '%v' exists, cannot add '%v'", pr.reserved[i], key.Name())
	}
	reserved := pr.reserved[:i]
	reserved = append(reserved, key.Name())
	pr.reserved = append(reserved, pr.reserved[i:]...)

	var sstype subStoreType
	switch typ {
	case types.StoreTypeDecoupled:
		sstype = subStorePersistent
	case types.StoreTypeMemory:
		sstype = subStoreMemory
	case types.StoreTypeTransient:
		sstype = subStoreTransient
	// case types.StoreTypeSMT: // could be used for external storage
	default:
		return fmt.Errorf("StoreType not supported: %v", typ)
	}
	pr.storeSchema[key.Name()] = sstype
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

func (rs *rootStore) Restore(height uint64, format uint32, chunks <-chan io.ReadCloser, ready chan<- struct{}) error {
	return nil
}
func (rs *rootStore) Snapshot(height uint64, format uint32) (<-chan io.ReadCloser, error) {
	return nil, nil
}
