package commitment

import (
	"fmt"
	"io"
	"strings"

	"github.com/stretchr/testify/suite"

	corelog "cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/snapshots"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
)

const (
	storeKey1 = "store1"
	storeKey2 = "store2"
	storeKey3 = "store3"
)

// CommitStoreTestSuite is a test suite to be used for all tree backends.
type CommitStoreTestSuite struct {
	suite.Suite

	NewStore func(db corestore.KVStoreWithBatch, dbDir string, storeKeys, oldStoreKeys []string, logger corelog.Logger) (*CommitStore, error)
	TreeType string
}

// TestStore_Snapshotter tests the snapshot functionality of the CommitStore.
// This test verifies that the store can correctly create snapshots and restore from them.
// The test follows these steps:
//
// 1. Setup & Data Population:
//   - Creates a new CommitStore with two stores (store1 and store2)
//   - Writes 10 versions of data (version 1-10)
//   - For each version, writes 10 key-value pairs to each store
//   - Total data: 2 stores * 10 versions * 10 pairs = 200 key-value pairs
//   - Keys are formatted as "key-{version}-{index}"
//   - Values are formatted as "value-{version}-{index}"
//   - Each version is committed to get a CommitInfo
//
// 2. Snapshot Creation:
//   - Creates a dummy extension item for metadata testing
//   - Sets up a new target store for restoration
//   - Creates a channel for snapshot chunks
//   - Launches a goroutine to:
//   - Create a snapshot writer
//   - Take a snapshot at version 10
//   - Write extension metadata
//
// 3. Snapshot Restoration:
//   - Creates a snapshot reader from the chunks
//   - Sets up a channel for state changes during restoration
//   - Launches a goroutine to collect restored key-value pairs
//   - Restores the snapshot into the target store
//   - Verifies the extension metadata was preserved
//
// 4. Verification:
//   - Confirms all 200 key-value pairs were restored correctly
//   - Verifies the format: "{storeKey}_key-{version}-{index}" -> "value-{version}-{index}"
//   - Checks that the restored store's Merkle tree hashes match the original
//   - Ensures store integrity by comparing CommitInfo hashes
func (s *CommitStoreTestSuite) TestStore_Snapshotter() {
	storeKeys := []string{storeKey1, storeKey2}
	commitStore, err := s.NewStore(dbm.NewMemDB(), s.T().TempDir(), storeKeys, nil, coretesting.NewNopLogger())
	s.Require().NoError(err)

	// We'll create 10 versions of data
	latestVersion := uint64(10)
	kvCount := 10
	var cInfo *proof.CommitInfo

	// For each version 1-10
	for i := uint64(1); i <= latestVersion; i++ {
		// Create KV pairs for each store
		kvPairs := make(map[string]corestore.KVPairs)
		for _, storeKey := range storeKeys {
			kvPairs[storeKey] = corestore.KVPairs{}
			// Create 10 KV pairs for this store
			for j := 0; j < kvCount; j++ {
				key := []byte(fmt.Sprintf("key-%d-%d", i, j))
				value := []byte(fmt.Sprintf("value-%d-%d", i, j))
				kvPairs[storeKey] = append(kvPairs[storeKey], corestore.KVPair{Key: key, Value: value})
			}
		}
		// Write and commit the changes for this version
		s.Require().NoError(commitStore.WriteChangeset(corestore.NewChangesetWithPairs(i, kvPairs)))
		cInfo, err = commitStore.Commit(i)
		s.Require().NoError(err)
	}

	s.Require().Equal(len(storeKeys), len(cInfo.StoreInfos))

	// create a snapshot
	dummyExtensionItem := snapshotstypes.SnapshotItem{
		Item: &snapshotstypes.SnapshotItem_Extension{
			Extension: &snapshotstypes.SnapshotExtensionMeta{
				Name:   "test",
				Format: 1,
			},
		},
	}

	targetStore, err := s.NewStore(dbm.NewMemDB(), s.T().TempDir(), storeKeys, nil, coretesting.NewNopLogger())
	s.Require().NoError(err)

	chunks := make(chan io.ReadCloser, kvCount*int(latestVersion))
	go func() {
		streamWriter := snapshots.NewStreamWriter(chunks)
		s.Require().NotNil(streamWriter)
		defer streamWriter.Close()
		err := commitStore.Snapshot(latestVersion, streamWriter)
		s.Require().NoError(err)
		// write an extension metadata
		err = streamWriter.WriteMsg(&dummyExtensionItem)
		s.Require().NoError(err)
	}()

	streamReader, err := snapshots.NewStreamReader(chunks)
	s.Require().NoError(err)

	nextItem, err := targetStore.Restore(latestVersion, snapshotstypes.CurrentFormat, streamReader)
	s.Require().NoError(err)
	s.Require().Equal(*dummyExtensionItem.GetExtension(), *nextItem.GetExtension())

	// check the restored tree hash
	targetCommitInfo, err := targetStore.GetCommitInfo(latestVersion)
	s.Require().NoError(err)
	for _, storeInfo := range targetCommitInfo.StoreInfos {
		matched := false
		for _, latestStoreInfo := range cInfo.StoreInfos {
			if strings.EqualFold(storeInfo.Name, latestStoreInfo.Name) {
				s.Require().Equal(latestStoreInfo.GetHash(), storeInfo.GetHash())
				matched = true
			}
		}
		s.Require().True(matched)
	}
}

