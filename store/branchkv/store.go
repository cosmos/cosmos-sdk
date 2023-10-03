package branchkv

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

	// storage reflects backing storage (SS) for reads that are not found in uncommitted volatile state
	storage store.VersionedDatabase

	// version indicates the latest version to handle reads falling through to SS
	version uint64

	// storeKey reflects the store key used for the store
	storeKey string

	// parent reflects a parent store if branched (it may be nil)
	parent store.KVStore

	// changeset reflects the uncommitted writes to the store
	changeset map[string]store.KVPair
}

func New(storeKey string, ss store.VersionedDatabase) (store.BranchedKVStore, error) {
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

func NewWithParent(parent store.KVStore) store.BranchedKVStore {
	return &Store{
		parent:    parent,
		storeKey:  parent.GetStoreKey(),
		changeset: make(map[string]store.KVPair),
	}
}

func (s *Store) GetStoreKey() string {
	return s.storeKey
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
		kvPair := s.changeset[key]
		pairs[i] = store.KVPair{
			Key:      []byte(key),
			Value:    slices.Clone(kvPair.Value),
			StoreKey: kvPair.StoreKey,
		}
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
	return NewWithParent(s)
}

func (s *Store) BranchWithTrace(w io.Writer, tc store.TraceContext) store.BranchedKVStore {
	panic("not implemented!")
}

func (s *Store) Has(key []byte) bool {
	store.AssertValidKey(key)

	s.mu.Lock()
	defer s.mu.Unlock()

	// if the write is present in the changeset, i.e. a dirty write, evaluate it
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
	store.AssertValidKey(key)

	s.mu.Lock()
	defer s.mu.Unlock()

	// if the write is present in the changeset, i.e. a dirty write, evaluate it
	if kvPair, ok := s.changeset[string(key)]; ok {
		if kvPair.Value == nil {
			return nil
		}

		return slices.Clone(kvPair.Value)
	}

	// if the store is branched, check the parent store
	if s.parent != nil {
		return s.parent.Get(key)
	}

	// otherwise, we fallback to SS
	bz, err := s.storage.Get(s.storeKey, s.version, key)
	if err != nil {
		panic(err)
	}

	return bz
}

func (s *Store) Set(key, value []byte) {
	store.AssertValidKey(key)
	store.AssertValidValue(value)

	s.mu.Lock()
	defer s.mu.Unlock()

	// omit the key as that can be inferred from the map key
	s.changeset[string(key)] = store.KVPair{Value: slices.Clone(value), StoreKey: s.storeKey}
}

func (s *Store) Delete(key []byte) {
	store.AssertValidKey(key)

	s.mu.Lock()
	defer s.mu.Unlock()

	// omit the key as that can be inferred from the map key
	s.changeset[string(key)] = store.KVPair{Value: nil, StoreKey: s.storeKey}
}

func (s *Store) Write() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Note, we're only flushing the writes up to the parent, if it exists. We are
	// not writing to the SS backend as that will happen in Commit().
	if s.parent != nil {
		keys := maps.Keys(s.changeset)
		slices.Sort(keys)

		// flush changes upstream to the parent in sorted order by key
		for _, key := range keys {
			kvPair := s.changeset[key]

			if kvPair.Value == nil {
				s.parent.Delete([]byte(key))
			} else {
				s.parent.Set([]byte(key), kvPair.Value)
			}
		}
	}
}

// Iterator creates an iterator over the domain [start, end), which walks over
// both the KVStore's changeset, i.e. dirty writes, and the parent iterator,
// which can either be another KVStore or the SS backend, at the same time.
//
// Note, writes that happen on the KVStore over an iterator will not affect the
// iterator. This is because when an iterator is created, it takes a current
// snapshot of the changeset.
func (s *Store) Iterator(start, end []byte) store.Iterator {
	var parentItr store.Iterator
	if s.parent != nil {
		parentItr = s.parent.Iterator(start, end)
	} else {
		var err error
		parentItr, err = s.storage.NewIterator(s.storeKey, s.version, start, end)
		if err != nil {
			panic(err)
		}
	}

	cs := maps.Clone(s.changeset)
	return newIterator(parentItr, start, end, cs, false)
}

// ReverseIterator creates a reverse iterator over the domain [start, end), which
// walks over both the KVStore's changeset, i.e. dirty writes, and the parent
// iterator, which can either be another KVStore or the SS backend, at the same
// time.
//
// Note, writes that happen on the KVStore over an iterator will not affect the
// iterator. This is because when an iterator is created, it takes a current
// snapshot of the changeset.
func (s *Store) ReverseIterator(start, end []byte) store.Iterator {
	var parentItr store.Iterator
	if s.parent != nil {
		parentItr = s.parent.ReverseIterator(start, end)
	} else {
		var err error
		parentItr, err = s.storage.NewReverseIterator(s.storeKey, s.version, start, end)
		if err != nil {
			panic(err)
		}
	}

	cs := maps.Clone(s.changeset)
	return newIterator(parentItr, start, end, cs, true)
}
