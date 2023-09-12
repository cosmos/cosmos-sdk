package branch

import (
	"io"

	"golang.org/x/exp/maps"
	"slices"

	"cosmossdk.io/store/v2"
)

var (
	_ store.KVStore         = (*Store)(nil)
	_ store.BranchedKVStore = (*Store)(nil)
)

// Store implements both a KVStore and BranchedKVStore interfaces. It is used to
// accumulate writes that can be later committed to backing SS and SC engines or
// discarded altogether. If a read is not found through an uncommitted write, it
// will be delegated to the SS backend.
type Store struct {
	// storage reflects backing storage for reads that are not found in uncommitted volatile state
	//
	// XXX/TODO: We use a SC backend here instead of SS since not all SS backends
	// may support reverse iteration (which is needed for state machine logic).
	storage store.Tree

	// storeKey reflects the store key used for the store
	storeKey string

	// parent reflects a parent store if branched (it may be nil)
	parent store.KVStore

	// changeSet reflects the uncommitted writes to the store
	//
	// Note, this field might be removed depending on how the branching fields
	// below are defined and used.
	changeSet map[string]store.KVPair

	// TODO: Fields for branching functionality. These fields should most likely
	// reflect what currently exists in cachekv.Store.
}

func New(storeKey string, sc store.Tree) store.KVStore {
	return &Store{
		storage:   sc,
		storeKey:  storeKey,
		changeSet: make(map[string]store.KVPair),
	}
}

func (s *Store) GetStoreType() store.StoreType {
	return store.StoreTypeBranch
}

// GetChangeSet returns the uncommitted writes to the store, ordered by key.
func (s *Store) GetChangeSet() *store.ChangeSet {
	keys := maps.Keys(s.changeSet)
	slices.Sort(keys)

	pairs := make([]store.KVPair, len(keys))
	for i, key := range keys {
		pairs[i] = s.changeSet[key]
	}

	return store.NewChangeSet(pairs...)
}

func (s *Store) Reset() {
	clear(s.changeSet)
}

func (s *Store) Branch() store.BranchedKVStore {
	panic("not implemented!")
}

func (s *Store) BranchWithTrace(w io.Writer, tc store.TraceContext) store.BranchedKVStore {
	panic("not implemented!")
}

func (s *Store) Iterator(start, end []byte) store.Iterator {
	panic("not implemented!")
}

func (s *Store) ReverseIterator(start, end []byte) store.Iterator {
	panic("not implemented!")
}

func (s *Store) Get(key []byte) []byte {
	panic("not implemented!")
}

func (s *Store) Has(key []byte) bool {
	panic("not implemented!")
}

func (s *Store) Set(key, value []byte) {
	panic("not implemented!")
}

func (s *Store) Delete(key []byte) {
	panic("not implemented!")
}

func (s *Store) Write() {
	panic("not implemented!")
}
