package iavl

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2/commitment"
)

func TestCommitterSuite(t *testing.T) {
	s := &commitment.CommitStoreTestSuite{
		NewStore: func(db dbm.DB, storeKeys []string, logger log.Logger) (*commitment.CommitStore, error) {
			multiTrees := make(map[string]commitment.Tree)
			cfg := DefaultConfig()
			for _, storeKey := range storeKeys {
				prefixDB := dbm.NewPrefixDB(db, []byte(storeKey))
				multiTrees[storeKey] = NewIavlTree(prefixDB, logger, cfg)
			}
			return commitment.NewCommitStore(multiTrees, logger)
		},
	}

	suite.Run(t, s)
}

func generateTree() *IavlTree {
	cfg := DefaultConfig()
	db := dbm.NewMemDB()
	return NewIavlTree(db, log.NewNopLogger(), cfg)
}

func TestIavlTree(t *testing.T) {
	// generate a new tree
	tree := generateTree()
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
