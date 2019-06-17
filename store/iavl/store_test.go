package iavl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/errors"
	"github.com/cosmos/cosmos-sdk/store/types"
)

var (
	cacheSize        = 100
	numRecent  int64 = 5
	storeEvery int64 = 3
)

var (
	treeData = map[string]string{
		"hello": "goodbye",
		"aloha": "shalom",
	}
	nMoreData = 0
)

// make a tree with data from above and save it
func newAlohaTree(t *testing.T, db dbm.DB) (*iavl.MutableTree, types.CommitID) {
	tree := iavl.NewMutableTree(db, cacheSize)
	for k, v := range treeData {
		tree.Set([]byte(k), []byte(v))
	}
	for i := 0; i < nMoreData; i++ {
		key := cmn.RandBytes(12)
		value := cmn.RandBytes(50)
		tree.Set(key, value)
	}
	hash, ver, err := tree.SaveVersion()
	require.Nil(t, err)
	return tree, types.CommitID{ver, hash}
}

func TestGetImmutable(t *testing.T) {
	db := dbm.NewMemDB()
	tree, cID := newAlohaTree(t, db)
	store := UnsafeNewStore(tree, 10, 10)

	require.True(t, tree.Set([]byte("hello"), []byte("adios")))
	hash, ver, err := tree.SaveVersion()
	cID = types.CommitID{ver, hash}
	require.Nil(t, err)

	_, err = store.GetImmutable(cID.Version + 1)
	require.Error(t, err)

	newStore, err := store.GetImmutable(cID.Version - 1)
	require.NoError(t, err)
	require.Equal(t, newStore.Get([]byte("hello")), []byte("goodbye"))

	newStore, err = store.GetImmutable(cID.Version)
	require.NoError(t, err)
	require.Equal(t, newStore.Get([]byte("hello")), []byte("adios"))

	res := newStore.Query(abci.RequestQuery{Data: []byte("hello"), Height: cID.Version, Path: "/key", Prove: true})
	require.Equal(t, res.Value, []byte("adios"))
	require.NotNil(t, res.Proof)

	require.Panics(t, func() { newStore.Set(nil, nil) })
	require.Panics(t, func() { newStore.Delete(nil) })
	require.Panics(t, func() { newStore.Commit() })
}

func TestTestGetImmutableIterator(t *testing.T) {
	db := dbm.NewMemDB()
	tree, cID := newAlohaTree(t, db)
	store := UnsafeNewStore(tree, 10, 10)

	newStore, err := store.GetImmutable(cID.Version)
	require.NoError(t, err)

	iter := newStore.Iterator([]byte("aloha"), []byte("hellz"))
	expected := []string{"aloha", "hello"}
	var i int

	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, treeData[expectedKey])
		i++
	}

	require.Equal(t, len(expected), i)
}

func TestIAVLStoreGetSetHasDelete(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newAlohaTree(t, db)
	iavlStore := UnsafeNewStore(tree, numRecent, storeEvery)

	key := "hello"

	exists := iavlStore.Has([]byte(key))
	require.True(t, exists)

	value := iavlStore.Get([]byte(key))
	require.EqualValues(t, value, treeData[key])

	value2 := "notgoodbye"
	iavlStore.Set([]byte(key), []byte(value2))

	value = iavlStore.Get([]byte(key))
	require.EqualValues(t, value, value2)

	iavlStore.Delete([]byte(key))

	exists = iavlStore.Has([]byte(key))
	require.False(t, exists)
}

func TestIAVLStoreNoNilSet(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newAlohaTree(t, db)
	iavlStore := UnsafeNewStore(tree, numRecent, storeEvery)
	require.Panics(t, func() { iavlStore.Set([]byte("key"), nil) }, "setting a nil value should panic")
}

