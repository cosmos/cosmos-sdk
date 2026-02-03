package iavl

import (
	"context"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	store "cosmossdk.io/store/types"
)

func TestCommitMultiTree_Reload(t *testing.T) {
	dir := t.TempDir()
	var db *CommitMultiTree

	testStoreKey := store.NewKVStoreKey("test")
	loadDb := func() {
		var err error
		db, err = LoadDB(dir, Options{})
		require.NoError(t, err)
		db.MountStoreWithDB(testStoreKey, store.StoreTypeIAVL, nil)
		require.NoError(t, db.LoadLatestVersion())
	}

	// open db & create some data
	loadDb()
	cacheMs := db.CacheMultiStore()
	testStore := cacheMs.GetKVStore(testStoreKey)
	testStore.Set([]byte("key1"), []byte("value1"))
	testStore.Set([]byte("key2"), []byte("value2"))
	committer, err := db.StartCommit(context.Background(), cacheMs, cmtproto.Header{})
	require.NoError(t, err)
	commitId, err := committer.Finalize()
	require.NoError(t, err)

	// reload the DB
	require.NoError(t, db.Close())
	loadDb()

	// verify data is still there
	cacheMs = db.CacheMultiStore()
	testStore = cacheMs.GetKVStore(testStoreKey)
	val1 := testStore.Get([]byte("key1"))
	require.Equal(t, []byte("value1"), val1)
	val2 := testStore.Get([]byte("key2"))
	require.Equal(t, []byte("value2"), val2)
	committer, err = db.StartCommit(context.Background(), cacheMs, cmtproto.Header{})
	require.NoError(t, err)
	commitId, err = committer.Finalize()
	require.NoError(t, err)

	// verify commit ID is the same
	require.Equal(t, commitId, db.LastCommitID())
}
