package store

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/tendermint/iavl"
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

func TestIAVLStoreLoader(t *testing.T) {
	db := dbm.NewMemDB()
	_, id := newTree(t, db)

	iavlLoader := NewIAVLStoreLoader(db, cacheSize, numHistory)
	commitStore, err := iavlLoader(id)
	assert.Nil(t, err)

	id2 := commitStore.Commit()

	assert.Equal(t, id.Hash, id2.Hash)
	assert.Equal(t, id.Version+1, id2.Version)
}

func TestIAVLStoreGetSetHasRemove(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newTree(t, db)
	iavlStore := newIAVLStore(tree, numHistory)

	key := "hello"

	exists := iavlStore.Has([]byte(key))
	assert.True(t, exists)

	value, exists := iavlStore.Get([]byte(key))
	assert.True(t, exists)
	assert.EqualValues(t, value, treeData[key])

	value2 := "notgoodbye"
	prev := iavlStore.Set([]byte(key), []byte(value2))
	assert.EqualValues(t, value, prev)

	value, exists = iavlStore.Get([]byte(key))
	assert.True(t, exists)
	assert.EqualValues(t, value, value2)

	prev, removed := iavlStore.Remove([]byte(key))
	assert.True(t, removed)
	assert.EqualValues(t, value2, prev)

	exists = iavlStore.Has([]byte(key))
	assert.False(t, exists)
}

func TestIAVLIterator(t *testing.T) {
	db := dbm.NewMemDB()
	tree, _ := newTree(t, db)
	iavlStore := newIAVLStore(tree, numHistory)
	iter := iavlStore.Iterator([]byte("aloha"), []byte("hellz"))
	expected := []string{"aloha", "hello"}
	for i := 0; iter.Valid(); iter.Next() {
		expectedKey := expected[i]
		key, value := iter.Key(), iter.Value()
		assert.EqualValues(t, key, expectedKey)
		assert.EqualValues(t, value, treeData[expectedKey])
		i += 1
	}
}
