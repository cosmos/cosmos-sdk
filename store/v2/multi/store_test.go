package root

import (
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	skey_1  = types.NewKVStoreKey("store1")
	skey_2  = types.NewKVStoreKey("store2")
	skey_3  = types.NewKVStoreKey("store3")
	skey_4  = types.NewKVStoreKey("store4")
	skey_1b = types.NewKVStoreKey("store1b")
	skey_2b = types.NewKVStoreKey("store2b")
	skey_3b = types.NewKVStoreKey("store3b")
)

func simpleStoreConfig(t *testing.T) StoreConfig {
	opts := DefaultStoreConfig()
	require.NoError(t, opts.RegisterSubstore(skey_1.Name(), types.StoreTypePersistent))
	return opts
}

func storeConfig123(t *testing.T) StoreConfig {
	opts := DefaultStoreConfig()
	opts.Pruning = types.PruneNothing
	require.NoError(t, opts.RegisterSubstore(skey_1.Name(), types.StoreTypePersistent))
	require.NoError(t, opts.RegisterSubstore(skey_2.Name(), types.StoreTypePersistent))
	require.NoError(t, opts.RegisterSubstore(skey_3.Name(), types.StoreTypePersistent))
	return opts
}

func newSubStoreWithData(t *testing.T, db dbm.DBConnection, storeData map[string]string) (*Store, types.KVStore) {
	root, err := NewStore(db, simpleStoreConfig(t))
	require.NoError(t, err)

	store := root.GetKVStore(skey_1)
	for k, v := range storeData {
		store.Set([]byte(k), []byte(v))
	}
	return root, store
}

func TestGetSetHasDelete(t *testing.T) {
	_, store := newSubStoreWithData(t, memdb.NewDB(), alohaData)
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

	require.Panics(t, func() { store.Get(nil) }, "Get(nil key) should panic")
	require.Panics(t, func() { store.Get([]byte{}) }, "Get(empty key) should panic")
	require.Panics(t, func() { store.Has(nil) }, "Has(nil key) should panic")
	require.Panics(t, func() { store.Has([]byte{}) }, "Has(empty key) should panic")
	require.Panics(t, func() { store.Set(nil, []byte("value")) }, "Set(nil key) should panic")
	require.Panics(t, func() { store.Set([]byte{}, []byte("value")) }, "Set(empty key) should panic")
	require.Panics(t, func() { store.Set([]byte("key"), nil) }, "Set(nil value) should panic")
	sub := store.(*substore)
	sub.indexBucket = rwCrudFails{sub.indexBucket, nil}
	require.Panics(t, func() {
		store.Set([]byte("key"), []byte("value"))
	}, "Set() when index fails should panic")
}

func TestConstructors(t *testing.T) {
	db := memdb.NewDB()

	store, err := NewStore(db, simpleStoreConfig(t))
	require.NoError(t, err)
	_ = store.GetKVStore(skey_1)
	store.Commit()
	require.NoError(t, store.Close())

	t.Run("fail to load if InitialVersion > lowest existing version", func(t *testing.T) {
		opts := StoreConfig{InitialVersion: 5, Pruning: types.PruneNothing}
		store, err = NewStore(db, opts)
		require.Error(t, err)
		db.Close()
	})

	t.Run("can't load store when db.Versions fails", func(t *testing.T) {
		store, err = NewStore(dbVersionsFails{memdb.NewDB()}, DefaultStoreConfig())
		require.Error(t, err)
		store, err = NewStore(db, StoreConfig{StateCommitmentDB: dbVersionsFails{memdb.NewDB()}})
		require.Error(t, err)
	})

	db = memdb.NewDB()
	merkledb := memdb.NewDB()
	w := db.Writer()
	t.Run("can't use a DB with open writers", func(t *testing.T) {
		store, err = NewStore(db, DefaultStoreConfig())
		require.Error(t, err)
		w.Discard()
		w = merkledb.Writer()
		store, err = NewStore(db, StoreConfig{StateCommitmentDB: merkledb})
		require.Error(t, err)
		w.Discard()
	})

	t.Run("can't use DBs with different version history", func(t *testing.T) {
		merkledb.SaveNextVersion()
		store, err = NewStore(db, StoreConfig{StateCommitmentDB: merkledb})
		require.Error(t, err)
	})
	merkledb.Close()

	t.Run("can't load existing store if we can't access root hash", func(t *testing.T) {
		store, err = NewStore(db, simpleStoreConfig(t))
		require.NoError(t, err)
		store.Commit()
		require.NoError(t, store.Close())
		// ...whether because root is misssing
		w = db.Writer()
		s1RootKey := append(contentPrefix, substorePrefix(skey_1.Name())...)
		s1RootKey = append(s1RootKey, merkleRootKey...)
		w.Delete(s1RootKey)
		w.Commit()
		db.SaveNextVersion()
		store, err = NewStore(db, DefaultStoreConfig())
		require.Error(t, err)
		// ...or, because of an error
		store, err = NewStore(dbRWCrudFails{db}, DefaultStoreConfig())
		require.Error(t, err)
	})
}

