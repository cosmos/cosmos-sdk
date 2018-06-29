package store

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	cacheSize        = 100
	numHistory int64 = 5
)

var (
	treeData = map[string]string{
		"hello": "goodbye",
		"aloha": "shalom",
	}
	nMoreData = 0
)

// make a tree and save it
func newTree(t *testing.T, db dbm.DB) (*iavl.VersionedTree, CommitID) {
	tree := iavl.NewVersionedTree(db, cacheSize)
	for k, v := range treeData {
		tree.Set([]byte(k), []byte(v))
	}
	for i := 0; i < nMoreData; i++ {
		key := cmn.RandBytes(12)
		value := cmn.RandBytes(50)
		tree.Set(key, value)
	}
	hash, ver, err := tree.SaveVersion()
	assert.Nil(t, err)
	return tree, CommitID{ver, hash}
}

func TestIAVLStoreGetSetHasDelete(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newTree(t, db)
	iavlStore := newIAVLStore(tree, numHistory)

	key := "hello"

	exists := iavlStore.Has([]byte(key))
	assert.True(t, exists)

	value := iavlStore.Get([]byte(key))
	assert.EqualValues(t, value, treeData[key])

	value2 := "notgoodbye"
	iavlStore.Set([]byte(key), []byte(value2))

	value = iavlStore.Get([]byte(key))
	assert.EqualValues(t, value, value2)

	iavlStore.Delete([]byte(key))

	exists = iavlStore.Has([]byte(key))
	assert.False(t, exists)
}

func TestIAVLIterator(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newTree(t, db)
	iavlStore := newIAVLStore(tree, numHistory)
	iter := iavlStore.Iterator([]byte("aloha"), []byte("hellz"))
	expected := []string{"aloha", "hello"}
	var i int

	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	assert.Equal(t, len(expected), i)

	iter = iavlStore.Iterator([]byte("golang"), []byte("rocks"))
	expected = []string{"hello"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	assert.Equal(t, len(expected), i)

	iter = iavlStore.Iterator(nil, []byte("golang"))
	expected = []string{"aloha"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	assert.Equal(t, len(expected), i)

	iter = iavlStore.Iterator(nil, []byte("shalom"))
	expected = []string{"aloha", "hello"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	assert.Equal(t, len(expected), i)

	iter = iavlStore.Iterator(nil, nil)
	expected = []string{"aloha", "hello"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	assert.Equal(t, len(expected), i)

	iter = iavlStore.Iterator([]byte("golang"), nil)
	expected = []string{"hello"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, treeData[expectedKey])
		i++
	}
	assert.Equal(t, len(expected), i)
}

func TestIAVLSubspaceIterator(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newTree(t, db)
	iavlStore := newIAVLStore(tree, numHistory)

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

	iter := sdk.KVStorePrefixIterator(iavlStore, []byte("test"))
	expected := []string{"test1", "test2", "test3"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, expectedKey)
		i++
	}
	assert.Equal(t, len(expected), i)

	iter = sdk.KVStorePrefixIterator(iavlStore, []byte{byte(55), byte(255), byte(255)})
	expected2 := [][]byte{
		{byte(55), byte(255), byte(255), byte(0)},
		{byte(55), byte(255), byte(255), byte(1)},
		{byte(55), byte(255), byte(255), byte(255)},
	}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected2[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, []byte("test4"))
		i++
	}
	assert.Equal(t, len(expected), i)

	iter = sdk.KVStorePrefixIterator(iavlStore, []byte{byte(255), byte(255)})
	expected2 = [][]byte{
		{byte(255), byte(255), byte(0)},
		{byte(255), byte(255), byte(1)},
		{byte(255), byte(255), byte(255)},
	}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected2[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, []byte("test4"))
		i++
	}
	assert.Equal(t, len(expected), i)
}