func TestIAVLIterator(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newAlohaTree(t, db)
	iavlStore := UnsafeNewStore(tree, numRecent, storeEvery)
	iter := iavlStore.Iterator([]byte("aloha"), []byte("hellz"))
	expected := []string{"aloha", "hello"}
	var i int

	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	require.Equal(t, len(expected), i)

	iter = iavlStore.Iterator([]byte("golang"), []byte("rocks"))
	expected = []string{"hello"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	require.Equal(t, len(expected), i)

	iter = iavlStore.Iterator(nil, []byte("golang"))
	expected = []string{"aloha"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	require.Equal(t, len(expected), i)

	iter = iavlStore.Iterator(nil, []byte("shalom"))
	expected = []string{"aloha", "hello"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	require.Equal(t, len(expected), i)

	iter = iavlStore.Iterator(nil, nil)
	expected = []string{"aloha", "hello"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	require.Equal(t, len(expected), i)

	iter = iavlStore.Iterator([]byte("golang"), nil)
	expected = []string{"hello"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	require.Equal(t, len(expected), i)
}

func TestIAVLReverseIterator(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewMutableTree(db, cacheSize)
	iavlStore := UnsafeNewStore(tree, numRecent, storeEvery)

	iavlStore.Set([]byte{0x00}, []byte("0"))
	iavlStore.Set([]byte{0x00, 0x00}, []byte("0 0"))
	iavlStore.Set([]byte{0x00, 0x01}, []byte("0 1"))
	iavlStore.Set([]byte{0x00, 0x02}, []byte("0 2"))
	iavlStore.Set([]byte{0x01}, []byte("1"))

	var testReverseIterator = func(t *testing.T, start []byte, end []byte, expected []string) {
		iter := iavlStore.ReverseIterator(start, end)
		var i int
		for i = 0; iter.Valid(); iter.Next() {
			expectedValue := expected[i]
			value := iter.Value()
			require.EqualValues(t, string(value), expectedValue)
			i++
		}
		require.Equal(t, len(expected), i)
	}

	testReverseIterator(t, nil, nil, []string{"1", "0 2", "0 1", "0 0", "0"})
	testReverseIterator(t, []byte{0x00}, nil, []string{"1", "0 2", "0 1", "0 0", "0"})
	testReverseIterator(t, []byte{0x00}, []byte{0x00, 0x01}, []string{"0 0", "0"})
	testReverseIterator(t, []byte{0x00}, []byte{0x01}, []string{"0 2", "0 1", "0 0", "0"})
	testReverseIterator(t, []byte{0x00, 0x01}, []byte{0x01}, []string{"0 2", "0 1"})
	testReverseIterator(t, nil, []byte{0x01}, []string{"0 2", "0 1", "0 0", "0"})
}

func TestIAVLPrefixIterator(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewMutableTree(db, cacheSize)
	iavlStore := UnsafeNewStore(tree, numRecent, storeEvery)

	iavlStore.Set([]byte("test1"), []byte("test1"))
	iavlStore.Set([]byte("test2"), []byte("test2"))
	iavlStore.Set([]byte("test3"), []byte("test3"))
	iavlStore.Set([]byte{byte(55), byte(255), byte(255), byte(0)}, []byte("test4"))
	iavlStore.Set([]byte{byte(55), byte(255), byte(255), byte(1)}, []byte("test4"))
	iavlStore.Set([]byte{byte(55), byte(255), byte(255), byte(255)}, []byte("test4"))
	iavlStore.Set([]byte{byte(255), byte(255), byte(0)}, []byte("test4"))
	iavlStore.Set([]byte{byte(255), byte(255), byte(1)}, []byte("test4"))
	iavlStore.Set([]byte{byte(255), byte(255), byte(255)}, []byte("test4"))

	var i int

	iter := types.KVStorePrefixIterator(iavlStore, []byte("test"))
	expected := []string{"test1", "test2", "test3"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, expectedKey)
		i++
	}
	iter.Close()
	require.Equal(t, len(expected), i)

	iter = types.KVStorePrefixIterator(iavlStore, []byte{byte(55), byte(255), byte(255)})
	expected2 := [][]byte{
		{byte(55), byte(255), byte(255), byte(0)},
		{byte(55), byte(255), byte(255), byte(1)},
		{byte(55), byte(255), byte(255), byte(255)},
	}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected2[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, []byte("test4"))
		i++
	}
	iter.Close()
	require.Equal(t, len(expected), i)

	iter = types.KVStorePrefixIterator(iavlStore, []byte{byte(255), byte(255)})
	expected2 = [][]byte{
		{byte(255), byte(255), byte(0)},
		{byte(255), byte(255), byte(1)},
		{byte(255), byte(255), byte(255)},
	}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected2[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, []byte("test4"))
		i++
	}
	iter.Close()
	require.Equal(t, len(expected), i)
}

func TestIAVLReversePrefixIterator(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewMutableTree(db, cacheSize)
	iavlStore := UnsafeNewStore(tree, numRecent, storeEvery)

	iavlStore.Set([]byte("test1"), []byte("test1"))
	iavlStore.Set([]byte("test2"), []byte("test2"))
	iavlStore.Set([]byte("test3"), []byte("test3"))
	iavlStore.Set([]byte{byte(55), byte(255), byte(255), byte(0)}, []byte("test4"))
	iavlStore.Set([]byte{byte(55), byte(255), byte(255), byte(1)}, []byte("test4"))
	iavlStore.Set([]byte{byte(55), byte(255), byte(255), byte(255)}, []byte("test4"))
	iavlStore.Set([]byte{byte(255), byte(255), byte(0)}, []byte("test4"))
	iavlStore.Set([]byte{byte(255), byte(255), byte(1)}, []byte("test4"))
	iavlStore.Set([]byte{byte(255), byte(255), byte(255)}, []byte("test4"))

	var i int

	iter := types.KVStoreReversePrefixIterator(iavlStore, []byte("test"))
	expected := []string{"test3", "test2", "test1"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, expectedKey)
		i++
	}
	require.Equal(t, len(expected), i)

	iter = types.KVStoreReversePrefixIterator(iavlStore, []byte{byte(55), byte(255), byte(255)})
	expected2 := [][]byte{
		{byte(55), byte(255), byte(255), byte(255)},
		{byte(55), byte(255), byte(255), byte(1)},
		{byte(55), byte(255), byte(255), byte(0)},
	}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected2[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, []byte("test4"))
		i++
	}
	require.Equal(t, len(expected), i)

	iter = types.KVStoreReversePrefixIterator(iavlStore, []byte{byte(255), byte(255)})
	expected2 = [][]byte{
		{byte(255), byte(255), byte(255)},
		{byte(255), byte(255), byte(1)},
		{byte(255), byte(255), byte(0)},
	}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected2[i]
		key, value := iter.Key(), iter.Value()
		require.EqualValues(t, key, expectedKey)
		require.EqualValues(t, value, []byte("test4"))
		i++
	}
	require.Equal(t, len(expected), i)
}

