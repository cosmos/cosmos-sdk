package commitment

import (
	"fmt"
	"io"
	"sync"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
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

	NewStore func(db dbm.DB, storeKeys []string, logger log.Logger) (*CommitStore, error)
}

func (s *CommitStoreTestSuite) TestSnapshotter() {
	storeKeys := []string{storeKey1, storeKey2}
	commitStore, err := s.NewStore(dbm.NewMemDB(), storeKeys, log.NewNopLogger())
	s.Require().NoError(err)

	latestVersion := uint64(10)
	kvCount := 10
	for i := uint64(1); i <= latestVersion; i++ {
		kvPairs := make(map[string]store.KVPairs)
		for _, storeKey := range storeKeys {
			kvPairs[storeKey] = store.KVPairs{}
			for j := 0; j < kvCount; j++ {
				key := []byte(fmt.Sprintf("key-%d-%d", i, j))
				value := []byte(fmt.Sprintf("value-%d-%d", i, j))
				kvPairs[storeKey] = append(kvPairs[storeKey], store.KVPair{Key: key, Value: value})
			}
		}
		s.Require().NoError(commitStore.WriteBatch(store.NewChangeset(kvPairs)))

		_, err = commitStore.Commit()
		s.Require().NoError(err)
	}

	latestStoreInfos := commitStore.WorkingStoreInfos(latestVersion)
	s.Require().Equal(len(storeKeys), len(latestStoreInfos))

	// create a snapshot
	dummyExtensionItem := snapshotstypes.SnapshotItem{
		Item: &snapshotstypes.SnapshotItem_Extension{
			Extension: &snapshotstypes.SnapshotExtensionMeta{
				Name:   "test",
				Format: 1,
			},
		},
	}

	targetStore, err := s.NewStore(dbm.NewMemDB(), storeKeys, log.NewNopLogger())
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
	chStorage := make(chan *store.KVPair, 100)
	leaves := make(map[string]string)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for kv := range chStorage {
			leaves[fmt.Sprintf("%s_%s", kv.StoreKey, kv.Key)] = string(kv.Value)
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
	targetStoreInfos := targetStore.WorkingStoreInfos(latestVersion)
	s.Require().Equal(len(storeKeys), len(targetStoreInfos))
	for _, storeInfo := range targetStoreInfos {
		matched := false
		for _, latestStoreInfo := range latestStoreInfos {
			if storeInfo.Name == latestStoreInfo.Name {
				s.Require().Equal(latestStoreInfo.GetHash(), storeInfo.GetHash())
				matched = true
			}
		}
		s.Require().True(matched)
	}
}
