package root

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/pruning"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/sqlite"
)

type UpgradeStoreTestSuite struct {
	suite.Suite

	commitDB  corestore.KVStoreWithBatch
	rootStore store.RootStore
}

func TestUpgradeStoreTestSuite(t *testing.T) {
	suite.Run(t, &UpgradeStoreTestSuite{})
}

func (s *UpgradeStoreTestSuite) SetupTest() {
	testLog := log.NewTestLogger(s.T())
	nopLog := coretesting.NewNopLogger()

	s.commitDB = dbm.NewMemDB()
	multiTrees := make(map[string]commitment.Tree)
	newTreeFn := func(storeKey string) (commitment.Tree, error) {
		prefixDB := dbm.NewPrefixDB(s.commitDB, []byte(storeKey))
		return iavl.NewIavlTree(prefixDB, nopLog, iavl.DefaultConfig()), nil
	}
	for _, storeKey := range storeKeys {
		multiTrees[storeKey], _ = newTreeFn(storeKey)
	}

	// create storage and commitment stores
	sqliteDB, err := sqlite.New(s.T().TempDir())
	s.Require().NoError(err)
	ss := storage.NewStorageStore(sqliteDB, testLog)
	sc, err := commitment.NewCommitStore(multiTrees, nil, s.commitDB, testLog)
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
	nopLog := coretesting.NewNopLogger()

	// create a new commitment store
	multiTrees := make(map[string]commitment.Tree)
	oldTrees := make(map[string]commitment.Tree)
	newTreeFn := func(storeKey string) (commitment.Tree, error) {
		prefixDB := dbm.NewPrefixDB(s.commitDB, []byte(storeKey))
		return iavl.NewIavlTree(prefixDB, nopLog, iavl.DefaultConfig()), nil
	}
	for _, storeKey := range storeKeys {
		multiTrees[storeKey], _ = newTreeFn(storeKey)
	}
	for _, added := range upgrades.Added {
		multiTrees[added], _ = newTreeFn(added)
	}
	for _, deleted := range upgrades.Deleted {
		oldTrees[deleted], _ = newTreeFn(deleted)
	}

	sc, err := commitment.NewCommitStore(multiTrees, oldTrees, s.commitDB, testLog)
	s.Require().NoError(err)
	pm := pruning.NewManager(sc, s.rootStore.GetStateStorage().(store.Pruner), nil, nil)
	s.rootStore, err = New(testLog, s.rootStore.GetStateStorage(), sc, pm, nil, nil)
	s.Require().NoError(err)
}

func (s *UpgradeStoreTestSuite) TestLoadVersionAndUpgrade() {
	// upgrade store keys
	upgrades := &corestore.StoreUpgrades{
		Added:   []string{"newStore1", "newStore2"},
		Deleted: []string{"store3"},
	}
	s.loadWithUpgrades(upgrades)

	// load the store with the upgrades
	v, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	err = s.rootStore.(store.UpgradeableStore).LoadVersionAndUpgrade(v, upgrades)
	s.Require().NoError(err)

	keyCount := 10
	// check old store keys are queryable
	oldStoreKeys := []string{"store1", "store3"}
	for _, storeKey := range oldStoreKeys {
		for version := uint64(1); version <= v; version++ {
			for i := 0; i < keyCount; i++ {
				proof, err := s.rootStore.Query([]byte(storeKey), version, []byte(fmt.Sprintf("key-%d-%d", version, i)), true)
				s.Require().NoError(err)
				s.Require().NotNil(proof)
			}
		}
	}

	// commit changeset
	newStoreKeys := []string{"newStore1", "newStore2"}
	toVersion := uint64(40)
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

	// check new store keys are queryable
	for _, storeKey := range newStoreKeys {
		for version := v + 1; version <= toVersion; version++ {
			for i := 0; i < keyCount; i++ {
				_, err := s.rootStore.Query([]byte(storeKey), version, []byte(fmt.Sprintf("key-%d-%d", version, i)), true)
				s.Require().NoError(err)
			}
		}
	}

	// check the original store key is queryable
	for version := uint64(1); version <= toVersion; version++ {
		for i := 0; i < keyCount; i++ {
			_, err := s.rootStore.Query([]byte("store2"), version, []byte(fmt.Sprintf("key-%d-%d", version, i)), true)
			s.Require().NoError(err)
		}
	}
}
