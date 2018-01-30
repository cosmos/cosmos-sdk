package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/merkle"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMultistoreCommitLoad(t *testing.T) {
	db := dbm.NewMemDB()
	store := newMultiStoreWithMounts(db)
	err := store.LoadLatestVersion()
	assert.Nil(t, err)

	// new store has empty last commit
	commitID := CommitID{}
	checkStore(t, store, commitID, commitID)

	// make sure we can get stores by name
	s1 := store.GetStoreByName("store1")
	assert.NotNil(t, s1)
	s3 := store.GetStoreByName("store3")
	assert.NotNil(t, s3)
	s77 := store.GetStoreByName("store77")
	assert.Nil(t, s77)

	// make a few commits and check them
	nCommits := int64(3)
	for i := int64(0); i < nCommits; i++ {
		commitID = store.Commit()
		expectedCommitID := getExpectedCommitID(store, i+1)
		checkStore(t, store, expectedCommitID, commitID)
	}

	// Load the latest multistore again and check version
	store = newMultiStoreWithMounts(db)
	err = store.LoadLatestVersion()
	assert.Nil(t, err)
	commitID = getExpectedCommitID(store, nCommits)
	checkStore(t, store, commitID, commitID)

	// commit and check version
	commitID = store.Commit()
	expectedCommitID := getExpectedCommitID(store, nCommits+1)
	checkStore(t, store, expectedCommitID, commitID)

	// Load an older multistore and check version
	ver := nCommits - 1
	store = newMultiStoreWithMounts(db)
	err = store.LoadVersion(ver)
	assert.Nil(t, err)
	commitID = getExpectedCommitID(store, ver)
	checkStore(t, store, commitID, commitID)

	// XXX: commit this older version
	commitID = store.Commit()
	expectedCommitID = getExpectedCommitID(store, ver+1)
	checkStore(t, store, expectedCommitID, commitID)

	// XXX: confirm old commit is overwritten and
	// we have rolled back LatestVersion
	store = newMultiStoreWithMounts(db)
	err = store.LoadLatestVersion()
	assert.Nil(t, err)
	commitID = getExpectedCommitID(store, ver+1)
	checkStore(t, store, commitID, commitID)
}

func TestParsePath(t *testing.T) {
	_, _, err := parsePath("foo")
	assert.Error(t, err)

	store, subpath, err := parsePath("/foo")
	assert.NoError(t, err)
	assert.Equal(t, store, "foo")
	assert.Equal(t, subpath, "")

	store, subpath, err = parsePath("/fizz/bang/baz")
	assert.NoError(t, err)
	assert.Equal(t, store, "fizz")
	assert.Equal(t, subpath, "/bang/baz")

	substore, subsubpath, err := parsePath(subpath)
	assert.NoError(t, err)
	assert.Equal(t, substore, "bang")
	assert.Equal(t, subsubpath, "/baz")

}

//-----------------------------------------------------------------------
// utils

func newMultiStoreWithMounts(db dbm.DB) *rootMultiStore {
	store := NewCommitMultiStore(db)
	store.MountStoreWithDB(
		sdk.NewKVStoreKey("store1"), sdk.StoreTypeIAVL, db)
	store.MountStoreWithDB(
		sdk.NewKVStoreKey("store2"), sdk.StoreTypeIAVL, db)
	store.MountStoreWithDB(
		sdk.NewKVStoreKey("store3"), sdk.StoreTypeIAVL, db)
	return store
}

func checkStore(t *testing.T, store *rootMultiStore, expect, got CommitID) {
	assert.Equal(t, expect, got)
	assert.Equal(t, expect, store.LastCommitID())

}

func getExpectedCommitID(store *rootMultiStore, ver int64) CommitID {
	return CommitID{
		Version: ver,
		Hash:    hashStores(store.stores),
	}
}

func hashStores(stores map[StoreKey]CommitStore) []byte {
	m := make(map[string]merkle.Hasher, len(stores))
	for key, store := range stores {
		name := key.Name()
		m[name] = storeInfo{
			Name: name,
			Core: storeCore{
				CommitID: store.LastCommitID(),
				// StoreType: store.GetStoreType(),
			},
		}
	}
	return merkle.SimpleHashFromMap(m)
}
