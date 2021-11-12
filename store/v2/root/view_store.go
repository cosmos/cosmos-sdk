package root

import (
	"errors"
	"io"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	prefixdb "github.com/cosmos/cosmos-sdk/db/prefix"
	util "github.com/cosmos/cosmos-sdk/internal"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	types "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/mem"
	"github.com/cosmos/cosmos-sdk/store/v2/smt"
	transkv "github.com/cosmos/cosmos-sdk/store/v2/transient"
)

var ErrReadOnly = errors.New("cannot modify read-only store")

func (s *viewStore) GetStateCommitmentStore() *smt.Store {
	return s.stateCommitmentStore
}

// Get implements KVStore.
func (s *viewStore) Get(key []byte) []byte {
	val, err := s.dataBucket.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Has implements KVStore.
func (s *viewStore) Has(key []byte) bool {
	has, err := s.dataBucket.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

// Set implements KVStore.
func (s *viewStore) Set(key []byte, value []byte) {
	panic(ErrReadOnly)
}

// Delete implements KVStore.
func (s *viewStore) Delete(key []byte) {
	panic(ErrReadOnly)
}

// Iterator implements KVStore.
func (s *viewStore) Iterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return newIterator(iter)
}

// ReverseIterator implements KVStore.
func (s *viewStore) ReverseIterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return newIterator(iter)
}

// GetStoreType implements Store.
func (s *viewStore) GetStoreType() types.StoreType {
	return types.StoreTypePersistent
}

func (st *viewStore) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(st)
}

func (st *viewStore) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(st, w, tc))
}

func (st *viewStore) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	return cachekv.NewStore(listenkv.NewStore(st, storeKey, listeners))
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
	root, err := stateView.Get(merkleRootKey)
	if err != nil {
		return
	}
	ret = &viewStore{
		stateView:            stateView,
		dataBucket:           prefixdb.NewPrefixReader(stateView, dataPrefix),
		indexBucket:          prefixdb.NewPrefixReader(stateView, indexPrefix),
		stateCommitmentView:  stateCommitmentView,
		stateCommitmentStore: loadSMT(dbm.ReaderAsReadWriter(stateCommitmentView), root),
	}
	// Now read this version's schema
	schemaView := prefixdb.NewPrefixReader(ret.stateView, schemaPrefix)
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
	ret.schema = pr.StoreSchema
	return
}

// if the schema indicates a mem/tran store, it's ignored
func (rv *viewStore) generic() rootGeneric { return rootGeneric{rv.schema, rv, nil, nil} }

func (rv *viewStore) GetKVStore(key types.StoreKey) types.KVStore {
	return rv.generic().getStore(key.Name())
}

func (rv *viewStore) CacheRootStore() types.CacheRootStore {
	return &cacheStore{
		CacheKVStore:  cachekv.NewStore(rv),
		mem:           cachekv.NewStore(mem.NewStore(memdb.NewDB())),
		tran:          cachekv.NewStore(transkv.NewStore(memdb.NewDB())),
		schema:        rv.schema,
		listenerMixin: &listenerMixin{},
		traceMixin:    &traceMixin{},
	}
}
