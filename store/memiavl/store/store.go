package memiavlstore

import (
	stderrors "errors"
	"fmt"
	"io"

	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	ics23 "github.com/cosmos/ics23/go"
	"github.com/crypto-org-chain/cronos/memiavl"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/store/cachekv"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/tracekv"
	"cosmossdk.io/store/types"
)

var (
	_ types.KVStore       = (*Store)(nil)
	_ types.CommitStore   = (*Store)(nil)
	_ types.CommitKVStore = (*Store)(nil)
	_ types.Queryable     = (*Store)(nil)
)

// Store Implements types.KVStore and CommitKVStore.
type Store struct {
	tree   *memiavl.Tree
	logger log.Logger

	changeSet memiavl.ChangeSet
}

func New(tree *memiavl.Tree, logger log.Logger) *Store {
	return &Store{tree: tree, logger: logger}
}

func (st *Store) SetTree(tree *memiavl.Tree) {
	st.tree = tree
}

func (st *Store) Commit() types.CommitID {
	panic("memiavl store is not supposed to be committed alone")
}

func (st *Store) LastCommitID() types.CommitID {
	hash := st.tree.RootHash()
	return types.CommitID{
		Version: st.tree.Version(),
		Hash:    hash,
	}
}

// SetPruning panics as pruning options should be provided at initialization
// since IAVl accepts pruning options directly.
func (st *Store) SetPruning(_ pruningtypes.PruningOptions) {
	panic("cannot set pruning options on an initialized IAVL store")
}

// SetPruning panics as pruning options should be provided at initialization
// since IAVl accepts pruning options directly.
func (st *Store) GetPruning() pruningtypes.PruningOptions {
	panic("cannot get pruning options on an initialized IAVL store")
}

// Implements Store.
func (st *Store) GetStoreType() types.StoreType {
	return types.StoreTypeIAVL
}

func (st *Store) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(st)
}

// CacheWrapWithTrace implements the Store interface.
func (st *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(st, w, tc))
}

// Implements types.KVStore.
//
// we assume Set is only called in `Commit`, so the written state is only visible after commit.
func (st *Store) Set(key, value []byte) {
	st.changeSet.Pairs = append(st.changeSet.Pairs, &memiavl.KVPair{
		Key: key, Value: value,
	})
}

// Implements types.KVStore.
func (st *Store) Get(key []byte) []byte {
	return st.tree.Get(key)
}

// Implements types.KVStore.
func (st *Store) Has(key []byte) bool {
	return st.tree.Has(key)
}

// Implements types.KVStore.
//
// we assume Delete is only called in `Commit`, so the written state is only visible after commit.
func (st *Store) Delete(key []byte) {
	st.changeSet.Pairs = append(st.changeSet.Pairs, &memiavl.KVPair{
		Key: key, Delete: true,
	})
}

func (st *Store) Iterator(start, end []byte) types.Iterator {
	return st.tree.Iterator(start, end, true)
}

func (st *Store) ReverseIterator(start, end []byte) types.Iterator {
	return st.tree.Iterator(start, end, false)
}

// SetInitialVersion sets the initial version of the IAVL tree. It is used when
// starting a new chain at an arbitrary height.
// implements interface StoreWithInitialVersion
func (st *Store) SetInitialVersion(version int64) {
	panic("memiavl store's SetInitialVersion is not supposed to be called directly")
}

// PopChangeSet returns the change set and clear it
func (st *Store) PopChangeSet() memiavl.ChangeSet {
	cs := st.changeSet
	st.changeSet = memiavl.ChangeSet{}
	return cs
}

func (st *Store) Query(req *types.RequestQuery) (res *types.ResponseQuery, err error) {
	if len(req.Data) == 0 {
		return nil, errors.Wrap(types.ErrTxDecode, "query cannot be zero length")
	}

	if req.Height > 0 && req.Height != st.tree.Version() {
		return nil, stderrors.New("invalid height")
	}

	res = &types.ResponseQuery{
		Height: st.tree.Version(),
	}

	switch req.Path {
	case "/key": // get by key
		res.Key = req.Data // data holds the key bytes
		res.Value = st.tree.Get(res.Key)

		if !req.Prove {
			break
		}

		// get proof from tree and convert to merkle.Proof before adding to result
		res.ProofOps = getProofFromTree(st.tree, req.Data, res.Value != nil)
	case "/subspace":
		pairs := memiavl.Pairs{
			Pairs: make([]memiavl.Pair, 0),
		}

		subspace := req.Data
		res.Key = subspace

		iterator := types.KVStorePrefixIterator(st, subspace)
		for ; iterator.Valid(); iterator.Next() {
			pairs.Pairs = append(pairs.Pairs, memiavl.Pair{Key: iterator.Key(), Value: iterator.Value()})
		}
		iterator.Close()

		bz, err := pairs.Marshal()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal KV pairs")
		}

		res.Value = bz
	default:
		return nil, errors.Wrapf(types.ErrUnknownRequest, "unexpected query path: %v", req.Path)
	}

	return res, nil
}

func (st *Store) WorkingHash() []byte {
	return st.tree.RootHash()
}

// Takes a MutableTree, a key, and a flag for creating existence or absence proof and returns the
// appropriate merkle.Proof. Since this must be called after querying for the value, this function should never error
// Thus, it will panic on error rather than returning it
func getProofFromTree(tree *memiavl.Tree, key []byte, exists bool) *cmtprotocrypto.ProofOps {
	var (
		commitmentProof *ics23.CommitmentProof
		err             error
	)

	if exists {
		// value was found
		commitmentProof, err = tree.GetMembershipProof(key)
		if err != nil {
			// sanity check: If value was found, membership proof must be creatable
			panic(fmt.Sprintf("unexpected value for empty proof: %s", err.Error()))
		}
	} else {
		// value wasn't found
		commitmentProof, err = tree.GetNonMembershipProof(key)
		if err != nil {
			// sanity check: If value wasn't found, nonmembership proof must be creatable
			panic(fmt.Sprintf("unexpected error for nonexistence proof: %s", err.Error()))
		}
	}

	op := types.NewIavlCommitmentOp(key, commitmentProof)
	return &cmtprotocrypto.ProofOps{Ops: []cmtprotocrypto.ProofOp{op.ProofOp()}}
}
