package iavl2

import (
	"errors"
	"fmt"
	io "io"
	"path/filepath"

	"cosmossdk.io/log"
	iavlv2 "github.com/cosmos/iavl/v2"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/metrics"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/tracekv"
	"cosmossdk.io/store/types"
)

var (
	_ types.KVStore                    = (*Store)(nil)
	_ types.CommitStore                = (*Store)(nil)
	_ types.CommitKVStore              = (*Store)(nil)
	_ types.CommitKVStoreWithImmutable = (*Store)(nil)
	_ types.Queryable                  = (*Store)(nil)
	_ types.StoreWithInitialVersion    = (*Store)(nil)
	_ io.Closer                        = (*Store)(nil)
)

// Store Implements types.KVStore and CommitKVStore.
type Store struct {
	path     string
	storeKey types.StoreKey
	tree     *iavlv2.Tree
	sqliteDb *iavlv2.SqliteDb
	nodePool *iavlv2.NodePool
	logger   log.Logger
	metrics  metrics.StoreMetrics
}

// LoadStore loads an IAVL v2 store from the given options.
func LoadStore(config Config, opts Options) (types.CommitKVStore, error) {
	// TODO: initial version handling
	logger := opts.Logger

	// we form the file path from the root config path and the key name
	path := filepath.Join(config.Path, opts.Key.Name())

	sqliteOpts := iavlv2.SqliteDbOptions{
		Path: path,
		//Logger:     logger,
		ShardTrees: true,
	}

	// create node pool for memory management
	nodePool := iavlv2.NewNodePool()

	sqliteDb, err := iavlv2.NewSqliteDb(nodePool, sqliteOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	treeOpts := iavlv2.DefaultTreeOptions()
	// TODO figure out what the right default checkpoint interval should be, it's set to 1 for historical queries
	treeOpts.CheckpointInterval = 1
	treeOpts.EvictionDepth = 127

	tree := iavlv2.NewTree(sqliteDb, nodePool, treeOpts)

	id := opts.CommitID
	if id.Version > 0 {
		err = tree.LoadVersion(id.Version)
		if err != nil {
			errClose := sqliteDb.Close()
			return nil, errors.Join(
				fmt.Errorf("failed to load version %d: %w", id.Version, err),
				errClose,
			)
		}
	}

	if logger != nil {
		logger.Debug("Finished loading IAVL v2 tree")
	}

	return &Store{
		path:     path,
		storeKey: opts.Key,
		tree:     tree,
		sqliteDb: sqliteDb,
		nodePool: nodePool,
		logger:   logger,
		metrics:  opts.Metrics,
	}, nil
}

//// UnsafeNewStore returns a reference to a new IAVL Store with a given mutable
//// IAVL tree reference. It should only be used for testing purposes.
////
//// CONTRACT: The IAVL tree should be fully loaded.
//// CONTRACT: PruningOptions passed in as argument must be the same as pruning options
//// passed into iavl.MutableTree
//func UnsafeNewStore(tree *iavl.MutableTree) *Store {
//	return &Store{
//		tree:    tree,
//		metrics: metrics.NewNoOpMetrics(),
//	}
//}

// GetImmutable returns a reference to a new store backed by an immutable IAVL
// tree at a specific version (height) without any pruning options. This should
// be used for querying and iteration only. If the version does not exist or has
// been pruned, an empty immutable IAVL tree will be used.
// Any mutable operations executed will result in a panic.
func (st *Store) GetImmutable(version int64) (types.KVStore, error) {
	treeVersion := st.tree.Version()
	switch treeVersion {
	case version:
	case version + 1:
	default:
		return nil, fmt.Errorf("unsupported query version %d, only current (%d) and previous (%d) versions are supported", version, treeVersion, treeVersion-1)
	}
	return &immutableTreeWrapper{
		version: version,
		store:   st,
	}, nil
}

type immutableTreeWrapper struct {
	version int64
	store   *Store
}

func (st *immutableTreeWrapper) GetStoreType() types.StoreType { return types.StoreTypeIAVL }

func (st *immutableTreeWrapper) CacheWrap() types.CacheWrap { return cachekv.NewStore(st) }

func (st *immutableTreeWrapper) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(st, w, tc))
}

