package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/merkle"
)

func TestMultistoreCommitLoad(t *testing.T) {
	db := dbm.NewMemDB()
	store := newMultiStoreWithLoaders(db)
	err := store.LoadLatestVersion()
	assert.Nil(t, err)

	// new store has empty last commit
	commitID := CommitID{}
	checkStore(t, store, commitID, commitID)

	// make a few commits and check them
	nCommits := int64(3)
	for i := int64(0); i < nCommits; i++ {
		commitID = store.Commit()
		expectedCommitID := getExpectedCommitID(store, i+1)
		checkStore(t, store, expectedCommitID, commitID)
	}

	// Load the latest multistore again and check version
	store = newMultiStoreWithLoaders(db)
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
	store = newMultiStoreWithLoaders(db)
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
	store = newMultiStoreWithLoaders(db)
	err = store.LoadLatestVersion()
	assert.Nil(t, err)
	commitID = getExpectedCommitID(store, ver+1)
	checkStore(t, store, commitID, commitID)
}

//-----------------------------------------------------------------------
// utils

func newMultiStoreWithLoaders(db dbm.DB) *rootMultiStore {
	store := NewMultiStore(db)
	storeLoaders := map[string]CommitStoreLoader{
		"store1": newMockCommitStore,
		"store2": newMockCommitStore,
		"store3": newMockCommitStore,
	}
	for name, loader := range storeLoaders {
		store.SetCommitStoreLoader(name, loader)
	}
	return store
}

func checkStore(t *testing.T, store *rootMultiStore, expect, got CommitID) {
	assert.EqualValues(t, expect.Version+1, store.GetCurrentVersion())
	assert.Equal(t, expect, got)
	assert.Equal(t, expect, store.LastCommitID())

}

func getExpectedCommitID(store *rootMultiStore, ver int64) CommitID {
	return CommitID{
		Version: ver,
		Hash:    hashStores(store.substores),
	}
}

func hashStores(stores map[string]CommitStore) []byte {
	m := make(map[string]interface{}, len(stores))
	for name, store := range stores {
		m[name] = substore{
			Name: name,
			substoreCore: substoreCore{
				CommitID: store.Commit(),
			},
		}
	}
	return merkle.SimpleHashFromMap(m)
}

//-----------------------------------------------------------------------
// mockCommitStore

var _ CommitStore = (*mockCommitStore)(nil)

type mockCommitStore struct {
	id CommitID
}

func newMockCommitStore(id CommitID) (CommitStore, error) {
	return &mockCommitStore{id}, nil
}

func (cs *mockCommitStore) Commit() CommitID {
	return cs.id
}
func (cs *mockCommitStore) CacheWrap() CacheWrap {
	cs2 := *cs
	return &cs2
}
func (cs *mockCommitStore) Write() {}