func (s *CommitStoreTestSuite) TestStore_LoadVersion() {
	storeKeys := []string{storeKey1, storeKey2}
	mdb := dbm.NewMemDB()
	dbDir := s.T().TempDir()
	commitStore, err := s.NewStore(mdb, dbDir, storeKeys, nil, coretesting.NewNopLogger())
	s.Require().NoError(err)

	latestVersion := uint64(10)
	kvCount := 10
	for i := uint64(1); i <= latestVersion; i++ {
		kvPairs := make(map[string]corestore.KVPairs)
		for _, storeKey := range storeKeys {
			kvPairs[storeKey] = corestore.KVPairs{}
			for j := 0; j < kvCount; j++ {
				key := []byte(fmt.Sprintf("key-%d-%d", i, j))
				value := []byte(fmt.Sprintf("value-%d-%d", i, j))
				kvPairs[storeKey] = append(kvPairs[storeKey], corestore.KVPair{Key: key, Value: value})
			}
		}
		s.Require().NoError(commitStore.WriteChangeset(corestore.NewChangesetWithPairs(i, kvPairs)))
		_, err = commitStore.Commit(i)
		s.Require().NoError(err)
	}

	// load the store with the latest version
	targetStore, err := s.NewStore(mdb, dbDir, storeKeys, nil, coretesting.NewNopLogger())
	s.Require().NoError(err)
	err = targetStore.LoadVersion(latestVersion)
	s.Require().NoError(err)
	// check the store
	for i := uint64(1); i <= latestVersion; i++ {
		commitInfo, _ := targetStore.GetCommitInfo(i)
		s.Require().NotNil(commitInfo)
		s.Require().Equal(i, uint64(commitInfo.Version))
	}

	// rollback to a previous version
	rollbackVersion := uint64(5)
	rollbackStore, err := s.NewStore(mdb, dbDir, storeKeys, nil, coretesting.NewNopLogger())
	s.Require().NoError(err)
	err = rollbackStore.LoadVersion(rollbackVersion)
	s.Require().NoError(err)
	// check the store
	v, err := rollbackStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(rollbackVersion, v)
	for i := uint64(1); i <= latestVersion; i++ {
		commitInfo, _ := rollbackStore.GetCommitInfo(i)
		if i > rollbackVersion {
			s.Require().Nil(commitInfo)
		} else {
			s.Require().NotNil(commitInfo)
		}
	}
}

