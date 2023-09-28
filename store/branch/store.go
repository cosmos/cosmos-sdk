package branch

import (
	"io"
	"slices"
	"sync"

	"golang.org/x/exp/maps"

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
	mu sync.Mutex

	// TODO: Consider wrapping storage (SS) in a KVStore wrapper to avoid having to
	// check SS and parent separately.

	// storage reflects backing storage (SS) for reads that are not found in uncommitted volatile state
	storage store.VersionedDatabase

	// version indicates the latest version to handle reads falling through to SS
	version uint64

	// storeKey reflects the store key used for the store
	storeKey string

	// parent reflects a parent store if branched (it may be nil)
	parent store.KVStore

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// changeset reflects the uncommitted writes to the store
	changeset map[string]store.KVPair
}

func New(storeKey string, ss store.VersionedDatabase) (store.KVStore, error) {
	latestVersion, err := ss.GetLatestVersion()
	if err != nil {
		return nil, err
	}

	return &Store{
		storage:   ss,
		storeKey:  storeKey,
		version:   latestVersion,
		changeset: make(map[string]store.KVPair),
	}, nil
}

func (s *Store) GetStoreType() store.StoreType {
	return store.StoreTypeBranch
}

// GetChangeset returns the uncommitted writes to the store, ordered by key.
func (s *Store) GetChangeset() *store.Changeset {
	keys := maps.Keys(s.changeset)
	slices.Sort(keys)

	pairs := make([]store.KVPair, len(keys))
	for i, key := range keys {
		pairs[i] = s.changeset[key]
	}

	return store.NewChangeSet(pairs...)
}

func (s *Store) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	latestVersion, err := s.storage.GetLatestVersion()
	if err != nil {
		return err
	}

	clear(s.changeset)
	s.version = latestVersion

	return nil
}

func (s *Store) Branch() store.BranchedKVStore {
	panic("not implemented!")
}

func (s *Store) BranchWithTrace(w io.Writer, tc store.TraceContext) store.BranchedKVStore {
	panic("not implemented!")
}

func (s *Store) Has(key []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	store.AssertValidKey(key)

	// if the write is present in the changeset, i.e. a dirty write, return it
	if kvPair, ok := s.changeset[string(key)]; ok {
		// a non-nil value indicates presence
		return kvPair.Value != nil
	}

	// if the store is branched, check the parent store
	if s.parent != nil {
		return s.parent.Has(key)
	}

	// otherwise, we fallback to SS
	ok, err := s.storage.Has(s.storeKey, s.version, key)
	if err != nil {
		panic(err)
	}

	return ok
}

func (s *Store) Get(key []byte) []byte {
	panic("not implemented!")
}

func (s *Store) Set(key, value []byte) {
	panic("not implemented!")
}

func (s *Store) Delete(key []byte) {
	panic("not implemented!")
}

func (s *Store) Iterator(start, end []byte) store.Iterator {
	panic("not implemented!")
}

func (s *Store) ReverseIterator(start, end []byte) store.Iterator {
	panic("not implemented!")
}

func (s *Store) Write() {
	panic("not implemented!")
}