func nextVersion(iavl *Store) {
	key := []byte(fmt.Sprintf("Key for tree: %d", iavl.LastCommitID().Version))
	value := []byte(fmt.Sprintf("Value for tree: %d", iavl.LastCommitID().Version))
	iavl.Set(key, value)
	iavl.Commit()
}

func TestIAVLDefaultPruning(t *testing.T) {
	//Expected stored / deleted version numbers for:
	//numRecent = 5, storeEvery = 3
	var states = []pruneState{
		{[]int64{}, []int64{}},
		{[]int64{1}, []int64{}},
		{[]int64{1, 2}, []int64{}},
		{[]int64{1, 2, 3}, []int64{}},
		{[]int64{1, 2, 3, 4}, []int64{}},
		{[]int64{1, 2, 3, 4, 5}, []int64{}},
		{[]int64{1, 2, 3, 4, 5, 6}, []int64{}},
		{[]int64{2, 3, 4, 5, 6, 7}, []int64{1}},
		{[]int64{3, 4, 5, 6, 7, 8}, []int64{1, 2}},
		{[]int64{3, 4, 5, 6, 7, 8, 9}, []int64{1, 2}},
		{[]int64{3, 5, 6, 7, 8, 9, 10}, []int64{1, 2, 4}},
		{[]int64{3, 6, 7, 8, 9, 10, 11}, []int64{1, 2, 4, 5}},
		{[]int64{3, 6, 7, 8, 9, 10, 11, 12}, []int64{1, 2, 4, 5}},
		{[]int64{3, 6, 8, 9, 10, 11, 12, 13}, []int64{1, 2, 4, 5, 7}},
		{[]int64{3, 6, 9, 10, 11, 12, 13, 14}, []int64{1, 2, 4, 5, 7, 8}},
		{[]int64{3, 6, 9, 10, 11, 12, 13, 14, 15}, []int64{1, 2, 4, 5, 7, 8}},
	}
	testPruning(t, int64(5), int64(3), states)
}

