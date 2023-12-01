package iavl

import (
	"fmt"
	"io"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/snapshots"
	snapshottypes "cosmossdk.io/store/v2/snapshots/types"
)

func generateTree(storeKey string) *IavlTree {
	cfg := DefaultConfig()
	db := dbm.NewMemDB()
	return NewIavlTree(db, log.NewNopLogger(), storeKey, cfg)
}

func TestIavlTree(t *testing.T) {
	// generate a new tree
	tree := generateTree("iavl")
	require.NotNil(t, tree)

	initVersion := tree.GetLatestVersion()
	require.Equal(t, uint64(0), initVersion)

	// write a batch of version 1
	require.NoError(t, tree.Set([]byte("key1"), []byte("value1")))
	require.NoError(t, tree.Set([]byte("key2"), []byte("value2")))
	require.NoError(t, tree.Set([]byte("key3"), []byte("value3")))

	workingHash := tree.WorkingHash()
	require.NotNil(t, workingHash)
	require.Equal(t, uint64(0), tree.GetLatestVersion())

	// commit the batch
	commitHash, err := tree.Commit()
	require.NoError(t, err)
	require.Equal(t, workingHash, commitHash)
	require.Equal(t, uint64(1), tree.GetLatestVersion())

	// write a batch of version 2
	require.NoError(t, tree.Set([]byte("key4"), []byte("value4")))
	require.NoError(t, tree.Set([]byte("key5"), []byte("value5")))
	require.NoError(t, tree.Set([]byte("key6"), []byte("value6")))
	require.NoError(t, tree.Remove([]byte("key1"))) // delete key1
	version2Hash := tree.WorkingHash()
	require.NotNil(t, version2Hash)
	commitHash, err = tree.Commit()
	require.NoError(t, err)
	require.Equal(t, version2Hash, commitHash)

	// get proof for key1
	proof, err := tree.GetProof(1, []byte("key1"))
	require.NoError(t, err)
	require.NotNil(t, proof.GetExist())

	proof, err = tree.GetProof(2, []byte("key1"))
	require.NoError(t, err)
	require.NotNil(t, proof.GetNonexist())

	// write a batch of version 3
	require.NoError(t, tree.Set([]byte("key7"), []byte("value7")))
	require.NoError(t, tree.Set([]byte("key8"), []byte("value8")))
	require.NoError(t, err)
	_, err = tree.Commit()
	require.NoError(t, err)

	// prune version 1
	err = tree.Prune(1)
	require.NoError(t, err)
	require.Equal(t, uint64(3), tree.GetLatestVersion())
	err = tree.LoadVersion(1)
	require.Error(t, err)

	// load version 2
	err = tree.LoadVersion(2)
	require.NoError(t, err)
	require.Equal(t, version2Hash, tree.WorkingHash())

	// close the db
	require.NoError(t, tree.Close())
}

func TestSnapshotter(t *testing.T) {
	// generate a new tree
	storeKey := "store"
	tree := generateTree(storeKey)
	require.NotNil(t, tree)

	latestVersion := uint64(10)
	kvCount := 10
	for i := uint64(1); i <= latestVersion; i++ {
		cs := store.NewChangeset()
		for j := 0; j < kvCount; j++ {
			key := []byte(fmt.Sprintf("key-%d-%d", i, j))
			value := []byte(fmt.Sprintf("value-%d-%d", i, j))
			cs.Add(key, value)
		}
		err := tree.WriteBatch(cs)
		require.NoError(t, err)

		_, err = tree.Commit()
		require.NoError(t, err)
	}

	latestHash := tree.WorkingHash()

	// create a snapshot
	dummyExtensionItem := snapshottypes.SnapshotItem{
		Item: &snapshottypes.SnapshotItem_Extension{
			Extension: &snapshottypes.SnapshotExtensionMeta{
				Name:   "test",
				Format: 1,
			},
		},
	}
	target := generateTree("")
	chunks := make(chan io.ReadCloser, kvCount*int(latestVersion))
	go func() {
		streamWriter := snapshots.NewStreamWriter(chunks)
		require.NotNil(t, streamWriter)
		defer streamWriter.Close()
		err := tree.Snapshot(latestVersion, streamWriter)
		require.NoError(t, err)
		// write an extension metadata
		err = streamWriter.WriteMsg(&dummyExtensionItem)
		require.NoError(t, err)
	}()

	streamReader, err := snapshots.NewStreamReader(chunks)
	chStorage := make(chan *store.KVPair, 100)
	require.NoError(t, err)
	nextItem, err := target.Restore(latestVersion, snapshottypes.CurrentFormat, streamReader, chStorage)
	require.NoError(t, err)
	require.Equal(t, *dummyExtensionItem.GetExtension(), *nextItem.GetExtension())

	// check the store key
	require.Equal(t, storeKey, target.storeKey)

	// check the restored tree hash
	require.Equal(t, latestHash, target.WorkingHash())
}
