package multi

import (
	"errors"
	"io"

	dbm "github.com/cosmos/cosmos-sdk/db"
	prefixdb "github.com/cosmos/cosmos-sdk/db/prefix"
	util "github.com/cosmos/cosmos-sdk/internal"
	dbutil "github.com/cosmos/cosmos-sdk/internal/db"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	types "github.com/cosmos/cosmos-sdk/store/v2alpha1"
	"github.com/cosmos/cosmos-sdk/store/v2alpha1/smt"
)

var ErrReadOnly = errors.New("cannot modify read-only store")

func (s *viewSubstore) GetStateCommitmentStore() *smt.Store {
	return s.stateCommitmentStore
}

// Get implements KVStore.
func (s *viewSubstore) Get(key []byte) []byte {
	val, err := s.dataBucket.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Has implements KVStore.
func (s *viewSubstore) Has(key []byte) bool {
	has, err := s.dataBucket.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

// Set implements KVStore.
func (s *viewSubstore) Set(key []byte, value []byte) {
	panic(ErrReadOnly)
}

// Delete implements KVStore.
func (s *viewSubstore) Delete(key []byte) {
	panic(ErrReadOnly)
}

// Iterator implements KVStore.
func (s *viewSubstore) Iterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return dbutil.ToStoreIterator(iter)
}

// ReverseIterator implements KVStore.
func (s *viewSubstore) ReverseIterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return dbutil.ToStoreIterator(iter)
}

// GetStoreType implements Store.
func (s *viewSubstore) GetStoreType() types.StoreType {
	return types.StoreTypePersistent
}

func (s *viewSubstore) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(s)
}

func (s *viewSubstore) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(s, w, tc))
}

func (s *viewSubstore) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	return cachekv.NewStore(listenkv.NewStore(s, storeKey, listeners))
}

func (s *viewStore) getMerkleRoots() (ret map[string][]byte, err error) {
	ret = map[string][]byte{}
	for key := range s.schema {
		sub, has := s.substoreCache[key]
		if !has {
			sub, err = s.getSubstore(key)
			if err != nil {
				return
			}
		}
		ret[key] = sub.stateCommitmentStore.Root()
	}
	return
}

func (store *Store) getView(version int64) (ret *viewStore, err error) {
	stateView, err := store.stateDB.ReaderAt(uint64(version))
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, stateView.Discard(), "stateView.Discard also failed")
		}
	}()

	stateCommitmentView := stateView
	if store.StateCommitmentDB != nil {
		stateCommitmentView, err = store.StateCommitmentDB.ReaderAt(uint64(version))
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				err = util.CombineErrors(err, stateCommitmentView.Discard(), "stateCommitmentView.Discard also failed")
			}
		}()
	}
	// Now read this version's schema
	schemaView := prefixdb.NewReader(stateView, schemaPrefix)
	defer func() {
		if err != nil {
			err = util.CombineErrors(err, schemaView.Discard(), "schemaView.Discard also failed")
		}
	}()
	pr, err := readSavedSchema(schemaView)
	if err != nil {
		return
	}
	// The migrated contents and schema are not committed until the next store.Commit
	ret = &viewStore{
		stateView:           stateView,
		stateCommitmentView: stateCommitmentView,
		substoreCache:       map[string]*viewSubstore{},
		schema:              pr.StoreSchema,
	}
	return ret, err
}

func (s *viewStore) GetKVStore(skey types.StoreKey) types.KVStore {
	key := skey.Name()
	if _, has := s.schema[key]; !has {
		panic(ErrStoreNotFound(key))
	}
	ret, err := s.getSubstore(key)
	if err != nil {
		panic(err)
	}
	s.substoreCache[key] = ret
	return ret
}

// Reads but does not update substore cache
func (s *viewStore) getSubstore(key string) (*viewSubstore, error) {
	if cached, has := s.substoreCache[key]; has {
		return cached, nil
	}
	pfx := substorePrefix(key)
	stateR := prefixdb.NewReader(s.stateView, pfx)
	stateCommitmentR := prefixdb.NewReader(s.stateCommitmentView, pfx)
	rootHash, err := stateR.Get(merkleRootKey)
	if err != nil {
		return nil, err
	}
	return &viewSubstore{
		root:                 s,
		name:                 key,
		dataBucket:           prefixdb.NewReader(stateR, dataPrefix),
		indexBucket:          prefixdb.NewReader(stateR, indexPrefix),
		stateCommitmentStore: loadSMT(dbm.ReaderAsReadWriter(stateCommitmentR), rootHash),
	}, nil
}