func (st *immutableTreeWrapper) Get(key []byte) []byte {
	ok, value, err := st.store.tree.GetRecent(st.version, key)
	if !ok {
		panic(fmt.Sprintf("version %d does not exist in IAVL tree", st.version))
	}
	if err != nil {
		panic(err)
	}
	return value
}

func (st *immutableTreeWrapper) Has(key []byte) bool {
	return len(st.Get(key)) > 0
}

func (st *immutableTreeWrapper) Iterator(start, end []byte) types.Iterator {
	ok, it := st.store.tree.IterateRecent(st.version, start, end, true)
	if !ok {
		panic(fmt.Sprintf("version %d does not exist in IAVL tree", st.version))
	}
	return it
}

func (st *immutableTreeWrapper) ReverseIterator(start, end []byte) types.Iterator {
	ok, it := st.store.tree.IterateRecent(st.version, start, end, false)
	if !ok {
		panic(fmt.Sprintf("version %d does not exist in IAVL tree", st.version))
	}
	return it
}

func (st *immutableTreeWrapper) Set(key, value []byte) {
	panic("tree is immutable, cannot set!")
}

func (st *immutableTreeWrapper) Delete(key []byte) {
	panic("tree is immutable, cannot delete!")
}

var _ types.KVStore = &immutableTreeWrapper{}

// Commit commits the current store state and returns a CommitID with the new
// version and hash.
func (st *Store) Commit() types.CommitID {
	defer st.metrics.MeasureSince("store", "iavl", "commit")

	hash, version, err := st.tree.SaveVersion()
	if err != nil {
		panic(err)
	}

	return types.CommitID{
		Version: version,
		Hash:    hash,
	}
}

// WorkingHash returns the hash of the current working tree.
func (st *Store) WorkingHash() []byte {
	// TODO is this correct?
	return st.tree.Hash()
}

// LastCommitID implements Committer.
func (st *Store) LastCommitID() types.CommitID {
	return types.CommitID{
		Version: st.tree.Version(),
		Hash:    st.tree.Hash(),
	}
}

// SetPruning panics as pruning options should be provided at initialization
// since IAVl accepts pruning options directly.
func (st *Store) SetPruning(_ pruningtypes.PruningOptions) {
	panic("cannot set pruning options on an initialized IAVL store")
}

// GetPruning panics as pruning options should be provided at initialization
// since IAVl accepts pruning options directly.
func (st *Store) GetPruning() pruningtypes.PruningOptions {
	panic("cannot get pruning options on an initialized IAVL store")
}

//// VersionExists returns whether or not a given version is stored.
//func (st *Store) VersionExists(version int64) bool {
//	return st.tree.VersionExists(version)
//}
//
//// GetAllVersions returns all versions in the iavl tree
//func (st *Store) GetAllVersions() []int {
//	return st.tree.AvailableVersions()
//}
//

func (st *Store) GetStoreType() types.StoreType { return types.StoreTypeIAVL }

func (st *Store) CacheWrap() types.CacheWrap { return cachekv.NewStore(st) }

func (st *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(st, w, tc))
}

// Set implements types.KVStore, creates a new key/value pair in the underlying IAVL tree.
func (st *Store) Set(key, value []byte) {
	types.AssertValidKey(key)
	types.AssertValidValue(value)
	_, err := st.tree.Set(key, value)
	if err != nil && st.logger != nil {
		st.logger.Error("iavl set error", "error", err.Error())
	}
}

// Get implements types.KVStore.
func (st *Store) Get(key []byte) []byte {
	defer st.metrics.MeasureSince("store", "iavl", "get")
	value, err := st.tree.Get(key)
	if err != nil {
		panic(err)
	}
	return value
}

