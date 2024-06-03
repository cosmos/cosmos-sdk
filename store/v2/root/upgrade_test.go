package root

import (
	"fmt"
	"testing"

	corelog "cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/pruning"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/sqlite"
	"github.com/stretchr/testify/suite"
)

type UpgradeStoreTestSuite struct {
	suite.Suite

	commitDB  corestore.KVStoreWithBatch
	rootStore store.RootStore
}

func TestUpgradeStoreTestSuite(t *testing.T) {
	suite.Run(t, &MigrateStoreTestSuite{})
}

func (s *UpgradeStoreTestSuite) SetupTest() {
	testLog := log.NewTestLogger(s.T())
	nopLog := corelog.NewNopLogger()

	s.commitDB = dbm.NewMemDB()
	multiTrees := make(map[string]commitment.Tree)
	for _, storeKey := range storeKeys {
		prefixDB := dbm.NewPrefixDB(s.commitDB, []byte(storeKey))
		multiTrees[storeKey] = iavl.NewIavlTree(prefixDB, nopLog, iavl.DefaultConfig())
	}

	// create storage and commitment stores
	sqliteDB, err := sqlite.New(s.T().TempDir())
	s.Require().NoError(err)
	ss := storage.NewStorageStore(sqliteDB, testLog)
	sc, err := commitment.NewCommitStore(multiTrees, s.commitDB, testLog)
	s.Require().NoError(err)
	pm := pruning.NewManager(sc, ss, nil, nil)
	s.rootStore, err = New(testLog, ss, sc, pm, nil, nil)
	s.Require().NoError(err)

	// commit changeset
	toVersion := uint64(20)
	keyCount := 10
	for version := uint64(1); version <= toVersion; version++ {
		cs := corestore.NewChangeset()
		for _, storeKey := range storeKeys {
			for i := 0; i < keyCount; i++ {
				cs.Add([]byte(storeKey), []byte(fmt.Sprintf("key-%d-%d", version, i)), []byte(fmt.Sprintf("value-%d-%d", version, i)), false)
			}
		}
		_, err = s.rootStore.Commit(cs)
		s.Require().NoError(err)
	}
}

func (s *UpgradeStoreTestSuite) loadWithUpgrades(upgrades *corestore.StoreUpgrades) {
	testLog := log.NewTestLogger(s.T())
	nopLog := corelog.NewNopLogger()

	// create a new commitment store
	multiTrees := make(map[string]commitment.Tree)
	for _, storeKey := range storeKeys {
		prefixDB := dbm.NewPrefixDB(s.commitDB, []byte(storeKey))
		multiTrees[storeKey] = iavl.NewIavlTree(prefixDB, nopLog, iavl.DefaultConfig())
	}
	for _, added := range upgrades.Added {
		prefixDB := dbm.NewPrefixDB(s.commitDB, []byte(added))
		multiTrees[added] = iavl.NewIavlTree(prefixDB, nopLog, iavl.DefaultConfig())
	}
	for _, rename := range upgrades.Renamed {
		prefixDB := dbm.NewPrefixDB(s.commitDB, []byte(rename.OldKey))
		multiTrees[rename.NewKey] = iavl.NewIavlTree(prefixDB, nopLog, iavl.DefaultConfig())
	}

	sc, err := commitment.NewCommitStore(multiTrees, s.commitDB, testLog)
	s.Require().NoError(err)
	pm := pruning.NewManager(sc, s.rootStore.GetStateStorage().(store.Pruner), nil, nil)
	s.rootStore, err = New(testLog, s.rootStore.GetStateStorage(), sc, pm, nil, nil)
	s.Require().NoError(err)
}

func (s *UpgradeStoreTestSuite) TestLoadVersionAndUpgrade() {
	// upgrade store keys
	upgrades := &corestore.StoreUpgrades{
		Added: []string{"newStore1", "newStore2"},
		Renamed: []corestore.StoreRename{
			{OldKey: "store1", NewKey: "renamedStore1"},
			{OldKey: "store2", NewKey: "renamedStore2"},
		},
		Deleted: []string{"store3"},
	}
	s.loadWithUpgrades(upgrades)

	// load the store with the upgrades
	v, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	err = s.rootStore.(store.UpgradeableRootStore).LoadVersionAndUpgrade(v, upgrades)
	s.Require().NoError(err)

	// commit changeset
	newStoreKeys := []string{"newStore1", "newStore2", "renamedStore1", "renamedStore2"}
	toVersion := uint64(40)
	keyCount := 10
	for version := v + 1; version <= toVersion; version++ {
		cs := corestore.NewChangeset()
		for _, storeKey := range newStoreKeys {
			for i := 0; i < keyCount; i++ {
				cs.Add([]byte(storeKey), []byte(fmt.Sprintf("key-%d-%d", version, i)), []byte(fmt.Sprintf("value-%d-%d", version, i)), false)
			}
		}
		_, err = s.rootStore.Commit(cs)
		s.Require().NoError(err)
	}
}
