package iavl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corelog "cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2/commitment"
	dbm "cosmossdk.io/store/v2/db"
)

func TestCommitterSuite(t *testing.T) {
	s := &commitment.CommitStoreTestSuite{
		NewStore: func(db corestore.KVStoreWithBatch, storeKeys, oldStoreKeys []string, logger corelog.Logger) (*commitment.CommitStore, error) {
			multiTrees := make(map[string]commitment.Tree)
			cfg := DefaultConfig()
			mountTreeFn := func(storeKey string) (commitment.Tree, error) {
				prefixDB := dbm.NewPrefixDB(db, []byte(storeKey))
				return NewIavlTree(prefixDB, logger, cfg), nil
			}
			for _, storeKey := range storeKeys {
				multiTrees[storeKey], _ = mountTreeFn(storeKey)
			}
			oldTrees := make(map[string]commitment.Tree)
			for _, storeKey := range oldStoreKeys {
				oldTrees[storeKey], _ = mountTreeFn(storeKey)
			}

			return commitment.NewCommitStore(multiTrees, oldTrees, db, logger)
		},
	}

	suite.Run(t, s)
}

func generateTree() *IavlTree {
	cfg := DefaultConfig()
	db := dbm.NewMemDB()
	return NewIavlTree(db, coretesting.NewNopLogger(), cfg)
}

func TestIavlTree(t *testing.T) {
	// generate a new tree
	tree := generateTree()
	require.NotNil(t, tree)

	initVersion, err := tree.GetLatestVersion()
	require.NoError(t, err)
	require.Equal(t, uint64(0), initVersion)

	// write a batch of version 1
	require.NoError(t, tree.Set([]byte("key1"), []byte("value1")))
	require.NoError(t, tree.Set([]byte("key2"), []byte("value2")))
	require.NoError(t, tree.Set([]byte("key3"), []byte("value3")))

	workingHash := tree.WorkingHash()
	require.NotNil(t, workingHash)
	v, err := tree.GetLatestVersion()
	require.NoError(t, err)
	require.Equal(t, uint64(0), v)

	// commit the batch
	commitHash, version, err := tree.Commit()
	require.NoError(t, err)
	require.Equal(t, version, uint64(1))
	require.Equal(t, workingHash, commitHash)
	v, err = tree.GetLatestVersion()
	require.NoError(t, err)
	require.Equal(t, uint64(1), v)

	// ensure we can get expected values
	bz, err := tree.Get(1, []byte("key1"))
	require.NoError(t, err)
	require.Equal(t, []byte("value1"), bz)

	bz, err = tree.Get(2, []byte("key1"))
	require.Error(t, err)
	require.Nil(t, bz)

	// write a batch of version 2
	require.NoError(t, tree.Set([]byte("key4"), []byte("value4")))
	require.NoError(t, tree.Set([]byte("key5"), []byte("value5")))
	require.NoError(t, tree.Set([]byte("key6"), []byte("value6")))
	require.NoError(t, tree.Remove([]byte("key1"))) // delete key1
	version2Hash := tree.WorkingHash()
	require.NotNil(t, version2Hash)
	commitHash, version, err = tree.Commit()
	require.NoError(t, err)
	require.Equal(t, version, uint64(2))
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
	_, _, err = tree.Commit()
	require.NoError(t, err)

	// prune version 1
	err = tree.Prune(1)
	require.NoError(t, err)
	v, err = tree.GetLatestVersion()
	require.NoError(t, err)
	require.Equal(t, uint64(3), v)
	// async pruning check
	checkErr := func() bool {
		if _, err := tree.tree.LoadVersion(1); err != nil {
			return true
		}
		return false
	}
	require.Eventually(t, checkErr, 2*time.Second, 100*time.Millisecond)

	// load version 2
	err = tree.LoadVersion(2)
	require.NoError(t, err)
	require.Equal(t, version2Hash, tree.WorkingHash())

	// close the db
	require.NoError(t, tree.Close())
}
