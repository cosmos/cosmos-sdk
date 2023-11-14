package iavl

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
)

func generateTree(treeType string) *IavlTree {
	cfg := DefaultConfig()
	db := dbm.NewMemDB()
	return NewIavlTree(db, log.NewNopLogger(), cfg)
}

func TestIavlTree(t *testing.T) {
	// generate a new tree
	tree := generateTree("iavl")
	require.NotNil(t, tree)

	initVersion := tree.GetLatestVersion()
	require.Equal(t, uint64(0), initVersion)

	// write a batch of version 1
	cs1 := store.NewChangeset()
	cs1.Add([]byte("key1"), []byte("value1"))
	cs1.Add([]byte("key2"), []byte("value2"))
	cs1.Add([]byte("key3"), []byte("value3"))

	err := tree.WriteBatch(cs1)
	require.NoError(t, err)

	workingHash := tree.WorkingHash()
	require.NotNil(t, workingHash)
	require.Equal(t, uint64(0), tree.GetLatestVersion())

	// commit the batch
	commitHash, err := tree.Commit()
	require.NoError(t, err)
	require.Equal(t, workingHash, commitHash)
	require.Equal(t, uint64(1), tree.GetLatestVersion())

	// write a batch of version 2
	cs2 := store.NewChangeset()
	cs2.Add([]byte("key4"), []byte("value4"))
	cs2.Add([]byte("key5"), []byte("value5"))
	cs2.Add([]byte("key6"), []byte("value6"))
	cs2.Add([]byte("key1"), nil) // delete key1
	err = tree.WriteBatch(cs2)
	require.NoError(t, err)
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
	cs3 := store.NewChangeset()
	cs3.Add([]byte("key7"), []byte("value7"))
	cs3.Add([]byte("key8"), []byte("value8"))
	err = tree.WriteBatch(cs3)
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
