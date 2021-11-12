package root

import (
	"crypto/sha256"
	"io"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/prefix"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/store/v2/smt"
)

var (
	_ types.KVStore = (*Store)(nil)
)

// Store is a CommitKVStore which handles state storage and commitments as separate concerns,
// optionally using separate backing key-value DBs for each.
// Allows synchronized R/W access by locking.

// var DefaultStoreConfig = StoreConfig{Pruning: types.PruneDefault, StateCommitmentDB: nil}

// NewStore creates a new Store, or loads one if the DB contains existing data.

// Get implements KVStore.
func (s *Store) Get(key []byte) []byte {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	val, err := s.dataBucket.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Has implements KVStore.
func (s *Store) Has(key []byte) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	has, err := s.dataBucket.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

// Set implements KVStore.
func (s *Store) Set(key, value []byte) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	err := s.dataBucket.Set(key, value)
	if err != nil {
		panic(err)
	}
	s.stateCommitmentStore.Set(key, value)
	khash := sha256.Sum256(key)
	err = s.indexBucket.Set(khash[:], key)
	if err != nil {
		panic(err)
	}
}

// Delete implements KVStore.
func (s *Store) Delete(key []byte) {
	khash := sha256.Sum256(key)
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.stateCommitmentStore.Delete(key)
	_ = s.indexBucket.Delete(khash[:])
	_ = s.dataBucket.Delete(key)
}

type contentsIterator struct {
	dbm.Iterator
	valid bool
}

func newIterator(source dbm.Iterator) *contentsIterator {
	ret := &contentsIterator{Iterator: source}
	ret.Next()
	return ret
}

func (it *contentsIterator) Next()       { it.valid = it.Iterator.Next() }
func (it *contentsIterator) Valid() bool { return it.valid }

// Iterator implements KVStore.
func (s *Store) Iterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return newIterator(iter)
}

// ReverseIterator implements KVStore.
func (s *Store) ReverseIterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return newIterator(iter)
}

// GetStoreType implements Store.
func (s *Store) GetStoreType() types.StoreType {
	return types.StoreTypePersistent
}

func (s *Store) GetPruning() types.PruningOptions   { return s.Pruning }
func (s *Store) SetPruning(po types.PruningOptions) { s.Pruning = po }

func loadSMT(stateCommitmentTxn dbm.DBReadWriter, root []byte) *smt.Store {
	merkleNodes := prefix.NewPrefixReadWriter(stateCommitmentTxn, merkleNodePrefix)
	merkleValues := prefix.NewPrefixReadWriter(stateCommitmentTxn, merkleValuePrefix)
	return smt.LoadStore(merkleNodes, merkleValues, root)
}

func (s *Store) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(s)
}

func (s *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(s, w, tc))
}

func (s *Store) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	return cachekv.NewStore(listenkv.NewStore(s, storeKey, listeners))
}