func (s *CommitStoreTestSuite) TestStore_Pruning() {
	storeKeys := []string{storeKey1, storeKey2}
	pruneOpts := store.NewPruningOptionWithCustom(10, 5)
	commitStore, err := s.NewStore(dbm.NewMemDB(), s.T().TempDir(), storeKeys, nil, coretesting.NewNopLogger())
	s.Require().NoError(err)

	latestVersion := uint64(100)
	kvCount := 10
	for i := uint64(1); i <= latestVersion; i++ {
		kvPairs := make(map[string]corestore.KVPairs)
		for _, storeKey := range storeKeys {
			kvPairs[storeKey] = corestore.KVPairs{}
			for j := 0; j < kvCount; j++ {
				key := []byte(fmt.Sprintf("key-%d-%d", i, j))
				value := []byte(fmt.Sprintf("value-%d-%d", i, j))
				kvPairs[storeKey] = append(kvPairs[storeKey], corestore.KVPair{Key: key, Value: value})
			}
		}
		s.Require().NoError(commitStore.WriteChangeset(corestore.NewChangesetWithPairs(i, kvPairs)))

		_, err = commitStore.Commit(i)
		s.Require().NoError(err)

		if prune, pruneVersion := pruneOpts.ShouldPrune(i); prune {
			s.Require().NoError(commitStore.Prune(pruneVersion))
		}

	}

	pruneVersion := latestVersion - pruneOpts.KeepRecent - 1
	// check the store
	for i := uint64(1); i <= latestVersion; i++ {
		commitInfo, _ := commitStore.GetCommitInfo(i)
		if i <= pruneVersion {
			s.Require().Nil(commitInfo)
		} else {
			s.Require().NotNil(commitInfo)
		}
	}
}

func (s *CommitStoreTestSuite) TestStore_GetProof() {
	storeKeys := []string{storeKey1, storeKey2}
	commitStore, err := s.NewStore(dbm.NewMemDB(), s.T().TempDir(), storeKeys, nil, coretesting.NewNopLogger())
	s.Require().NoError(err)

	toVersion := uint64(10)
	keyCount := 5

	// commit some changes
	for version := uint64(1); version <= toVersion; version++ {
		cs := corestore.NewChangeset(version)
		for _, storeKey := range storeKeys {
			for i := 0; i < keyCount; i++ {
				cs.Add([]byte(storeKey), []byte(fmt.Sprintf("key-%d-%d", version, i)), []byte(fmt.Sprintf("value-%d-%d", version, i)), false)
			}
		}
		err := commitStore.WriteChangeset(cs)
		s.Require().NoError(err)
		_, err = commitStore.Commit(version)
		s.Require().NoError(err)
	}

	// get proof
	for version := uint64(1); version <= toVersion; version++ {
		for _, storeKey := range storeKeys {
			for i := 0; i < keyCount; i++ {
				_, err := commitStore.GetProof([]byte(storeKey), version, []byte(fmt.Sprintf("key-%d-%d", version, i)))
				s.Require().NoError(err)
			}
		}
	}

	// prune version 1
	s.Require().NoError(commitStore.Prune(1))

	// check if proof for version 1 is pruned
	_, err = commitStore.GetProof([]byte(storeKeys[0]), 1, []byte(fmt.Sprintf("key-%d-%d", 1, 0)))
	s.Require().Error(err)
	// check the commit info
	commit, _ := commitStore.GetCommitInfo(1)
	s.Require().Nil(commit)
}

func (s *CommitStoreTestSuite) TestStore_Get() {
	storeKeys := []string{storeKey1, storeKey2}
	commitStore, err := s.NewStore(dbm.NewMemDB(), s.T().TempDir(), storeKeys, nil, coretesting.NewNopLogger())
	s.Require().NoError(err)

	toVersion := uint64(10)
	keyCount := 5

	// commit some changes
	for version := uint64(1); version <= toVersion; version++ {
		cs := corestore.NewChangeset(version)
		for _, storeKey := range storeKeys {
			for i := 0; i < keyCount; i++ {
				cs.Add([]byte(storeKey), []byte(fmt.Sprintf("key-%d-%d", version, i)), []byte(fmt.Sprintf("value-%d-%d", version, i)), false)
			}
		}
		err := commitStore.WriteChangeset(cs)
		s.Require().NoError(err)
		_, err = commitStore.Commit(version)
		s.Require().NoError(err)
	}

	// get proof
	for version := uint64(1); version <= toVersion; version++ {
		for _, storeKey := range storeKeys {
			for i := 0; i < keyCount; i++ {
				val, err := commitStore.Get([]byte(storeKey), version, []byte(fmt.Sprintf("key-%d-%d", version, i)))
				s.Require().NoError(err)
				s.Require().Equal([]byte(fmt.Sprintf("value-%d-%d", version, i)), val)
			}
		}
	}
}