func TestIAVLReverseSubspaceIterator(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newTree(t, db)
	iavlStore := newIAVLStore(tree, numHistory)

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

	iter := sdk.KVStoreReversePrefixIterator(iavlStore, []byte("test"))
	expected := []string{"test3", "test2", "test1"}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, expectedKey)
		i++
	}
	assert.Equal(t, len(expected), i)

	iter = sdk.KVStoreReversePrefixIterator(iavlStore, []byte{byte(55), byte(255), byte(255)})
	expected2 := [][]byte{
		{byte(55), byte(255), byte(255), byte(255)},
		{byte(55), byte(255), byte(255), byte(1)},
		{byte(55), byte(255), byte(255), byte(0)},
	}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected2[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, []byte("test4"))
		i++
	}
	assert.Equal(t, len(expected), i)

	iter = sdk.KVStoreReversePrefixIterator(iavlStore, []byte{byte(255), byte(255)})
	expected2 = [][]byte{
		{byte(255), byte(255), byte(255)},
		{byte(255), byte(255), byte(1)},
		{byte(255), byte(255), byte(0)},
	}
	for i = 0; iter.Valid(); iter.Next() {
		expectedKey := expected2[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, []byte("test4"))
		i++
	}
	assert.Equal(t, len(expected), i)
}

func TestIAVLStoreQuery(t *testing.T) {
	db := dbm.NewMemDB()
	tree := iavl.NewVersionedTree(db, cacheSize)
	iavlStore := newIAVLStore(tree, numHistory)

	k1, v1 := []byte("key1"), []byte("val1")
	k2, v2 := []byte("key2"), []byte("val2")
	v3 := []byte("val3")

	ksub := []byte("key")
	KVs0 := []KVPair{}
	KVs1 := []KVPair{
		{k1, v1},
		{k2, v2},
	}
	KVs2 := []KVPair{
		{k1, v3},
		{k2, v2},
	}
	valExpSubEmpty := cdc.MustMarshalBinary(KVs0)
	valExpSub1 := cdc.MustMarshalBinary(KVs1)
	valExpSub2 := cdc.MustMarshalBinary(KVs2)

	cid := iavlStore.Commit()
	ver := cid.Version
	query := abci.RequestQuery{Path: "/key", Data: k1, Height: ver}
	querySub := abci.RequestQuery{Path: "/subspace", Data: ksub, Height: ver}

	// query subspace before anything set
	qres := iavlStore.Query(querySub)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Equal(t, valExpSubEmpty, qres.Value)

	// set data
	iavlStore.Set(k1, v1)
	iavlStore.Set(k2, v2)

	// set data without commit, doesn't show up
	qres = iavlStore.Query(query)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Nil(t, qres.Value)

	// commit it, but still don't see on old version
	cid = iavlStore.Commit()
	qres = iavlStore.Query(query)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Nil(t, qres.Value)

	// but yes on the new version
	query.Height = cid.Version
	qres = iavlStore.Query(query)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Equal(t, v1, qres.Value)

	// and for the subspace
	qres = iavlStore.Query(querySub)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Equal(t, valExpSub1, qres.Value)

	// modify
	iavlStore.Set(k1, v3)
	cid = iavlStore.Commit()

	// query will return old values, as height is fixed
	qres = iavlStore.Query(query)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Equal(t, v1, qres.Value)

	// update to latest in the query and we are happy
	query.Height = cid.Version
	qres = iavlStore.Query(query)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Equal(t, v3, qres.Value)
	query2 := abci.RequestQuery{Path: "/key", Data: k2, Height: cid.Version}
	qres = iavlStore.Query(query2)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Equal(t, v2, qres.Value)
	// and for the subspace
	qres = iavlStore.Query(querySub)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Equal(t, valExpSub2, qres.Value)

	// default (height 0) will show latest -1
	query0 := abci.RequestQuery{Path: "/store", Data: k1}
	qres = iavlStore.Query(query0)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Equal(t, v1, qres.Value)
}
