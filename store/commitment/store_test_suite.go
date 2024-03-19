package commitment

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/stretchr/testify/suite"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/snapshots"
	snapshotstypes "cosmossdk.io/store/v2/snapshots/types"
)

const (
	storeKey1 = "store1"
	storeKey2 = "store2"
)

// CommitStoreTestSuite is a test suite to be used for all tree backends.
type CommitStoreTestSuite struct {
	suite.Suite

	NewStore func(db store.RawDB, storeKeys []string, pruneOpts *store.PruneOptions, logger log.Logger) (*CommitStore, error)
}

func (s *CommitStoreTestSuite) TestStore_Snapshotter() {
	storeKeys := []string{storeKey1, storeKey2}
	commitStore, err := s.NewStore(dbm.NewMemDB(), storeKeys, nil, log.NewNopLogger())
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
		s.Require().NoError(commitStore.WriteBatch(corestore.NewChangesetWithPairs(kvPairs)))

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

	targetStore, err := s.NewStore(dbm.NewMemDB(), storeKeys, nil, log.NewNopLogger())
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

func (s *CommitStoreTestSuite) TestStore_Pruning() {
	storeKeys := []string{storeKey1, storeKey2}
	pruneOpts := &store.PruneOptions{
		KeepRecent: 10,
		Interval:   5,
	}
	commitStore, err := s.NewStore(dbm.NewMemDB(), storeKeys, pruneOpts, log.NewNopLogger())
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
		s.Require().NoError(commitStore.WriteBatch(corestore.NewChangesetWithPairs(kvPairs)))

		_, err = commitStore.Commit(i)
		s.Require().NoError(err)
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