func (s *CommitStoreTestSuite) TestStore_Upgrades() {
	storeKeys := []string{storeKey1, storeKey2, storeKey3}
	commitDB := dbm.NewMemDB()
	commitDir := s.T().TempDir()
	commitStore, err := s.NewStore(commitDB, commitDir, storeKeys, nil, coretesting.NewNopLogger())
	s.Require().NoError(err)

	latestVersion := uint64(10)
	kvCount := 10
	for i := uint64(1); i <= latestVersion; i++ {
		kvPairs := make(map[string]corestore.KVPairs)
		for _, storeKey := range storeKeys {
			kvPairs[storeKey] = corestore.KVPairs{}
			for j := 0; j < kvCount; j++ {
				key := []byte(fmt.Sprintf("key-%d-%d", i, j))
				value := []byte(fmt.Sprintf("value-%d-%d", i, j))
				kvPairs[storeKey] = append(kvPairs[storeKey], corestore.KVPair{Key: key, Value: value})
			}
		}
		s.Require().NoError(commitStore.WriteChangeset(corestore.NewChangesetWithPairs(i, kvPairs)))
		_, err = commitStore.Commit(i)
		s.Require().NoError(err)
	}

	// create a new commitment store with upgrades
	upgrades := &corestore.StoreUpgrades{
		Added:   []string{"newStore1", "newStore2"},
		Deleted: []string{storeKey3},
	}
	newStoreKeys := []string{storeKey1, storeKey2, storeKey3, "newStore1", "newStore2"}
	realStoreKeys := []string{storeKey1, storeKey2, "newStore1", "newStore2"}
	oldStoreKeys := []string{storeKey3}
	commitStore, err = s.NewStore(commitDB, commitDir, newStoreKeys, oldStoreKeys, coretesting.NewNopLogger())
	s.Require().NoError(err)
	err = commitStore.LoadVersionAndUpgrade(latestVersion, upgrades)
	s.Require().NoError(err)

	// GetProof should work for the old stores
	for _, storeKey := range []string{storeKey3} {
		for i := uint64(1); i <= latestVersion; i++ {
			for j := 0; j < kvCount; j++ {
				proof, err := commitStore.GetProof([]byte(storeKey), i, []byte(fmt.Sprintf("key-%d-%d", i, j)))
				s.Require().NoError(err)
				s.Require().NotNil(proof)
			}
		}
	}
	// GetProof should fail for the new stores against the old versions
	for _, storeKey := range []string{"newStore1", "newStore2"} {
		for i := uint64(1); i <= latestVersion; i++ {
			for j := 0; j < kvCount; j++ {
				_, err := commitStore.GetProof([]byte(storeKey), i, []byte(fmt.Sprintf("key-%d-%d", i, j)))
				s.Require().Error(err)
			}
		}
	}

	// apply the changeset again
	for i := latestVersion + 1; i < latestVersion*2; i++ {
		kvPairs := make(map[string]corestore.KVPairs)
		for _, storeKey := range realStoreKeys {
			kvPairs[storeKey] = corestore.KVPairs{}
			for j := 0; j < kvCount; j++ {
				key := []byte(fmt.Sprintf("key-%d-%d", i, j))
				value := []byte(fmt.Sprintf("value-%d-%d", i, j))
				kvPairs[storeKey] = append(kvPairs[storeKey], corestore.KVPair{Key: key, Value: value})
			}
		}
		s.Require().NoError(commitStore.WriteChangeset(corestore.NewChangesetWithPairs(i, kvPairs)))
		commitInfo, err := commitStore.Commit(i)
		s.Require().NoError(err)
		s.Require().NotNil(commitInfo)
		s.Require().Equal(len(realStoreKeys), len(commitInfo.StoreInfos))
		for _, storeKey := range realStoreKeys {
			s.Require().NotNil(commitInfo.GetStoreCommitID([]byte(storeKey)))
		}
	}

	// verify new stores
	for _, storeKey := range []string{"newStore1", "newStore2"} {
		for i := latestVersion + 1; i < latestVersion*2; i++ {
			for j := 0; j < kvCount; j++ {
				proof, err := commitStore.GetProof([]byte(storeKey), i, []byte(fmt.Sprintf("key-%d-%d", i, j)))
				s.Require().NoError(err)
				s.Require().NotNil(proof)
			}
		}
	}

	// verify existing store
	for i := uint64(1); i < latestVersion*2; i++ {
		for j := 0; j < kvCount; j++ {
			prf, err := commitStore.GetProof([]byte(storeKey2), i, []byte(fmt.Sprintf("key-%d-%d", i, j)))
			s.Require().NoError(err)
			s.Require().NotNil(prf)
		}
	}

	// create a new commitment store with one more upgrades
	upgrades = &corestore.StoreUpgrades{
		Deleted: []string{storeKey2},
		Added:   []string{"newStore3"},
	}
	newRealStoreKeys := []string{storeKey1, "newStore1", "newStore2", "newStore3"}
	oldStoreKeys = []string{storeKey2, storeKey3}
	commitStore, err = s.NewStore(commitDB, commitDir, newRealStoreKeys, oldStoreKeys, coretesting.NewNopLogger())
	s.Require().NoError(err)
	err = commitStore.LoadVersionAndUpgrade(2*latestVersion-1, upgrades)
	s.Require().NoError(err)

	// apply the changeset again
	for i := latestVersion * 2; i < latestVersion*3; i++ {
		kvPairs := make(map[string]corestore.KVPairs)
		for _, storeKey := range newRealStoreKeys {
			kvPairs[storeKey] = corestore.KVPairs{}
			for j := 0; j < kvCount; j++ {
				key := []byte(fmt.Sprintf("key-%d-%d", i, j))
				value := []byte(fmt.Sprintf("value-%d-%d", i, j))
				kvPairs[storeKey] = append(kvPairs[storeKey], corestore.KVPair{Key: key, Value: value})
			}
		}
		err = commitStore.WriteChangeset(corestore.NewChangesetWithPairs(i, kvPairs))
		s.Require().NoError(err)
		commitInfo, err := commitStore.Commit(i)
		s.Require().NoError(err)
		s.Require().NotNil(commitInfo)
		s.Require().Equal(len(newRealStoreKeys), len(commitInfo.StoreInfos))
		for _, storeKey := range newRealStoreKeys {
			s.Require().NotNil(commitInfo.GetStoreCommitID([]byte(storeKey)))
		}
	}

	// prune the old stores
	s.Require().NoError(commitStore.Prune(latestVersion))
	s.T().Logf("prune to version %d", latestVersion)
	// GetProof should fail for the old stores
	for _, storeKey := range []string{storeKey1, storeKey3} {
		for i := uint64(1); i <= latestVersion; i++ {
			for j := 0; j < kvCount; j++ {
				_, err := commitStore.GetProof([]byte(storeKey), i, []byte(fmt.Sprintf("key-%d-%d", i, j)))
				s.Require().Error(err)
			}
		}
	}
	s.T().Log("GetProof should work for the new stores")
	// GetProof should not fail for the newly removed store
	for i := latestVersion + 1; i < latestVersion*2; i++ {
		for j := 0; j < kvCount; j++ {
			proof, err := commitStore.GetProof([]byte(storeKey2), i, []byte(fmt.Sprintf("key-%d-%d", i, j)))
			s.Require().NoError(err)
			s.Require().NotNil(proof)
		}
	}

	s.T().Logf("Prune to version %d", latestVersion*2)
	s.Require().NoError(commitStore.Prune(latestVersion * 2))
	// GetProof should fail for the newly deleted stores
	for i := uint64(1); i < latestVersion*2; i++ {
		for j := 0; j < kvCount; j++ {
			_, err := commitStore.GetProof([]byte(storeKey2), i, []byte(fmt.Sprintf("key-%d-%d", i, j)))
			s.Require().Error(err)
		}
	}
	s.T().Log("GetProof should work for the new added store")
	// GetProof should work for the new added store
	for i := latestVersion*2 + 1; i < latestVersion*3; i++ {
		for j := 0; j < kvCount; j++ {
			proof, err := commitStore.GetProof([]byte("newStore3"), i, []byte(fmt.Sprintf("key-%d-%d", i, j)))
			s.Require().NoError(err)
			s.Require().NotNil(proof)
		}
	}
}