func TestIAVLAlternativePruning(t *testing.T) {
	//Expected stored / deleted version numbers for:
	//numRecent = 3, storeEvery = 5
	var states = []pruneState{
		{[]int64{}, []int64{}},
		{[]int64{1}, []int64{}},
		{[]int64{1, 2}, []int64{}},
		{[]int64{1, 2, 3}, []int64{}},
		{[]int64{1, 2, 3, 4}, []int64{}},
		{[]int64{2, 3, 4, 5}, []int64{1}},
		{[]int64{3, 4, 5, 6}, []int64{1, 2}},
		{[]int64{4, 5, 6, 7}, []int64{1, 2, 3}},
		{[]int64{5, 6, 7, 8}, []int64{1, 2, 3, 4}},
		{[]int64{5, 6, 7, 8, 9}, []int64{1, 2, 3, 4}},
		{[]int64{5, 7, 8, 9, 10}, []int64{1, 2, 3, 4, 6}},
		{[]int64{5, 8, 9, 10, 11}, []int64{1, 2, 3, 4, 6, 7}},
		{[]int64{5, 9, 10, 11, 12}, []int64{1, 2, 3, 4, 6, 7, 8}},
		{[]int64{5, 10, 11, 12, 13}, []int64{1, 2, 3, 4, 6, 7, 8, 9}},
		{[]int64{5, 10, 11, 12, 13, 14}, []int64{1, 2, 3, 4, 6, 7, 8, 9}},
		{[]int64{5, 10, 12, 13, 14, 15}, []int64{1, 2, 3, 4, 6, 7, 8, 9, 11}},
	}
	testPruning(t, int64(3), int64(5), states)
}

type pruneState struct {
	stored  []int64
	deleted []int64
}

func testPruning(t *testing.T, numRecent int64, storeEvery int64, states []pruneState) {
	db := dbm.NewMemDB()
	tree := iavl.NewMutableTree(db, cacheSize)
	iavlStore := UnsafeNewStore(tree, numRecent, storeEvery)
	for step, state := range states {
		for _, ver := range state.stored {
			require.True(t, iavlStore.VersionExists(ver),
				"Missing version %d with latest version %d. Should save last %d and every %d",
				ver, step, numRecent, storeEvery)
		}
		for _, ver := range state.deleted {
			require.False(t, iavlStore.VersionExists(ver),
				"Unpruned version %d with latest version %d. Should prune all but last %d and every %d",
				ver, step, numRecent, storeEvery)
		}
		nextVersion(iavlStore)
	}
}

func TestIAVLNoPrune(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewMutableTree(db, cacheSize)
	iavlStore := UnsafeNewStore(tree, numRecent, int64(1))
	nextVersion(iavlStore)
	for i := 1; i < 100; i++ {
		for j := 1; j <= i; j++ {
			require.True(t, iavlStore.VersionExists(int64(j)),
				"Missing version %d with latest version %d. Should be storing all versions",
				j, i)
		}
		nextVersion(iavlStore)
	}
}

func TestIAVLPruneEverything(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewMutableTree(db, cacheSize)
	iavlStore := UnsafeNewStore(tree, int64(0), int64(0))
	nextVersion(iavlStore)
	for i := 1; i < 100; i++ {
		for j := 1; j < i; j++ {
			require.False(t, iavlStore.VersionExists(int64(j)),
				"Unpruned version %d with latest version %d. Should prune all old versions",
				j, i)
		}
		require.True(t, iavlStore.VersionExists(int64(i)),
			"Missing current version on step %d, should not prune current state tree",
			i)
		nextVersion(iavlStore)
	}
}