// Has implements types.KVStore, returns true if the key exists in the underlying IAVL tree.
func (st *Store) Has(key []byte) (exists bool) {
	defer st.metrics.MeasureSince("store", "iavl", "has")
	has, err := st.tree.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

// Delete implements types.KVStore, removes the given key from the underlying IAVL tree.
func (st *Store) Delete(key []byte) {
	defer st.metrics.MeasureSince("store", "iavl", "delete")
	_, _, err := st.tree.Remove(key)
	if err != nil {
		panic(err)
	}
}

// DeleteVersionsTo deletes versions up to the given version from the MutableTree. An error
// is returned if any single version is invalid or the delete fails. All writes
// happen in a single batch with a single commit.
func (st *Store) DeleteVersionsTo(version int64) error {
	return st.tree.DeleteVersionsTo(version)
}

//
//// LoadVersionForOverwriting attempts to load a tree at a previously committed
//// version. Any versions greater than targetVersion will be deleted.
//func (st *Store) LoadVersionForOverwriting(targetVersion int64) error {
//	return st.tree.LoadVersionForOverwriting(targetVersion)
//}

// Iterator implements types.KVStore, returns an iterator from the underlying IAVL tree.
func (st *Store) Iterator(start, end []byte) types.Iterator {
	iterator, err := st.tree.Iterator(start, end, false)
	if err != nil {
		panic(err)
	}
	return iterator
}

// ReverseIterator implements types.KVStore, returns a reverse iterator from the underlying IAVL tree.
func (st *Store) ReverseIterator(start, end []byte) types.Iterator {
	iterator, err := st.tree.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return iterator
}

// SetInitialVersion sets the initial version of the IAVL tree. It is used when
// starting a new chain at an arbitrary height.
func (st *Store) SetInitialVersion(version int64) {
	panic("TODO")
}

//// Export exports the IAVL store at the given version, returning an iavl.Exporter for the tree.
//func (st *Store) Export(version int64) (*iavl.Exporter, error) {
//	istore, err := st.GetImmutable(version)
//	if err != nil {
//		return nil, errorsmod.Wrapf(err, "iavl export failed for version %v", version)
//	}
//	tree, ok := istore.tree.(*immutableTree)
//	if !ok || tree == nil {
//		return nil, fmt.Errorf("iavl export failed: unable to fetch tree for version %v", version)
//	}
//	return tree.Export()
//}

//// Import imports an IAVL tree at the given version, returning an iavl.Importer for importing.
//func (st *Store) Import(version int64) (*iavl.Importer, error) {
//	tree, ok := st.tree.(*iavl.MutableTree)
//	if !ok {
//		return nil, errors.New("iavl import failed: unable to find mutable tree")
//	}
//	return tree.Import(version)
//}

// Handle gatest the latest height, if height is 0
func getHeight(tree *iavlv2.Tree, req *types.RequestQuery) int64 {
	height := req.Height
	if height == 0 {
		panic("TODO")
		//latest := tree.Version()
		//if tree.VersionExists(latest - 1) {
		//	height = latest - 1
		//} else {
		//	height = latest
		//}
	}
	return height
}

// Query implements ABCI interface, allows queries
//
// by default we will return from (latest height -1),
// as we will have merkle proofs immediately (header height = data height + 1)
// If latest-1 is not present, use latest (which must be present)
// if you care to have the latest data to see a tx results, you must
// explicitly set the height you want to see
func (st *Store) Query(req *types.RequestQuery) (res *types.ResponseQuery, err error) {
	// TODO try to share code between this and iavl.Store.Query because this is a pretty large piece of code to duplicate
	defer st.metrics.MeasureSince("store", "iavl", "query")
	panic("TODO")

	//if len(req.Data) == 0 {
	//	return &types.ResponseQuery{}, errorsmod.Wrap(types.ErrTxDecode, "query cannot be zero length")
	//}
	//
	//tree := st.tree
	//
	//// store the height we chose in the response, with 0 being changed to the
	//// latest height
	//res = &types.ResponseQuery{
	//	Height: getHeight(tree, req),
	//}
	//
	//switch req.Path {
	//case "/key": // get by key
	//	key := req.Data // data holds the key bytes
	//
	//	res.Key = key
	//	if !st.VersionExists(res.Height) {
	//		res.Log = iavl.ErrVersionDoesNotExist.Error()
	//		break
	//	}
	//
	//	value, err := tree.GetVersioned(key, res.Height)
	//	if err != nil {
	//		panic(err)
	//	}
	//	res.Value = value
	//
	//	if !req.Prove {
	//		break
	//	}
	//
	//	// Continue to prove existence/absence of value
	//	// Must convert store.Tree to iavl.MutableTree with given version to use in CreateProof
	//	iTree, err := tree.GetImmutable(res.Height)
	//	if err != nil {
	//		// sanity check: If value for given version was retrieved, immutable tree must also be retrievable
	//		panic(fmt.Sprintf("version exists in store but could not retrieve corresponding versioned tree in store, %s", err.Error()))
	//	}
	//	mtree := &iavl.MutableTree{
	//		ImmutableTree: iTree,
	//	}
	//
	//	// get proof from tree and convert to merkle.Proof before adding to result
	//	res.ProofOps = getProofFromTree(mtree, req.Data, res.Value != nil)
	//
	//case "/subspace":
	//	pairs := kv.Pairs{
	//		Pairs: make([]kv.Pair, 0),
	//	}
	//
	//	subspace := req.Data
	//	res.Key = subspace
	//
	//	iterator := types.KVStorePrefixIterator(st, subspace)
	//	for ; iterator.Valid(); iterator.Next() {
	//		pairs.Pairs = append(pairs.Pairs, kv.Pair{Key: iterator.Key(), Value: iterator.Value()})
	//	}
	//	if err := iterator.Close(); err != nil {
	//		panic(fmt.Errorf("failed to close iterator: %w", err))
	//	}
	//
	//	bz, err := pairs.Marshal()
	//	if err != nil {
	//		panic(fmt.Errorf("failed to marshal KV pairs: %w", err))
	//	}
	//
	//	res.Value = bz
	//
	//default:
	//	return &types.ResponseQuery{}, errorsmod.Wrapf(types.ErrUnknownRequest, "unexpected query path: %v", req.Path)
	//}
	//
	//return res, err
}

//
//// TraverseStateChanges traverses the state changes between two versions and calls the given function.
//func (st *Store) TraverseStateChanges(startVersion, endVersion int64, fn func(version int64, changeSet *iavl.ChangeSet) error) error {
//	return st.tree.TraverseStateChanges(startVersion, endVersion, fn)
//}
//
//// Takes a MutableTree, a key, and a flag for creating existence or absence proof and returns the
//// appropriate merkle.Proof. Since this must be called after querying for the value, this function should never error
//// Thus, it will panic on error rather than returning it
//func getProofFromTree(tree *iavl.MutableTree, key []byte, exists bool) *cmtprotocrypto.ProofOps {
//	var (
//		commitmentProof *ics23.CommitmentProof
//		err             error
//	)
//
//	if exists {
//		// value was found
//		commitmentProof, err = tree.GetMembershipProof(key)
//		if err != nil {
//			// sanity check: If value was found, membership proof must be creatable
//			panic(fmt.Sprintf("unexpected value for empty proof: %s", err.Error()))
//		}
//	} else {
//		// value wasn't found
//		commitmentProof, err = tree.GetNonMembershipProof(key)
//		if err != nil {
//			// sanity check: If value wasn't found, nonmembership proof must be creatable
//			panic(fmt.Sprintf("unexpected error for nonexistence proof: %s", err.Error()))
//		}
//	}
//
//	op := types.NewIavlCommitmentOp(key, commitmentProof)
//	return &cmtprotocrypto.ProofOps{Ops: []cmtprotocrypto.ProofOp{op.ProofOp()}}
//}

// TODO we should make sure that layers above this proper call close
func (st *Store) Close() error {
	return st.tree.Close()
}
