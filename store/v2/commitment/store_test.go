package commitment_test

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"testing"

	"cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/snapshots"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
	"github.com/stretchr/testify/suite"
)

var (
	storeKey1 = "store1"
	storeKey2 = "store2"
	storeKeys = []string{storeKey1, storeKey2}
)

func SetupTest() *commitment.CommitStoreTestSuite {
	s := &commitment.CommitStoreTestSuite{
		NewStore: func(db corestore.KVStoreWithBatch, storeKeys []string, logger log.Logger) (*commitment.CommitStore, error) {
			multiTrees := make(map[string]commitment.Tree)
			for _, storeKey := range []string{storeKey1, storeKey2} {
				prefixDB := dbm.NewPrefixDB(db, []byte(storeKey))
				multiTrees[storeKey] = iavl.NewIavlTree(prefixDB, coretesting.NewNopLogger(), iavl.DefaultConfig())
			}
			sc, err := commitment.NewCommitStore(multiTrees, db, logger)
			return sc, err
		},
	}

	return s
}

func TestStore_Snapshotter(t *testing.T) {
	s := SetupTest()
	suite.Run(t, s)

	commitStore, err := s.NewStore(dbm.NewMemDB(), storeKeys, coretesting.NewNopLogger())
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
		s.Require().NoError(commitStore.WriteChangeset(corestore.NewChangesetWithPairs(kvPairs)))

		_, err = commitStore.Commit(i)
		s.Require().NoError(err)
	}

	cInfo := commitStore.WorkingCommitInfo(latestVersion)
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

	targetStore, err := s.NewStore(dbm.NewMemDB(), storeKeys, coretesting.NewNopLogger())
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
	chStorage := make(chan *corestore.StateChanges, 100)
	leaves := make(map[string]string)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for kv := range chStorage {
			for _, actor := range kv.StateChanges {
				leaves[fmt.Sprintf("%s_%s", kv.Actor, actor.Key)] = string(actor.Value)
			}
		}
		wg.Done()
	}()
	nextItem, err := targetStore.Restore(latestVersion, snapshotstypes.CurrentFormat, streamReader, chStorage)
	s.Require().NoError(err)
	s.Require().Equal(*dummyExtensionItem.GetExtension(), *nextItem.GetExtension())

	close(chStorage)
	wg.Wait()
	s.Require().Equal(len(storeKeys)*kvCount*int(latestVersion), len(leaves))
	for _, storeKey := range storeKeys {
		for i := 1; i <= int(latestVersion); i++ {
			for j := 0; j < kvCount; j++ {
				key := fmt.Sprintf("%s_key-%d-%d", storeKey, i, j)
				s.Require().Equal(leaves[key], fmt.Sprintf("value-%d-%d", i, j))
			}
		}
	}

	// check the restored tree hash
	targetCommitInfo := targetStore.WorkingCommitInfo(latestVersion)
	for _, storeInfo := range targetCommitInfo.StoreInfos {
		matched := false
		for _, latestStoreInfo := range cInfo.StoreInfos {
			if bytes.Equal(storeInfo.Name, latestStoreInfo.Name) {
				s.Require().Equal(latestStoreInfo.GetHash(), storeInfo.GetHash())
				matched = true
			}
		}
		s.Require().True(matched)
	}
}

func TestStore_Pruning(t *testing.T) {
	pruneOpts := store.NewPruningOptionWithCustom(10, 5)
	s := SetupTest()
	suite.Run(t, s)

	commitStore, err := s.NewStore(dbm.NewMemDB(), storeKeys, coretesting.NewNopLogger())
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
		s.Require().NoError(commitStore.WriteChangeset(corestore.NewChangesetWithPairs(kvPairs)))

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

func TestStore_GetProof(t *testing.T) {
	s := SetupTest()
	suite.Run(t, s)

	commitStore, err := s.NewStore(dbm.NewMemDB(), storeKeys, coretesting.NewNopLogger())
	s.Require().NoError(err)

	toVersion := uint64(10)
	keyCount := 5

	// commit some changes
	for version := uint64(1); version <= toVersion; version++ {
		cs := corestore.NewChangeset()
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

func TestStore_Get(t *testing.T) {
	s := SetupTest()
	suite.Run(t, s)

	commitStore, err := s.NewStore(dbm.NewMemDB(), storeKeys, coretesting.NewNopLogger())
	s.Require().NoError(err)

	toVersion := uint64(10)
	keyCount := 5

	// commit some changes
	for version := uint64(1); version <= toVersion; version++ {
		cs := corestore.NewChangeset()
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