func TestIAVLStoreQuery(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewMutableTree(db, cacheSize)
	iavlStore := UnsafeNewStore(tree, numRecent, storeEvery)

	k1, v1 := []byte("key1"), []byte("val1")
	k2, v2 := []byte("key2"), []byte("val2")
	v3 := []byte("val3")

	ksub := []byte("key")
	KVs0 := []types.KVPair{}
	KVs1 := []types.KVPair{
		{Key: k1, Value: v1},
		{Key: k2, Value: v2},
	}
	KVs2 := []types.KVPair{
		{Key: k1, Value: v3},
		{Key: k2, Value: v2},
	}
	valExpSubEmpty := cdc.MustMarshalBinaryLengthPrefixed(KVs0)
	valExpSub1 := cdc.MustMarshalBinaryLengthPrefixed(KVs1)
	valExpSub2 := cdc.MustMarshalBinaryLengthPrefixed(KVs2)

	cid := iavlStore.Commit()
	ver := cid.Version
	query := abci.RequestQuery{Path: "/key", Data: k1, Height: ver}
	querySub := abci.RequestQuery{Path: "/subspace", Data: ksub, Height: ver}

	// query subspace before anything set
	qres := iavlStore.Query(querySub)
	require.Equal(t, uint32(errors.CodeOK), qres.Code)
	require.Equal(t, valExpSubEmpty, qres.Value)

	// set data
	iavlStore.Set(k1, v1)
	iavlStore.Set(k2, v2)

	// set data without commit, doesn't show up
	qres = iavlStore.Query(query)
	require.Equal(t, uint32(errors.CodeOK), qres.Code)
	require.Nil(t, qres.Value)

	// commit it, but still don't see on old version
	cid = iavlStore.Commit()
	qres = iavlStore.Query(query)
	require.Equal(t, uint32(errors.CodeOK), qres.Code)
	require.Nil(t, qres.Value)

	// but yes on the new version
	query.Height = cid.Version
	qres = iavlStore.Query(query)
	require.Equal(t, uint32(errors.CodeOK), qres.Code)
	require.Equal(t, v1, qres.Value)

	// and for the subspace
	qres = iavlStore.Query(querySub)
	require.Equal(t, uint32(errors.CodeOK), qres.Code)
	require.Equal(t, valExpSub1, qres.Value)

	// modify
	iavlStore.Set(k1, v3)
	cid = iavlStore.Commit()

	// query will return old values, as height is fixed
	qres = iavlStore.Query(query)
	require.Equal(t, uint32(errors.CodeOK), qres.Code)
	require.Equal(t, v1, qres.Value)

	// update to latest in the query and we are happy
	query.Height = cid.Version
	qres = iavlStore.Query(query)
	require.Equal(t, uint32(errors.CodeOK), qres.Code)
	require.Equal(t, v3, qres.Value)
	query2 := abci.RequestQuery{Path: "/key", Data: k2, Height: cid.Version}

	qres = iavlStore.Query(query2)
	require.Equal(t, uint32(errors.CodeOK), qres.Code)
	require.Equal(t, v2, qres.Value)
	// and for the subspace
	qres = iavlStore.Query(querySub)
	require.Equal(t, uint32(errors.CodeOK), qres.Code)
	require.Equal(t, valExpSub2, qres.Value)

	// default (height 0) will show latest -1
	query0 := abci.RequestQuery{Path: "/key", Data: k1}
	qres = iavlStore.Query(query0)
	require.Equal(t, uint32(errors.CodeOK), qres.Code)
	require.Equal(t, v1, qres.Value)
}

func BenchmarkIAVLIteratorNext(b *testing.B) {
	db := dbm.NewMemDB()
	treeSize := 1000
	tree := iavl.NewMutableTree(db, cacheSize)
	for i := 0; i < treeSize; i++ {
		key := cmn.RandBytes(4)
		value := cmn.RandBytes(50)
		tree.Set(key, value)
	}
	iavlStore := UnsafeNewStore(tree, numRecent, storeEvery)
	iterators := make([]types.Iterator, b.N/treeSize)
	for i := 0; i < len(iterators); i++ {
		iterators[i] = iavlStore.Iterator([]byte{0}, []byte{255, 255, 255, 255, 255})
	}
	b.ResetTimer()
	for i := 0; i < len(iterators); i++ {
		iter := iterators[i]
		for j := 0; j < treeSize; j++ {
			iter.Next()
		}
	}
}