func TestIterators(t *testing.T) {
	_, store := newSubStoreWithData(t, memdb.NewDB(), map[string]string{
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
	testCase(t, store.Iterator([]byte{0}, nil),
		[]string{"0", "0 0", "0 1", "0 2", "1"})
	testCase(t, store.Iterator([]byte{0}, []byte{0, 1}),
		[]string{"0", "0 0"})
	testCase(t, store.Iterator([]byte{0}, []byte{1}),
		[]string{"0", "0 0", "0 1", "0 2"})
	testCase(t, store.Iterator([]byte{0, 1}, []byte{1}),
		[]string{"0 1", "0 2"})
	testCase(t, store.Iterator(nil, []byte{1}),
		[]string{"0", "0 0", "0 1", "0 2"})
	testCase(t, store.Iterator([]byte{0}, []byte{0}), []string{}) // start = end
	testCase(t, store.Iterator([]byte{1}, []byte{0}), []string{}) // start > end

	testCase(t, store.ReverseIterator(nil, nil),
		[]string{"1", "0 2", "0 1", "0 0", "0"})
	testCase(t, store.ReverseIterator([]byte{0}, nil),
		[]string{"1", "0 2", "0 1", "0 0", "0"})
	testCase(t, store.ReverseIterator([]byte{0}, []byte{0, 1}),
		[]string{"0 0", "0"})
	testCase(t, store.ReverseIterator([]byte{0}, []byte{1}),
		[]string{"0 2", "0 1", "0 0", "0"})
	testCase(t, store.ReverseIterator([]byte{0, 1}, []byte{1}),
		[]string{"0 2", "0 1"})
	testCase(t, store.ReverseIterator(nil, []byte{1}),
		[]string{"0 2", "0 1", "0 0", "0"})
	testCase(t, store.ReverseIterator([]byte{0}, []byte{0}), []string{}) // start = end
	testCase(t, store.ReverseIterator([]byte{1}, []byte{0}), []string{}) // start > end

	testCase(t, types.KVStorePrefixIterator(store, []byte{0}),
		[]string{"0", "0 0", "0 1", "0 2"})
	testCase(t, types.KVStoreReversePrefixIterator(store, []byte{0}),
		[]string{"0 2", "0 1", "0 0", "0"})

	require.Panics(t, func() { store.Iterator([]byte{}, nil) }, "Iterator(empty key) should panic")
	require.Panics(t, func() { store.Iterator(nil, []byte{}) }, "Iterator(empty key) should panic")
	require.Panics(t, func() { store.ReverseIterator([]byte{}, nil) }, "Iterator(empty key) should panic")
	require.Panics(t, func() { store.ReverseIterator(nil, []byte{}) }, "Iterator(empty key) should panic")
}

func TestCommit(t *testing.T) {
	testBasic := func(opts StoreConfig) {
		db := memdb.NewDB()
		store, err := NewStore(db, opts)
		require.NoError(t, err)
		require.Zero(t, store.LastCommitID())
		idNew := store.Commit()

		// Adding one record changes the hash
		s1 := store.GetKVStore(skey_1)
		s1.Set([]byte{0}, []byte{0})
		idOne := store.Commit()
		require.Equal(t, idNew.Version+1, idOne.Version)
		require.NotEqual(t, idNew.Hash, idOne.Hash)

		// Hash of emptied store is same as new store
		s1.Delete([]byte{0})
		idEmptied := store.Commit()
		require.Equal(t, idNew.Hash, idEmptied.Hash)

		previd := idOne
		for i := byte(1); i < 5; i++ {
			s1.Set([]byte{i}, []byte{i})
			id := store.Commit()
			lastid := store.LastCommitID()
			require.Equal(t, id.Hash, lastid.Hash)
			require.Equal(t, id.Version, lastid.Version)
			require.NotEqual(t, previd.Hash, id.Hash)
			require.NotEqual(t, previd.Version, id.Version)
		}
	}
	basicOpts := simpleStoreConfig(t)
	basicOpts.Pruning = types.PruneNothing
	t.Run("sanity tests for Merkle hashing", func(t *testing.T) {
		testBasic(basicOpts)
	})
	t.Run("sanity tests for Merkle hashing with separate DBs", func(t *testing.T) {
		basicOpts.StateCommitmentDB = memdb.NewDB()
		testBasic(basicOpts)
	})

	// test that we can recover from a failed commit
	testFailedCommit := func(t *testing.T,
		store *Store,
		db dbm.DBConnection,
		opts StoreConfig) {
		if db == nil {
			db = store.stateDB
		}
		s1 := store.GetKVStore(skey_1)
		s1.Set([]byte{0}, []byte{0})
		require.Panics(t, func() { store.Commit() })
		require.NoError(t, store.Close())

		// No version should be saved in the backing DB(s)
		versions, _ := db.Versions()
		require.Equal(t, 0, versions.Count())
		if store.StateCommitmentDB != nil {
			versions, _ = store.StateCommitmentDB.Versions()
			require.Equal(t, 0, versions.Count())
		}

		// The store should now be reloaded successfully
		store, err := NewStore(db, opts)
		require.NoError(t, err)
		s1 = store.GetKVStore(skey_1)
		require.Nil(t, s1.Get([]byte{0}))
		require.NoError(t, store.Close())
	}

	opts := simpleStoreConfig(t)
	opts.Pruning = types.PruneNothing

	// Ensure Store's commit is rolled back in each failure case...
	t.Run("recover after failed Commit", func(t *testing.T) {
		store, err := NewStore(dbRWCommitFails{memdb.NewDB()}, opts)
		require.NoError(t, err)
		testFailedCommit(t, store, nil, opts)
	})
	// If SaveVersion and Revert both fail during Store.Commit, the DB will contain
	// committed data that belongs to no version: non-atomic behavior from the Store user's perspective.
	// So, that data must be reverted when the store is reloaded.
	t.Run("recover after failed SaveVersion and Revert", func(t *testing.T) {
		var db dbm.DBConnection
		db = dbSaveVersionFails{memdb.NewDB()}
		// Revert should succeed in initial NewStore call, but fail during Commit
		db = dbRevertFails{db, []bool{false, true}}
		store, err := NewStore(db, opts)
		require.NoError(t, err)
		testFailedCommit(t, store, nil, opts)
	})
	// Repeat the above for StateCommitmentDB
	t.Run("recover after failed StateCommitmentDB Commit", func(t *testing.T) {
		opts.StateCommitmentDB = dbRWCommitFails{memdb.NewDB()}
		store, err := NewStore(memdb.NewDB(), opts)
		require.NoError(t, err)
		testFailedCommit(t, store, nil, opts)
	})
	t.Run("recover after failed StateCommitmentDB SaveVersion and Revert", func(t *testing.T) {
		var db dbm.DBConnection
		db = dbSaveVersionFails{memdb.NewDB()}
		db = dbRevertFails{db, []bool{false, true}}
		opts.StateCommitmentDB = db
		store, err := NewStore(memdb.NewDB(), opts)
		require.NoError(t, err)
		testFailedCommit(t, store, nil, opts)
	})

	opts = simpleStoreConfig(t)
	t.Run("recover after stateDB.Versions error triggers failure", func(t *testing.T) {
		db := memdb.NewDB()
		store, err := NewStore(db, opts)
		require.NoError(t, err)
		store.stateDB = dbVersionsFails{store.stateDB}
		testFailedCommit(t, store, db, opts)
	})
	t.Run("recover after stateTxn.Set error triggers failure", func(t *testing.T) {
		store, err := NewStore(memdb.NewDB(), opts)
		require.NoError(t, err)
		store.stateTxn = rwCrudFails{store.stateTxn, merkleRootKey}
		testFailedCommit(t, store, nil, opts)
	})

	t.Run("stateDB.DeleteVersion error triggers failure", func(t *testing.T) {
		opts.StateCommitmentDB = memdb.NewDB()
		store, err := NewStore(memdb.NewDB(), opts)
		require.NoError(t, err)
		store.stateCommitmentTxn = rwCommitFails{store.stateCommitmentTxn}
		store.stateDB = dbDeleteVersionFails{store.stateDB}
		require.Panics(t, func() { store.Commit() })
	})
	t.Run("height overflow triggers failure", func(t *testing.T) {
		opts.StateCommitmentDB = nil
		opts.InitialVersion = math.MaxInt64
		opts.Pruning = types.PruneNothing
		store, err := NewStore(memdb.NewDB(), opts)
		require.NoError(t, err)
		require.Equal(t, int64(math.MaxInt64), store.Commit().Version)
		require.Panics(t, func() { store.Commit() })
		require.Equal(t, int64(math.MaxInt64), store.LastCommitID().Version) // version history not modified
	})

	t.Run("first commit version matches InitialVersion", func(t *testing.T) {
		opts = simpleStoreConfig(t)
		opts.InitialVersion = 5
		opts.Pruning = types.PruneNothing
		opts.StateCommitmentDB = memdb.NewDB()
		store, err := NewStore(memdb.NewDB(), opts)
		require.NoError(t, err)
		require.Equal(t, int64(5), store.Commit().Version)
	})

	// test improbable failures to fill out test coverage
	opts = simpleStoreConfig(t)
	store, err := NewStore(memdb.NewDB(), opts)
	require.NoError(t, err)
	store.Commit()
	store.stateDB = dbVersionsFails{store.stateDB}
	require.Panics(t, func() { store.LastCommitID() })

	opts = simpleStoreConfig(t)
	opts.StateCommitmentDB = memdb.NewDB()
	store, err = NewStore(memdb.NewDB(), opts)
	require.NoError(t, err)
	store.Commit()
	store.stateTxn = rwCrudFails{store.stateTxn, nil}
	require.Panics(t, func() { store.LastCommitID() })
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
		types.PruningOptions
		kept []uint64
	}{
		{types.PruningOptions{2, 4, 10}, []uint64{4, 8, 9, 10}},
		{types.PruningOptions{0, 4, 10}, []uint64{4, 8, 10}},
		{types.PruneEverything, []uint64{10}},
		{types.PruneNothing, []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}

	for tci, tc := range testCases {
		dbs := []dbm.DBConnection{memdb.NewDB(), memdb.NewDB()}
		opts := simpleStoreConfig(t)
		opts.Pruning = tc.PruningOptions
		opts.StateCommitmentDB = dbs[1]
		store, err := NewStore(dbs[0], opts)
		require.NoError(t, err)

		s1 := store.GetKVStore(skey_1)
		for i := byte(1); i <= 10; i++ {
			s1.Set([]byte{i}, []byte{i})
			cid := store.Commit()
			latest := uint64(i)
			require.Equal(t, latest, uint64(cid.Version))
		}

		for _, db := range dbs {
			versions, err := db.Versions()
			require.NoError(t, err)
			kept := sliceToSet(tc.kept)
			for v := uint64(1); v <= 10; v++ {
				_, has := kept[v]
				require.Equal(t, has, versions.Exists(v), "Version = %v; tc #%d", v, tci)
			}
		}
	}

	// Test pruning interval
	// Save up to 20th version while checking history at specific version checkpoints
	testCheckPoints := map[uint64][]uint64{
		5:  []uint64{1, 2, 3, 4, 5},
		10: []uint64{5, 10},
		15: []uint64{5, 10, 11, 12, 13, 14, 15},
		20: []uint64{5, 10, 15, 20},
	}
	db := memdb.NewDB()
	opts := simpleStoreConfig(t)
	opts.Pruning = types.PruningOptions{0, 5, 10}
	store, err := NewStore(db, opts)
	require.NoError(t, err)

	for i := byte(1); i <= 20; i++ {
		store.GetKVStore(skey_1).Set([]byte{i}, []byte{i})
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

func queryPath(skey types.StoreKey, endp string) string { return "/" + skey.Name() + endp }

func TestQuery(t *testing.T) {
	k1, v1 := []byte("k1"), []byte("v1")
	k2, v2 := []byte("k2"), []byte("v2")
	v3 := []byte("v3")

	ksub := []byte("k")
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

	store, err := NewStore(memdb.NewDB(), simpleStoreConfig(t))
	require.NoError(t, err)
	cid := store.Commit()
	ver := cid.Version
	query := abci.RequestQuery{Path: queryPath(skey_1, "/key"), Data: k1, Height: ver}
	querySub := abci.RequestQuery{Path: queryPath(skey_1, "/subspace"), Data: ksub, Height: ver}
	queryHeight0 := abci.RequestQuery{Path: queryPath(skey_1, "/key"), Data: k1}

	// query subspace before anything set
	qres := store.Query(querySub)
	require.True(t, qres.IsOK(), qres.Log)
	require.Equal(t, valExpSubEmpty, qres.Value)

	sub := store.GetKVStore(skey_1)
	require.NotNil(t, sub)
	// set data
	sub.Set(k1, v1)
	sub.Set(k2, v2)

	t.Run("basic queries", func(t *testing.T) {
		// set data without commit, doesn't show up
		qres = store.Query(query)
		require.True(t, qres.IsOK(), qres.Log)
		require.Nil(t, qres.Value)

		// commit it, but still don't see on old version
		cid = store.Commit()
		qres = store.Query(query)
		require.True(t, qres.IsOK(), qres.Log)
		require.Nil(t, qres.Value)

		// but yes on the new version
		query.Height = cid.Version
		qres = store.Query(query)
		require.True(t, qres.IsOK(), qres.Log)
		require.Equal(t, v1, qres.Value)
		// and for the subspace
		querySub.Height = cid.Version
		qres = store.Query(querySub)
		require.True(t, qres.IsOK(), qres.Log)
		require.Equal(t, valExpSub1, qres.Value)

		// modify
		sub.Set(k1, v3)
		cid = store.Commit()

		// query will return old values, as height is fixed
		qres = store.Query(query)
		require.True(t, qres.IsOK(), qres.Log)
		require.Equal(t, v1, qres.Value)

		// update to latest height in the query and we are happy
		query.Height = cid.Version
		qres = store.Query(query)
		require.True(t, qres.IsOK(), qres.Log)
		require.Equal(t, v3, qres.Value)
		// try other key
		query2 := abci.RequestQuery{Path: queryPath(skey_1, "/key"), Data: k2, Height: cid.Version}
		qres = store.Query(query2)
		require.True(t, qres.IsOK(), qres.Log)
		require.Equal(t, v2, qres.Value)
		// and for the subspace
		querySub.Height = cid.Version
		qres = store.Query(querySub)
		require.True(t, qres.IsOK(), qres.Log)
		require.Equal(t, valExpSub2, qres.Value)

		// default (height 0) will show latest-1
		qres = store.Query(queryHeight0)
		require.True(t, qres.IsOK(), qres.Log)
		require.Equal(t, v1, qres.Value)
	})

	// querying an empty store will fail
	store2, err := NewStore(memdb.NewDB(), simpleStoreConfig(t))
	require.NoError(t, err)
	qres = store2.Query(queryHeight0)
	require.True(t, qres.IsErr())

	// default shows latest, if latest-1 does not exist
	store2.GetKVStore(skey_1).Set(k1, v1)
	store2.Commit()
	qres = store2.Query(queryHeight0)
	require.True(t, qres.IsOK(), qres.Log)
	require.Equal(t, v1, qres.Value)
	store2.Close()

	t.Run("failed queries", func(t *testing.T) {
		// artificial error cases for coverage (should never happen with prescribed usage)
		// ensure that height overflow triggers an error
		require.NoError(t, err)
		store2.stateDB = dbVersionsIs{store2.stateDB, dbm.NewVersionManager([]uint64{uint64(math.MaxInt64) + 1})}
		qres = store2.Query(queryHeight0)
		require.True(t, qres.IsErr())
		// failure to access versions triggers an error
		store2.stateDB = dbVersionsFails{store.stateDB}
		qres = store2.Query(queryHeight0)
		require.True(t, qres.IsErr())
		store2.Close()

		// query with a nil or empty key fails
		badquery := abci.RequestQuery{Path: queryPath(skey_1, "/key"), Data: []byte{}}
		qres = store.Query(badquery)
		require.True(t, qres.IsErr())
		badquery.Data = nil
		qres = store.Query(badquery)
		require.True(t, qres.IsErr())
		// querying an invalid height will fail
		badquery = abci.RequestQuery{Path: queryPath(skey_1, "/key"), Data: k1, Height: store.LastCommitID().Version + 1}
		qres = store.Query(badquery)
		require.True(t, qres.IsErr())
		// or an invalid path
		badquery = abci.RequestQuery{Path: queryPath(skey_1, "/badpath"), Data: k1}
		qres = store.Query(badquery)
		require.True(t, qres.IsErr())
	})

	t.Run("queries with proof", func(t *testing.T) {
		// test that proofs are generated with single and separate DBs
		testProve := func() {
			queryProve0 := abci.RequestQuery{Path: queryPath(skey_1, "/key"), Data: k1, Prove: true}
			qres = store.Query(queryProve0)
			require.True(t, qres.IsOK(), qres.Log)
			require.Equal(t, v1, qres.Value)
			require.NotNil(t, qres.ProofOps)
		}
		testProve()
		store.Close()

		opts := simpleStoreConfig(t)
		opts.StateCommitmentDB = memdb.NewDB()
		store, err = NewStore(memdb.NewDB(), opts)
		require.NoError(t, err)
		store.GetKVStore(skey_1).Set(k1, v1)
		store.Commit()
		testProve()
		store.Close()
	})
}

func TestStoreConfig(t *testing.T) {
	opts := DefaultStoreConfig()
	// Fail with invalid types
	require.Error(t, opts.RegisterSubstore(skey_1.Name(), types.StoreTypeDB))
	require.Error(t, opts.RegisterSubstore(skey_1.Name(), types.StoreTypeSMT))
	// Ensure that no prefix conflicts are allowed
	require.NoError(t, opts.RegisterSubstore(skey_1.Name(), types.StoreTypePersistent))
	require.NoError(t, opts.RegisterSubstore(skey_2.Name(), types.StoreTypeMemory))
	require.NoError(t, opts.RegisterSubstore(skey_3b.Name(), types.StoreTypeTransient))
	require.Error(t, opts.RegisterSubstore(skey_1b.Name(), types.StoreTypePersistent))
	require.Error(t, opts.RegisterSubstore(skey_2b.Name(), types.StoreTypePersistent))
	require.Error(t, opts.RegisterSubstore(skey_3.Name(), types.StoreTypePersistent))
}

func TestMultiStoreBasic(t *testing.T) {
	opts := DefaultStoreConfig()
	err := opts.RegisterSubstore(skey_1.Name(), types.StoreTypePersistent)
	require.NoError(t, err)
	db := memdb.NewDB()
	store, err := NewStore(db, opts)
	require.NoError(t, err)

	store_1 := store.GetKVStore(skey_1)
	require.NotNil(t, store_1)
	store_1.Set([]byte{0}, []byte{0})
	val := store_1.Get([]byte{0})
	require.Equal(t, []byte{0}, val)
	store_1.Delete([]byte{0})
	val = store_1.Get([]byte{0})
	require.Equal(t, []byte(nil), val)
}

func TestGetVersion(t *testing.T) {
	db := memdb.NewDB()
	opts := storeConfig123(t)
	store, err := NewStore(db, opts)
	require.NoError(t, err)

	cid := store.Commit()
	view, err := store.GetVersion(cid.Version)
	require.NoError(t, err)
	subview := view.GetKVStore(skey_1)
	require.NotNil(t, subview)

	// version view should be read-only
	require.Panics(t, func() { subview.Set([]byte{1}, []byte{1}) })
	require.Panics(t, func() { subview.Delete([]byte{0}) })
	// nonexistent version shouldn't be accessible
	view, err = store.GetVersion(cid.Version + 1)
	require.Equal(t, ErrVersionDoesNotExist, err)

	substore := store.GetKVStore(skey_1)
	require.NotNil(t, substore)
	substore.Set([]byte{0}, []byte{0})
	// setting a value shouldn't affect old version
	require.False(t, subview.Has([]byte{0}))

	cid = store.Commit()
	view, err = store.GetVersion(cid.Version)
	require.NoError(t, err)
	subview = view.GetKVStore(skey_1)
	require.NotNil(t, subview)
	// deleting a value shouldn't affect old version
	substore.Delete([]byte{0})
	require.Equal(t, []byte{0}, subview.Get([]byte{0}))
}

func TestMultiStoreMigration(t *testing.T) {
	db := memdb.NewDB()
	opts := storeConfig123(t)
	store, err := NewStore(db, opts)
	require.NoError(t, err)

	// write some data in all stores
	k1, v1 := []byte("first"), []byte("store")
	s1 := store.GetKVStore(skey_1)
	require.NotNil(t, s1)
	s1.Set(k1, v1)

	k2, v2 := []byte("second"), []byte("restore")
	s2 := store.GetKVStore(skey_2)
	require.NotNil(t, s2)
	s2.Set(k2, v2)

	k3, v3 := []byte("third"), []byte("dropped")
	s3 := store.GetKVStore(skey_3)
	require.NotNil(t, s3)
	s3.Set(k3, v3)

	k4, v4 := []byte("fourth"), []byte("created")
	require.Panics(t, func() { store.GetKVStore(skey_4) })

	cid := store.Commit()
	require.NoError(t, store.Close())
	var migratedID types.CommitID

	// Load without changes and make sure it is sensible
	store, err = NewStore(db, opts)
	require.NoError(t, err)

	// let's query data to see it was saved properly
	s2 = store.GetKVStore(skey_2)
	require.NotNil(t, s2)
	require.Equal(t, v2, s2.Get(k2))
	require.NoError(t, store.Close())

	t.Run("basic migration", func(t *testing.T) {
		// now, let's load with upgrades...
		opts.Upgrades = []types.StoreUpgrades{
			types.StoreUpgrades{
				Added: []string{skey_4.Name()},
				Renamed: []types.StoreRename{{
					OldKey: skey_2.Name(),
					NewKey: skey_2b.Name(),
				}},
				Deleted: []string{skey_3.Name()},
			},
		}
		store, err = NewStore(db, opts)
		require.Nil(t, err)

		// s1 was not changed
		s1 = store.GetKVStore(skey_1)
		require.NotNil(t, s1)
		require.Equal(t, v1, s1.Get(k1))

		// store2 is no longer valid
		require.Panics(t, func() { store.GetKVStore(skey_2) })
		// store2b has the old data
		rs2 := store.GetKVStore(skey_2b)
		require.NotNil(t, rs2)
		require.Equal(t, v2, rs2.Get(k2))

		// store3 is gone
		require.Panics(t, func() { s3 = store.GetKVStore(skey_3) })

		// store4 is valid
		s4 := store.GetKVStore(skey_4)
		require.NotNil(t, s4)
		values := 0
		it := s4.Iterator(nil, nil)
		for ; it.Valid(); it.Next() {
			values += 1
		}
		require.Zero(t, values)
		require.NoError(t, it.Close())
		// write something inside store4
		s4.Set(k4, v4)

		// store this migrated data, and load it again without migrations
		migratedID = store.Commit()
		require.Equal(t, migratedID.Version, int64(2))
		require.NoError(t, store.Close())
	})

	t.Run("reload after migrations", func(t *testing.T) {
		// fail to load the migrated store with the old schema
		store, err = NewStore(db, storeConfig123(t))
		require.Error(t, err)

		// pass in a schema reflecting the migrations
		migratedOpts := DefaultStoreConfig()
		err = migratedOpts.RegisterSubstore(skey_1.Name(), types.StoreTypePersistent)
		require.NoError(t, err)
		err = migratedOpts.RegisterSubstore(skey_2b.Name(), types.StoreTypePersistent)
		require.NoError(t, err)
		err = migratedOpts.RegisterSubstore(skey_4.Name(), types.StoreTypePersistent)
		require.NoError(t, err)
		store, err = NewStore(db, migratedOpts)
		require.Nil(t, err)
		require.Equal(t, migratedID, store.LastCommitID())

		// query this new store
		rl1 := store.GetKVStore(skey_1)
		require.NotNil(t, rl1)
		require.Equal(t, v1, rl1.Get(k1))

		rl2 := store.GetKVStore(skey_2b)
		require.NotNil(t, rl2)
		require.Equal(t, v2, rl2.Get(k2))

		rl4 := store.GetKVStore(skey_4)
		require.NotNil(t, rl4)
		require.Equal(t, v4, rl4.Get(k4))
	})

	t.Run("load view from before migrations", func(t *testing.T) {
		// load and check a view of the store at first commit
		view, err := store.GetVersion(cid.Version)
		require.NoError(t, err)

		s1 = view.GetKVStore(skey_1)
		require.NotNil(t, s1)
		require.Equal(t, v1, s1.Get(k1))

		s2 = view.GetKVStore(skey_2)
		require.NotNil(t, s2)
		require.Equal(t, v2, s2.Get(k2))

		s3 = view.GetKVStore(skey_3)
		require.NotNil(t, s3)
		require.Equal(t, v3, s3.Get(k3))

		require.Panics(t, func() {
			view.GetKVStore(skey_4)
		})
	})
}

func TestTrace(t *testing.T) {
	key, value := []byte("test-key"), []byte("test-value")
	tctx := types.TraceContext(map[string]interface{}{"blockHeight": 64})

	expected_Set := "{\"operation\":\"write\",\"key\":\"dGVzdC1rZXk=\",\"value\":\"dGVzdC12YWx1ZQ==\",\"metadata\":{\"blockHeight\":64}}\n"
	expected_Get := "{\"operation\":\"read\",\"key\":\"dGVzdC1rZXk=\",\"value\":\"dGVzdC12YWx1ZQ==\",\"metadata\":{\"blockHeight\":64}}\n"
	expected_Get_missing := "{\"operation\":\"read\",\"key\":\"dGVzdC1rZXk=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n"
	expected_Delete := "{\"operation\":\"delete\",\"key\":\"dGVzdC1rZXk=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n"
	expected_IterKey := "{\"operation\":\"iterKey\",\"key\":\"dGVzdC1rZXk=\",\"value\":\"\",\"metadata\":{\"blockHeight\":64}}\n"
	expected_IterValue := "{\"operation\":\"iterValue\",\"key\":\"\",\"value\":\"dGVzdC12YWx1ZQ==\",\"metadata\":{\"blockHeight\":64}}\n"

	db := memdb.NewDB()
	opts := simpleStoreConfig(t)
	require.NoError(t, opts.RegisterSubstore(skey_2.Name(), types.StoreTypeMemory))
	require.NoError(t, opts.RegisterSubstore(skey_3.Name(), types.StoreTypeTransient))

	store, err := NewStore(db, opts)
	require.NoError(t, err)
	store.SetTraceContext(tctx)
	require.False(t, store.TracingEnabled())

	var buf bytes.Buffer
	store.SetTracer(&buf)
	require.True(t, store.TracingEnabled())

	for _, skey := range []types.StoreKey{skey_1, skey_2, skey_3} {
		buf.Reset()
		store.GetKVStore(skey).Get(key)
		require.Equal(t, expected_Get_missing, buf.String())

		buf.Reset()
		store.GetKVStore(skey).Set(key, value)
		require.Equal(t, expected_Set, buf.String())

		buf.Reset()
		require.Equal(t, value, store.GetKVStore(skey).Get(key))
		require.Equal(t, expected_Get, buf.String())

		iter := store.GetKVStore(skey).Iterator(nil, nil)
		buf.Reset()
		require.Equal(t, key, iter.Key())
		require.Equal(t, expected_IterKey, buf.String())
		buf.Reset()
		require.Equal(t, value, iter.Value())
		require.Equal(t, expected_IterValue, buf.String())
		require.NoError(t, iter.Close())

		buf.Reset()
		store.GetKVStore(skey).Delete(key)
		require.Equal(t, expected_Delete, buf.String())

	}
	store.SetTracer(nil)
	require.False(t, store.TracingEnabled())
	require.NoError(t, store.Close())
}

func TestListeners(t *testing.T) {
	kvPairs := []types.KVPair{
		{Key: []byte{1}, Value: []byte("v1")},
		{Key: []byte{2}, Value: []byte("v2")},
		{Key: []byte{3}, Value: []byte("v3")},
	}

	testCases := []struct {
		key   []byte
		value []byte
		skey  types.StoreKey
	}{
		{
			key:   kvPairs[0].Key,
			value: kvPairs[0].Value,
			skey:  skey_1,
		},
		{
			key:   kvPairs[1].Key,
			value: kvPairs[1].Value,
			skey:  skey_2,
		},
		{
			key:   kvPairs[2].Key,
			value: kvPairs[2].Value,
			skey:  skey_3,
		},
	}

	var interfaceRegistry = codecTypes.NewInterfaceRegistry()
	var marshaller = codec.NewProtoCodec(interfaceRegistry)

	db := memdb.NewDB()
	opts := simpleStoreConfig(t)
	require.NoError(t, opts.RegisterSubstore(skey_2.Name(), types.StoreTypeMemory))
	require.NoError(t, opts.RegisterSubstore(skey_3.Name(), types.StoreTypeTransient))

	store, err := NewStore(db, opts)
	require.NoError(t, err)

	for i, tc := range testCases {
		var buf bytes.Buffer
		listener := types.NewStoreKVPairWriteListener(&buf, marshaller)
		store.AddListeners(tc.skey, []types.WriteListener{listener})
		require.True(t, store.ListeningEnabled(tc.skey))

		// Set case
		expected := types.StoreKVPair{
			Key:      tc.key,
			Value:    tc.value,
			StoreKey: tc.skey.Name(),
			Delete:   false,
		}
		var kvpair types.StoreKVPair

		buf.Reset()
		store.GetKVStore(tc.skey).Set(tc.key, tc.value)
		require.NoError(t, marshaller.UnmarshalLengthPrefixed(buf.Bytes(), &kvpair))
		require.Equal(t, expected, kvpair, i)

		// Delete case
		expected = types.StoreKVPair{
			Key:      tc.key,
			Value:    nil,
			StoreKey: tc.skey.Name(),
			Delete:   true,
		}
		kvpair = types.StoreKVPair{}

		buf.Reset()
		store.GetKVStore(tc.skey).Delete(tc.key)
		require.NoError(t, marshaller.UnmarshalLengthPrefixed(buf.Bytes(), &kvpair))
		require.Equal(t, expected, kvpair, i)
	}
	require.NoError(t, store.Close())
}
