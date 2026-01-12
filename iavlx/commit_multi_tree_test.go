package iavlx

import (
	"log/slog"
	"testing"

	store "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"
)

func TestCommitMultiTree_Reload(t *testing.T) {
	dir := t.TempDir()
	var db *CommitMultiTree

	testStoreKey := store.NewKVStoreKey("test")
	loadDb := func() {
		var err error
		db, err = LoadDB(dir, &Options{}, slog.Default())
		require.NoError(t, err)
		db.MountStoreWithDB(testStoreKey, store.StoreTypeIAVL, nil)
		require.NoError(t, db.LoadLatestVersion())
	}

	// open db & create some data
	loadDb()
	testStore := db.GetCommitKVStore(testStoreKey)
	testStore.Set([]byte("key1"), []byte("value1"))
	testStore.Set([]byte("key2"), []byte("value2"))
	commitId := db.Commit()

	// reload the DB
	require.NoError(t, db.Close())
	loadDb()

	// verify data is still there
	testStore = db.GetCommitKVStore(testStoreKey)
	val1 := testStore.Get([]byte("key1"))
	require.Equal(t, []byte("value1"), val1)
	val2 := testStore.Get([]byte("key2"))
	require.Equal(t, []byte("value2"), val2)

	// verify commit ID is the same
	require.Equal(t, commitId, db.LastCommitID())
}
