package pruning

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/sqlite"
)

var storeKeys = []string{"store1", "store2", "store3"}

type PruningManagerTestSuite struct {
	suite.Suite

	manager *Manager
	sc      *commitment.CommitStore
	ss      *storage.StorageStore
}

func TestPruningManagerTestSuite(t *testing.T) {
	suite.Run(t, &PruningManagerTestSuite{})
}

func (s *PruningManagerTestSuite) SetupTest() {
	nopLog := log.NewNopLogger()
	var err error

	mdb := dbm.NewMemDB()
	multiTrees := make(map[string]commitment.Tree)
	for _, storeKey := range storeKeys {
		prefixDB := dbm.NewPrefixDB(mdb, []byte(storeKey))
		multiTrees[storeKey] = iavl.NewIavlTree(prefixDB, nopLog, iavl.DefaultConfig())
	}
	s.sc, err = commitment.NewCommitStore(multiTrees, mdb, nopLog)
	s.Require().NoError(err)

	sqliteDB, err := sqlite.New(s.T().TempDir())
	s.Require().NoError(err)
	s.ss = storage.NewStorageStore(sqliteDB, nopLog)
	scPruneOptions := &store.PruneOptions{
		KeepRecent: 0,
		Interval:   1,
	} // prune all
	ssPruneOptions := &store.PruneOptions{
		KeepRecent: 5,
		Interval:   10,
	} // prune some
	s.manager = NewManager(s.sc, s.ss, scPruneOptions, ssPruneOptions)
}

func (s *PruningManagerTestSuite) TestPrune() {
	// commit changesets with pruning
	toVersion := uint64(100)
	keyCount := 10
	for version := uint64(1); version <= toVersion; version++ {
		cs := corestore.NewChangeset()
		for _, storeKey := range storeKeys {
			for i := 0; i < keyCount; i++ {
				cs.Add([]byte(storeKey), []byte(fmt.Sprintf("key-%d-%d", version, i)), []byte(fmt.Sprintf("value-%d-%d", version, i)), false)
			}
		}
		s.Require().NoError(s.sc.WriteChangeset(cs))
		_, err := s.sc.Commit(version)
		s.Require().NoError(err)

		s.Require().NoError(s.ss.ApplyChangeset(version, cs))

		s.Require().NoError(s.manager.Prune(version))
	}

	// wait for the pruning to finish in the commitment store
	checkSCPrune := func() bool {
		count := 0
		for _, storeKey := range storeKeys {
			_, err := s.sc.GetProof([]byte(storeKey), toVersion-1, []byte(fmt.Sprintf("key-%d-%d", toVersion-1, 0)))
			if err != nil {
				count++
			}
		}

		return count == len(storeKeys)
	}
	s.Require().Eventually(checkSCPrune, 10*time.Second, 1*time.Second)

	// check the storage store
	_, pruneVersion := s.manager.ssPruningOptions.ShouldPrune(toVersion)
	for version := uint64(1); version <= toVersion; version++ {
		for _, storeKey := range storeKeys {
			for i := 0; i < keyCount; i++ {
				key := []byte(fmt.Sprintf("key-%d-%d", version, i))
				value, err := s.ss.Get([]byte(storeKey), version, key)
				if version <= pruneVersion {
					s.Require().Nil(value)
					s.Require().Error(err)
				} else {
					s.Require().NoError(err)
					s.Require().Equal([]byte(fmt.Sprintf("value-%d-%d", version, i)), value)
				}
			}
		}
	}
}

func TestPruneOptions(t *testing.T) {
	testCases := []struct {
		name         string
		options      *store.PruneOptions
		version      uint64
		pruning      bool
		pruneVersion uint64
	}{
		{
			name: "no pruning",
			options: &store.PruneOptions{
				KeepRecent: 100,
				Interval:   0,
			},
			version:      100,
			pruning:      false,
			pruneVersion: 0,
		},
		{
			name: "prune all",
			options: &store.PruneOptions{
				KeepRecent: 0,
				Interval:   1,
			},
			version:      19,
			pruning:      true,
			pruneVersion: 18,
		},
		{
			name: "prune none",
			options: &store.PruneOptions{
				KeepRecent: 100,
				Interval:   10,
			},
			version:      19,
			pruning:      false,
			pruneVersion: 0,
		},
		{
			name: "prune some",
			options: &store.PruneOptions{
				KeepRecent: 10,
				Interval:   50,
			},
			version:      100,
			pruning:      true,
			pruneVersion: 89,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pruning, pruneVersion := tc.options.ShouldPrune(tc.version)
			require.Equal(t, tc.pruning, pruning)
			require.Equal(t, tc.pruneVersion, pruneVersion)
		})
	}
}
