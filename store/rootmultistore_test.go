package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/merkle"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const useDebugDB = false

func TestMultistoreCommitLoad(t *testing.T) {
	var db dbm.DB = dbm.NewMemDB()
	if useDebugDB {
		db = dbm.NewDebugDB("CMS", db)
	}
	store := newMultiStoreWithMounts(db)
	err := store.LoadLatestVersion()
	assert.Nil(t, err)

	// New store has empty last commit.
	commitID := CommitID{}
	checkStore(t, store, commitID, commitID)

	// Make sure we can get stores by name.
	s1 := store.getStoreByName("store1")
	assert.NotNil(t, s1)
	s3 := store.getStoreByName("store3")
	assert.NotNil(t, s3)
	s77 := store.getStoreByName("store77")
	assert.Nil(t, s77)

	// Make a few commits and check them.
	nCommits := int64(3)
	for i := int64(0); i < nCommits; i++ {
		commitID = store.Commit()
		expectedCommitID := getExpectedCommitID(store, i+1)
		checkStore(t, store, expectedCommitID, commitID)
	}

	// Load the latest multistore again and check version.
	store = newMultiStoreWithMounts(db)
	err = store.LoadLatestVersion()
	assert.Nil(t, err)
	commitID = getExpectedCommitID(store, nCommits)
	checkStore(t, store, commitID, commitID)

	// Commit and check version.
	commitID = store.Commit()
	expectedCommitID := getExpectedCommitID(store, nCommits+1)
	checkStore(t, store, expectedCommitID, commitID)

	// Load an older multistore and check version.
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

	// XXX: confirm old commit is overwritten and we have rolled back
	// LatestVersion
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

func TestMultiStoreQuery(t *testing.T) {
	db := dbm.NewMemDB()
	multi := newMultiStoreWithMounts(db)
	err := multi.LoadLatestVersion()
	assert.Nil(t, err)

	k, v := []byte("wind"), []byte("blows")
	k2, v2 := []byte("water"), []byte("flows")
	// v3 := []byte("is cold")

	cid := multi.Commit()

	// Make sure we can get by name.
	garbage := multi.getStoreByName("bad-name")
	assert.Nil(t, garbage)

	// Set and commit data in one store.
	store1 := multi.getStoreByName("store1").(KVStore)
	store1.Set(k, v)

	// ... and another.
	store2 := multi.getStoreByName("store2").(KVStore)
	store2.Set(k2, v2)

	// Commit the multistore.
	cid = multi.Commit()
	ver := cid.Version

	// Test bad path.
	query := abci.RequestQuery{Path: "/key", Data: k, Height: ver}
	qres := multi.Query(query)
	assert.Equal(t, uint32(sdk.CodeUnknownRequest), qres.Code)

	query.Path = "h897fy32890rf63296r92"
	qres = multi.Query(query)
	assert.Equal(t, uint32(sdk.CodeUnknownRequest), qres.Code)

	// Test invalid store name.
	query.Path = "/garbage/key"
	qres = multi.Query(query)
	assert.Equal(t, uint32(sdk.CodeUnknownRequest), qres.Code)

	// Test valid query with data.
	query.Path = "/store1/key"
	qres = multi.Query(query)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Equal(t, v, qres.Value)

	// Test valid but empty query.
	query.Path = "/store2/key"
	query.Prove = true
	qres = multi.Query(query)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Nil(t, qres.Value)

	// Test store2 data.
	query.Data = k2
	qres = multi.Query(query)
	assert.Equal(t, uint32(sdk.CodeOK), qres.Code)
	assert.Equal(t, v2, qres.Value)
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
