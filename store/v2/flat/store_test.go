package flat

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/store/types"
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

	require.Panics(t, func() { store.Get(nil) }, "Get(nil key) should panic")
	require.Panics(t, func() { store.Get([]byte{}) }, "Get(empty key) should panic")
	require.Panics(t, func() { store.Has(nil) }, "Has(nil key) should panic")
	require.Panics(t, func() { store.Has([]byte{}) }, "Has(empty key) should panic")
	require.Panics(t, func() { store.Set(nil, []byte("value")) }, "Set(nil key) should panic")
	require.Panics(t, func() { store.Set([]byte{}, []byte("value")) }, "Set(empty key) should panic")
	require.Panics(t, func() { store.Set([]byte("key"), nil) }, "Set(nil value) should panic")
	store.indexTxn = rwCrudFails{store.indexTxn}
	require.Panics(t, func() { store.Set([]byte("key"), []byte("value")) },
		"Set() when index fails should panic")
}

func TestConstructors(t *testing.T) {
	db := memdb.NewDB()

	store := newAlohaStore(t, db)
	store.Commit()
	require.NoError(t, store.Close())

	store, err := NewStore(db, DefaultStoreConfig)
	require.NoError(t, err)
	value := store.Get([]byte("hello"))
	require.Equal(t, []byte("goodbye"), value)
	require.NoError(t, store.Close())

	// Loading with an initial version beyond the lowest should error
	opts := StoreConfig{InitialVersion: 5, Pruning: types.PruneNothing}
	store, err = NewStore(db, opts)
	require.Error(t, err)
	db.Close()

	store, err = NewStore(dbVersionsFails{memdb.NewDB()}, DefaultStoreConfig)
	require.Error(t, err)
	store, err = NewStore(db, StoreConfig{MerkleDB: dbVersionsFails{memdb.NewDB()}})
	require.Error(t, err)

	// can't use a DB with open writers
	db = memdb.NewDB()
	merkledb := memdb.NewDB()
	w := db.Writer()
	store, err = NewStore(db, DefaultStoreConfig)
	require.Error(t, err)
	w.Discard()
	w = merkledb.Writer()
	store, err = NewStore(db, StoreConfig{MerkleDB: merkledb})
	require.Error(t, err)
	w.Discard()

	// can't use DBs with different version history
	merkledb.SaveNextVersion()
	store, err = NewStore(db, StoreConfig{MerkleDB: merkledb})
	require.Error(t, err)
	merkledb.Close()

	// can't load existing store when we can't access the latest Merkle root hash
	store, err = NewStore(db, DefaultStoreConfig)
	require.NoError(t, err)
	store.Commit()
	require.NoError(t, store.Close())
	// because root is misssing
	w = db.Writer()
	w.Delete(merkleRootKey)
	w.Commit()
	db.SaveNextVersion()
	store, err = NewStore(db, DefaultStoreConfig)
	require.Error(t, err)
	// or, because of an error
	store, err = NewStore(dbRWCrudFails{db}, DefaultStoreConfig)
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
		// Sanity test for Merkle hashing
		store, err := NewStore(memdb.NewDB(), opts)
		require.NoError(t, err)
		require.Zero(t, store.LastCommitID())
		idNew := store.Commit()
		store.Set([]byte{0}, []byte{0})
		idOne := store.Commit()
		require.Equal(t, idNew.Version+1, idOne.Version)
		require.NotEqual(t, idNew.Hash, idOne.Hash)

		// Hash of emptied store is same as new store
		store.Delete([]byte{0})
		idEmptied := store.Commit()
		require.Equal(t, idNew.Hash, idEmptied.Hash)

		previd := idEmptied
		for i := byte(1); i < 5; i++ {
			store.Set([]byte{i}, []byte{i})
			id := store.Commit()
			lastid := store.LastCommitID()
			require.Equal(t, id.Hash, lastid.Hash)
			require.Equal(t, id.Version, lastid.Version)
			require.NotEqual(t, previd.Hash, id.Hash)
			require.NotEqual(t, previd.Version, id.Version)
		}
	}
	testBasic(StoreConfig{Pruning: types.PruneNothing})
	testBasic(StoreConfig{Pruning: types.PruneNothing, MerkleDB: memdb.NewDB()})

	testFailedCommit := func(t *testing.T, store *Store, db dbm.DBConnection) {
		opts := store.opts
		if db == nil {
			db = store.stateDB
		}

		store.Set([]byte{0}, []byte{0})
		require.Panics(t, func() { store.Commit() })
		require.NoError(t, store.Close())

		versions, _ := db.Versions()
		require.Equal(t, 0, versions.Count())
		if opts.MerkleDB != nil {
			versions, _ = opts.MerkleDB.Versions()
			require.Equal(t, 0, versions.Count())
		}

		store, err := NewStore(db, opts)
		require.NoError(t, err)
		require.Nil(t, store.Get([]byte{0}))
		require.NoError(t, store.Close())
	}

	// Ensure storage commit is rolled back in each failure case
	t.Run("recover after failed Commit", func(t *testing.T) {
		store, err := NewStore(
			dbRWCommitFails{memdb.NewDB()},
			StoreConfig{Pruning: types.PruneNothing})
		require.NoError(t, err)
		testFailedCommit(t, store, nil)
	})
	t.Run("recover after failed SaveVersion", func(t *testing.T) {
		store, err := NewStore(
			dbSaveVersionFails{memdb.NewDB()},
			StoreConfig{Pruning: types.PruneNothing})
		require.NoError(t, err)
		testFailedCommit(t, store, nil)
	})
	t.Run("recover after failed MerkleDB Commit", func(t *testing.T) {
		store, err := NewStore(memdb.NewDB(),
			StoreConfig{MerkleDB: dbRWCommitFails{memdb.NewDB()}, Pruning: types.PruneNothing})
		require.NoError(t, err)
		testFailedCommit(t, store, nil)
	})
	t.Run("recover after failed MerkleDB SaveVersion", func(t *testing.T) {
		store, err := NewStore(memdb.NewDB(),
			StoreConfig{MerkleDB: dbSaveVersionFails{memdb.NewDB()}, Pruning: types.PruneNothing})
		require.NoError(t, err)
		testFailedCommit(t, store, nil)
	})

	t.Run("recover after stateDB.Versions error triggers failure", func(t *testing.T) {
		db := memdb.NewDB()
		store, err := NewStore(db, DefaultStoreConfig)
		require.NoError(t, err)
		store.stateDB = dbVersionsFails{store.stateDB}
		testFailedCommit(t, store, db)
	})
	t.Run("recover after stateTxn.Set error triggers failure", func(t *testing.T) {
		store, err := NewStore(memdb.NewDB(), DefaultStoreConfig)
		require.NoError(t, err)
		store.stateTxn = rwCrudFails{store.stateTxn}
		testFailedCommit(t, store, nil)
	})

	t.Run("stateDB.DeleteVersion error triggers failure", func(t *testing.T) {
		store, err := NewStore(memdb.NewDB(), StoreConfig{MerkleDB: memdb.NewDB()})
		require.NoError(t, err)
		store.merkleTxn = rwCommitFails{store.merkleTxn}
		store.stateDB = dbDeleteVersionFails{store.stateDB}
		require.Panics(t, func() { store.Commit() })
	})
	t.Run("height overflow triggers failure", func(t *testing.T) {
		store, err := NewStore(memdb.NewDB(),
			StoreConfig{InitialVersion: math.MaxInt64, Pruning: types.PruneNothing})
		require.NoError(t, err)
		require.Equal(t, int64(math.MaxInt64), store.Commit().Version)
		require.Panics(t, func() { store.Commit() })
		require.Equal(t, int64(math.MaxInt64), store.LastCommitID().Version) // version history not modified
	})

	// setting initial version
	store, err := NewStore(memdb.NewDB(),
		StoreConfig{InitialVersion: 5, Pruning: types.PruneNothing, MerkleDB: memdb.NewDB()})
	require.NoError(t, err)
	require.Equal(t, int64(5), store.Commit().Version)

	store, err = NewStore(memdb.NewDB(), StoreConfig{MerkleDB: memdb.NewDB()})
	require.NoError(t, err)
	store.Commit()
	store.stateDB = dbVersionsFails{store.stateDB}
	require.Panics(t, func() { store.LastCommitID() })

	store, err = NewStore(memdb.NewDB(), StoreConfig{MerkleDB: memdb.NewDB()})
	require.NoError(t, err)
	store.Commit()
	store.stateTxn = rwCrudFails{store.stateTxn}
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
		store, err := NewStore(dbs[0], StoreConfig{Pruning: tc.PruningOptions, MerkleDB: dbs[1]})
		require.NoError(t, err)

		for i := byte(1); i <= 10; i++ {
			store.Set([]byte{i}, []byte{i})
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
	require.True(t, qres.IsOK())
	require.Equal(t, valExpSubEmpty, qres.Value)

	// set data
	store.Set(k1, v1)
	store.Set(k2, v2)

	// set data without commit, doesn't show up
	qres = store.Query(query)
	require.True(t, qres.IsOK())
	require.Nil(t, qres.Value)

	// commit it, but still don't see on old version
	cid = store.Commit()
	qres = store.Query(query)
	require.True(t, qres.IsOK())
	require.Nil(t, qres.Value)

	// but yes on the new version
	query.Height = cid.Version
	qres = store.Query(query)
	require.True(t, qres.IsOK())
	require.Equal(t, v1, qres.Value)
	// and for the subspace
	qres = store.Query(querySub)
	require.True(t, qres.IsOK())
	require.Equal(t, valExpSub1, qres.Value)

	// modify
	store.Set(k1, v3)
	cid = store.Commit()

	// query will return old values, as height is fixed
	qres = store.Query(query)
	require.True(t, qres.IsOK())
	require.Equal(t, v1, qres.Value)

	// update to latest in the query and we are happy
	query.Height = cid.Version
	qres = store.Query(query)
	require.True(t, qres.IsOK())
	require.Equal(t, v3, qres.Value)

	query2 := abci.RequestQuery{Path: "/key", Data: k2, Height: cid.Version}
	qres = store.Query(query2)
	require.True(t, qres.IsOK())
	require.Equal(t, v2, qres.Value)
	// and for the subspace
	qres = store.Query(querySub)
	require.True(t, qres.IsOK())
	require.Equal(t, valExpSub2, qres.Value)

	// default (height 0) will show latest -1
	query0 := abci.RequestQuery{Path: "/key", Data: k1}
	qres = store.Query(query0)
	require.True(t, qres.IsOK())
	require.Equal(t, v1, qres.Value)

	// querying an empty store will fail
	store2, err := NewStore(memdb.NewDB(), DefaultStoreConfig)
	require.NoError(t, err)
	qres = store2.Query(query0)
	require.True(t, qres.IsErr())

	// default shows latest, if latest-1 does not exist
	store2.Set(k1, v1)
	store2.Commit()
	qres = store2.Query(query0)
	require.True(t, qres.IsOK())
	require.Equal(t, v1, qres.Value)
	store2.Close()

	// artificial error cases for coverage (should never happen with defined usage)
	// ensure that height overflow triggers an error
	require.NoError(t, err)
	store2.stateDB = dbVersionsIs{store2.stateDB, dbm.NewVersionManager([]uint64{uint64(math.MaxInt64) + 1})}
	qres = store2.Query(query0)
	require.True(t, qres.IsErr())
	// failure to access versions triggers an error
	store2.stateDB = dbVersionsFails{store.stateDB}
	qres = store2.Query(query0)
	require.True(t, qres.IsErr())
	store2.Close()

	// query with a nil or empty key fails
	badquery := abci.RequestQuery{Path: "/key", Data: []byte{}}
	qres = store.Query(badquery)
	require.True(t, qres.IsErr())
	badquery.Data = nil
	qres = store.Query(badquery)
	require.True(t, qres.IsErr())
	// querying an invalid height will fail
	badquery = abci.RequestQuery{Path: "/key", Data: k1, Height: store.LastCommitID().Version + 1}
	qres = store.Query(badquery)
	require.True(t, qres.IsErr())
	// or an invalid path
	badquery = abci.RequestQuery{Path: "/badpath", Data: k1}
	qres = store.Query(badquery)
	require.True(t, qres.IsErr())

	// test that proofs are generated with single and separate DBs
	testProve := func() {
		queryProve0 := abci.RequestQuery{Path: "/key", Data: k1, Prove: true}
		store.Query(queryProve0)
		qres = store.Query(queryProve0)
		require.True(t, qres.IsOK())
		require.Equal(t, v1, qres.Value)
		require.NotNil(t, qres.ProofOps)
	}
	testProve()
	store.Close()

	store, err = NewStore(memdb.NewDB(), StoreConfig{MerkleDB: memdb.NewDB()})
	require.NoError(t, err)
	store.Set(k1, v1)
	store.Commit()
	testProve()
	store.Close()
}

