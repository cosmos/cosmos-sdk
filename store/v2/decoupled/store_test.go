package decoupled

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	types "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

var (
	cacheSize = 100
	alohaData = map[string]string{
		"hello": "goodbye",
		"aloha": "shalom",
	}
)

func newStoreWithData(t *testing.T, db dbm.DBConnection, storeData map[string]string) *Store {
	store, err := NewStore(db, DefaultStoreConfig)
	require.NoError(t, err)

	for k, v := range storeData {
		store.Set([]byte(k), []byte(v))
	}
	return store
}

func newAlohaStore(t *testing.T, db dbm.DBConnection) *Store {
	return newStoreWithData(t, db, alohaData)
}

func TestGetSetHasDelete(t *testing.T) {
	store := newAlohaStore(t, memdb.NewDB())
	key := "hello"

	exists := store.Has([]byte(key))
	require.True(t, exists)

	require.EqualValues(t, []byte(alohaData[key]), store.Get([]byte(key)))

	value2 := "notgoodbye"
	store.Set([]byte(key), []byte(value2))

	require.EqualValues(t, value2, store.Get([]byte(key)))

	store.Delete([]byte(key))

	exists = store.Has([]byte(key))
	require.False(t, exists)
}

func TestStoreNoNilSet(t *testing.T) {
	store := newAlohaStore(t, memdb.NewDB())

	require.Panics(t, func() { store.Set(nil, []byte("value")) }, "setting a nil key should panic")
	require.Panics(t, func() { store.Set([]byte(""), []byte("value")) }, "setting an empty key should panic")

	require.Panics(t, func() { store.Set([]byte("key"), nil) }, "setting a nil value should panic")
}

func TestLoadStore(t *testing.T) {
	db := memdb.NewDB()
	store := newAlohaStore(t, db)
	store.Commit()

	store, err := LoadStore(db, DefaultStoreConfig)
	require.NoError(t, err)

	value := store.Get([]byte("hello"))
	require.Equal(t, []byte("goodbye"), value)

	// Loading with an initial version beyond the lowest should error
	opts := StoreConfig{InitialVersion: 5, Pruning: types.PruneNothing}
	store, err = LoadStore(db, opts)
	require.Error(t, err)
}

func TestIterators(t *testing.T) {
	store := newStoreWithData(t, memdb.NewDB(), map[string]string{
		string([]byte{0x00}):       "0",
		string([]byte{0x00, 0x00}): "0 0",
		string([]byte{0x00, 0x01}): "0 1",
		string([]byte{0x00, 0x02}): "0 2",
		string([]byte{0x01}):       "1",
	})

	var testCase = func(t *testing.T, iter types.Iterator, expected []string) {
		var i int
		for i = 0; iter.Valid(); iter.Next() {
			expectedValue := expected[i]
			value := iter.Value()
			require.EqualValues(t, string(value), expectedValue)
			i++
		}
		require.Equal(t, len(expected), i)
	}

	testCase(t, store.Iterator(nil, nil),
		[]string{"0", "0 0", "0 1", "0 2", "1"})
	testCase(t, store.Iterator([]byte{0x00}, nil),
		[]string{"0", "0 0", "0 1", "0 2", "1"})
	testCase(t, store.Iterator([]byte{0x00}, []byte{0x00, 0x01}),
		[]string{"0", "0 0"})
	testCase(t, store.Iterator([]byte{0x00}, []byte{0x01}),
		[]string{"0", "0 0", "0 1", "0 2"})
	testCase(t, store.Iterator([]byte{0x00, 0x01}, []byte{0x01}),
		[]string{"0 1", "0 2"})
	testCase(t, store.Iterator(nil, []byte{0x01}),
		[]string{"0", "0 0", "0 1", "0 2"})

	testCase(t, store.ReverseIterator(nil, nil),
		[]string{"1", "0 2", "0 1", "0 0", "0"})
	testCase(t, store.ReverseIterator([]byte{0x00}, nil),
		[]string{"1", "0 2", "0 1", "0 0", "0"})
	testCase(t, store.ReverseIterator([]byte{0x00}, []byte{0x00, 0x01}),
		[]string{"0 0", "0"})
	testCase(t, store.ReverseIterator([]byte{0x00}, []byte{0x01}),
		[]string{"0 2", "0 1", "0 0", "0"})
	testCase(t, store.ReverseIterator([]byte{0x00, 0x01}, []byte{0x01}),
		[]string{"0 2", "0 1"})
	testCase(t, store.ReverseIterator(nil, []byte{0x01}),
		[]string{"0 2", "0 1", "0 0", "0"})

	testCase(t, types.KVStorePrefixIterator(store, []byte{0}),
		[]string{"0", "0 0", "0 1", "0 2"})
	testCase(t, types.KVStoreReversePrefixIterator(store, []byte{0}),
		[]string{"0 2", "0 1", "0 0", "0"})
}

