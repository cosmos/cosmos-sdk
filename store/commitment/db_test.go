package commitment

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2/commitment/iavl"
	"cosmossdk.io/store/v2/commitment/types"
)

func generateTree(treeType string) types.Tree {
	if treeType == "iavl" {
		cfg := iavl.DefaultConfig()
		db := dbm.NewMemDB()
		tree := iavl.NewIavlTree(db, log.NewNopLogger(), cfg)

		return tree
	}

	return nil
}

func TestIavlTree(t *testing.T) {
	// generate a new tree
	tree := generateTree("iavl")
	require.NotNil(t, tree)

	intialVersion := tree.GetLatestVersion()
	require.Equal(t, uint64(0), intialVersion)

	// write a batch of version 1
	batch1 := types.NewBatch()
	batch1.Add([]byte("key1"), []byte("value1"))
	batch1.Add([]byte("key2"), []byte("value2"))
	batch1.Add([]byte("key3"), []byte("value3"))

	err := tree.WriteBatch(batch1)
	require.NoError(t, err)

	workingHash := tree.WorkingHash()
	require.NotNil(t, workingHash)
	require.Equal(t, uint64(0), tree.GetLatestVersion())

	// commit the batch
	commitHash, err := tree.Commit()
	require.NoError(t, err)
	require.Equal(t, workingHash, commitHash)
	require.Equal(t, uint64(1), tree.GetLatestVersion())
	version1Hash := tree.WorkingHash()

	// write a batch of version 2
	batch2 := types.NewBatch()
	batch2.Add([]byte("key4"), []byte("value4"))
	batch2.Add([]byte("key5"), []byte("value5"))
	batch2.Add([]byte("key6"), []byte("value6"))
	batch2.Add([]byte("key1"), nil) // delete key1
	err = tree.WriteBatch(batch2)
	require.NoError(t, err)
	workingHash = tree.WorkingHash()
	require.NotNil(t, workingHash)
	commitHash, err = tree.Commit()
	require.NoError(t, err)
	require.Equal(t, workingHash, commitHash)

	// get proof for key1
	proof, err := tree.GetProof(1, []byte("key1"))
	require.NoError(t, err)
	require.NotNil(t, proof.GetExist())

	proof, err = tree.GetProof(2, []byte("key1"))
	require.NoError(t, err)
	require.NotNil(t, proof.GetNonexist())

	// load version 1
	err = tree.LoadVersion(1)
	require.NoError(t, err)
	require.Equal(t, version1Hash, tree.WorkingHash())

	// close the db
	require.NoError(t, tree.Close())
}