type dbDeleteVersionFails struct{ dbm.DBConnection }
type dbRWCommitFails struct{ *memdb.MemDB }
type dbRWCrudFails struct{ dbm.DBConnection }
type dbSaveVersionFails struct{ *memdb.MemDB }
type dbVersionsIs struct {
	dbm.DBConnection
	vset dbm.VersionSet
}
type dbVersionsFails struct{ dbm.DBConnection }
type rwCommitFails struct{ dbm.DBReadWriter }
type rwCrudFails struct{ dbm.DBReadWriter }

func (dbVersionsFails) Versions() (dbm.VersionSet, error) { return nil, errors.New("dbVersionsFails") }
func (db dbVersionsIs) Versions() (dbm.VersionSet, error) { return db.vset, nil }
func (db dbRWCrudFails) ReadWriter() dbm.DBReadWriter {
	return rwCrudFails{db.DBConnection.ReadWriter()}
}
func (dbSaveVersionFails) SaveVersion(uint64) error     { return errors.New("dbSaveVersionFails") }
func (dbDeleteVersionFails) DeleteVersion(uint64) error { return errors.New("dbDeleteVersionFails") }
func (tx rwCommitFails) Commit() error {
	tx.Discard()
	return errors.New("rwCommitFails")
}
func (db dbRWCommitFails) ReadWriter() dbm.DBReadWriter {
	return rwCommitFails{db.MemDB.ReadWriter()}
}

func (rwCrudFails) Get([]byte) ([]byte, error) { return nil, errors.New("rwCrudFails.Get") }
func (rwCrudFails) Has([]byte) (bool, error)   { return false, errors.New("rwCrudFails.Has") }
func (rwCrudFails) Set([]byte, []byte) error   { return errors.New("rwCrudFails.Set") }
func (rwCrudFails) Delete([]byte) error        { return errors.New("rwCrudFails.Delete") }