type unsavableDB struct{ *memdb.MemDB }

func (unsavableDB) SaveVersion(uint64) error { return errors.New("unsavable DB") }

func TestCommit(t *testing.T) {
	// Sanity test for Merkle hashing
	store := newStoreWithData(t, memdb.NewDB(), nil)
	idNew := store.Commit()
	store.Set([]byte{0x00}, []byte("a"))
	idOne := store.Commit()
	require.Equal(t, idNew.Version+1, idOne.Version)
	require.NotEqual(t, idNew.Hash, idOne.Hash)

	// Hash of emptied store is same as new store
	store.Delete([]byte{0x00})
	idEmptied := store.Commit()
	require.Equal(t, idNew.Hash, idEmptied.Hash)

	for i := byte(0); i < 10; i++ {
		store.Set([]byte{i}, []byte{i})
		id := store.Commit()
		lastid := store.LastCommitID()
		require.Equal(t, id.Hash, lastid.Hash)
		require.Equal(t, id.Version, lastid.Version)
	}

	// Storage commit is rolled back if Merkle commit fails
	opts := StoreConfig{MerkleDB: unsavableDB{memdb.NewDB()}, Pruning: types.PruneNothing}
	db := memdb.NewDB()
	store, err := NewStore(db, opts)
	require.NoError(t, err)
	require.Panics(t, func() { _ = store.Commit() })
	versions, err := db.Versions()
	require.NoError(t, err)
	require.Equal(t, 0, versions.Count())

	opts = StoreConfig{InitialVersion: 5, Pruning: types.PruneNothing}
	store, err = NewStore(memdb.NewDB(), opts)
	require.NoError(t, err)
	cid := store.Commit()
	require.Equal(t, int64(5), cid.Version)
}

func sliceToSet(slice []uint64) map[uint64]struct{} {
	res := make(map[uint64]struct{})
	for _, x := range slice {
		res[x] = struct{}{}
	}
	return res
}

func TestPruning(t *testing.T) {
	// Save versions up to 10 and verify pruning at final commit
	testCases := []struct {
		keepRecent uint64
		keepEvery  uint64
		interval   uint64
		kept       []uint64
	}{
		{2, 4, 10, []uint64{4, 8, 9, 10}},
		{0, 4, 10, []uint64{4, 8, 10}},
		{0, 0, 10, []uint64{10}},                           // everything
		{0, 1, 0, []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}, // nothing
	}

	for tci, tc := range testCases {
		opts := types.PruningOptions{tc.keepRecent, tc.keepEvery, tc.interval}
		db := memdb.NewDB()
		store, err := NewStore(db, StoreConfig{Pruning: opts})
		require.NoError(t, err)

		for i := byte(1); i <= 10; i++ {
			store.Set([]byte{i}, []byte{i})
			cid := store.Commit()
			latest := uint64(i)
			require.Equal(t, latest, uint64(cid.Version))
		}

		versions, err := db.Versions()
		require.NoError(t, err)
		kept := sliceToSet(tc.kept)
		for v := uint64(1); v <= 10; v++ {
			_, has := kept[v]
			require.Equal(t, has, versions.Exists(v), "Version = %v; tc #%d", v, tci)
		}
	}

	// Test pruning interval
	// Save up to 20th version while checking history at specific version checkpoints
	opts := types.PruningOptions{0, 5, 10}
	testCheckPoints := map[uint64][]uint64{
		5:  []uint64{1, 2, 3, 4, 5},
		10: []uint64{5, 10},
		15: []uint64{5, 10, 11, 12, 13, 14, 15},
		20: []uint64{5, 10, 15, 20},
	}
	db := memdb.NewDB()
	store, err := NewStore(db, StoreConfig{Pruning: opts})
	require.NoError(t, err)

	for i := byte(1); i <= 20; i++ {
		store.Set([]byte{i}, []byte{i})
		cid := store.Commit()
		latest := uint64(i)
		require.Equal(t, latest, uint64(cid.Version))

		kept, has := testCheckPoints[latest]
		if !has {
			continue
		}
		versions, err := db.Versions()
		require.NoError(t, err)
		keptMap := sliceToSet(kept)
		for v := uint64(1); v <= latest; v++ {
			_, has := keptMap[v]
			require.Equal(t, has, versions.Exists(v), "Version = %v; tc #%d", v, i)
		}
	}
}

