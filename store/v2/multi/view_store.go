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
	types "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/smt"
)

var ErrReadOnly = errors.New("cannot modify read-only store")

// Read-only store for querying past versions
type viewStore struct {
	stateView           dbm.DBReader
	stateCommitmentView dbm.DBReader
	substoreCache       map[string]*viewSubstore
	schema              StoreSchema
}

type viewSubstore struct {
	root                 *viewStore
	name                 string
	dataBucket           dbm.DBReader
	indexBucket          dbm.DBReader
	stateCommitmentStore *smt.Store
}

func (vs *viewStore) GetKVStore(skey types.StoreKey) types.KVStore {
	key := skey.Name()
	if _, has := vs.schema[key]; !has {
		panic(ErrStoreNotFound(key))
	}
	ret, err := vs.getSubstore(key)
	if err != nil {
		panic(err)
	}
	vs.substoreCache[key] = ret
	return ret
}

// Reads but does not update substore cache
func (vs *viewStore) getSubstore(key string) (*viewSubstore, error) {
	if cached, has := vs.substoreCache[key]; has {
		return cached, nil
	}
	pfx := substorePrefix(key)
	stateR := prefixdb.NewPrefixReader(vs.stateView, pfx)
	stateCommitmentR := prefixdb.NewPrefixReader(vs.stateCommitmentView, pfx)
	rootHash, err := stateR.Get(merkleRootKey)
	if err != nil {
		return nil, err
	}
	return &viewSubstore{
		root:                 vs,
		name:                 key,
		dataBucket:           prefixdb.NewPrefixReader(stateR, dataPrefix),
		indexBucket:          prefixdb.NewPrefixReader(stateR, indexPrefix),
		stateCommitmentStore: loadSMT(dbm.ReaderAsReadWriter(stateCommitmentR), rootHash),
	}, nil
}

// CacheWrap implements MultiStore.
// Because this store is a read-only view, the returned store's Write operation is a no-op.
func (vs *viewStore) CacheWrap() types.CacheMultiStore {
	return noopCacheStore{newCacheStore(vs)}
}

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
	return dbutil.DBToStoreIterator(iter)
}

// ReverseIterator implements KVStore.
func (s *viewSubstore) ReverseIterator(start, end []byte) types.Iterator {
	iter, err := s.dataBucket.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return dbutil.DBToStoreIterator(iter)
}

// GetStoreType implements Store.
func (s *viewSubstore) GetStoreType() types.StoreType {
	return types.StoreTypePersistent
}

func (st *viewSubstore) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(st)
}

func (st *viewSubstore) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(st, w, tc))
}

func (st *viewSubstore) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	return cachekv.NewStore(listenkv.NewStore(st, storeKey, listeners))
}

func (s *viewStore) getMerkleRoots() (ret map[string][]byte, err error) {
	ret = map[string][]byte{}
	for key, _ := range s.schema {
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
	schemaView := prefixdb.NewPrefixReader(stateView, schemaPrefix)
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
	return
}