func TestQuery(t *testing.T) {
	store := newStoreWithData(t, memdb.NewDB(), nil)

	k1, v1 := []byte("key1"), []byte("val1")
	k2, v2 := []byte("key2"), []byte("val2")
	v3 := []byte("val3")

	ksub := []byte("key")
	KVs0 := kv.Pairs{}
	KVs1 := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: k1, Value: v1},
			{Key: k2, Value: v2},
		},
	}
	KVs2 := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: k1, Value: v3},
			{Key: k2, Value: v2},
		},
	}

	valExpSubEmpty, err := KVs0.Marshal()
	require.NoError(t, err)

	valExpSub1, err := KVs1.Marshal()
	require.NoError(t, err)

	valExpSub2, err := KVs2.Marshal()
	require.NoError(t, err)

	cid := store.Commit()
	ver := cid.Version
	query := abci.RequestQuery{Path: "/key", Data: k1, Height: ver}
	querySub := abci.RequestQuery{Path: "/subspace", Data: ksub, Height: ver}

	// query subspace before anything set
	qres := store.Query(querySub)
	require.Equal(t, uint32(0), qres.Code)
	require.Equal(t, valExpSubEmpty, qres.Value)

	// set data
	store.Set(k1, v1)
	store.Set(k2, v2)

	// set data without commit, doesn't show up
	qres = store.Query(query)
	require.Equal(t, uint32(0), qres.Code)
	require.Nil(t, qres.Value)

	// commit it, but still don't see on old version
	cid = store.Commit()
	qres = store.Query(query)
	require.Equal(t, uint32(0), qres.Code)
	require.Nil(t, qres.Value)

	// but yes on the new version
	query.Height = cid.Version
	qres = store.Query(query)
	require.Equal(t, uint32(0), qres.Code)
	require.Equal(t, v1, qres.Value)

	// and for the subspace
	qres = store.Query(querySub)
	require.Equal(t, uint32(0), qres.Code)
	require.Equal(t, valExpSub1, qres.Value)

	// modify
	store.Set(k1, v3)
	cid = store.Commit()

	// query will return old values, as height is fixed
	qres = store.Query(query)
	require.Equal(t, uint32(0), qres.Code)
	require.Equal(t, v1, qres.Value)

	// update to latest in the query and we are happy
	query.Height = cid.Version
	qres = store.Query(query)
	require.Equal(t, uint32(0), qres.Code)
	require.Equal(t, v3, qres.Value)
	query2 := abci.RequestQuery{Path: "/key", Data: k2, Height: cid.Version}

	qres = store.Query(query2)
	require.Equal(t, uint32(0), qres.Code)
	require.Equal(t, v2, qres.Value)
	// and for the subspace
	qres = store.Query(querySub)
	require.Equal(t, uint32(0), qres.Code)
	require.Equal(t, valExpSub2, qres.Value)

	// default (height 0) will show latest -1
	query0 := abci.RequestQuery{Path: "/key", Data: k1}
	qres = store.Query(query0)
	require.Equal(t, uint32(0), qres.Code)
	require.Equal(t, v1, qres.Value)
}
