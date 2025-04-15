package rootmulti_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	"cosmossdk.io/store/memiavl/rootmulti"
	"cosmossdk.io/store/types"
)

type StoreTestSuite struct {
	suite.Suite
	store *rootmulti.Store
	dir   string
}

func (s *StoreTestSuite) SetupTest() {
	dir := s.Suite.T().TempDir()
	s.dir = dir
	s.store = rootmulti.NewStore(dir, log.NewNopLogger(), false)
}

func (s *StoreTestSuite) TearDownTest() {
	if s.store != nil {
		err := s.store.Close()
		// Log error if closing fails, but proceed with cleanup regardless
		if err != nil {
			s.T().Logf("Error closing store during teardown: %v", err)
		}
	}
	err := os.RemoveAll(s.dir)
	s.Require().NoError(err)
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (s *StoreTestSuite) TestLoadCommit() {
	require := s.Require()

	// Define store keys
	key1 := types.NewKVStoreKey("store1")
	key2 := types.NewKVStoreKey("store2")
	tkey := types.NewTransientStoreKey("tstore")
	memKey := types.NewMemoryStoreKey("memstore")

	// Mount stores
	s.store.MountStoreWithDB(key1, types.StoreTypeIAVL, nil)
	s.store.MountStoreWithDB(key2, types.StoreTypeIAVL, nil)
	s.store.MountStoreWithDB(tkey, types.StoreTypeTransient, nil)
	s.store.MountStoreWithDB(memKey, types.StoreTypeMemory, nil)

	// Load latest version (should be 0)
	err := s.store.LoadLatestVersion()
	require.NoError(err)
	require.Equal(int64(0), s.store.LastCommitID().Version)

	// --- Test commit and load ---
	// Write some data
	kvStore1 := s.store.GetKVStore(key1)
	kvStore1.Set([]byte("key1_1"), []byte("value1_1"))
	kvStore2 := s.store.GetKVStore(key2)
	kvStore2.Set([]byte("key2_1"), []byte("value2_1"))
	tStore := s.store.GetKVStore(tkey)
	tStore.Set([]byte("tkey1"), []byte("tvalue1")) // Transient data
	memStore := s.store.GetKVStore(memKey)
	memStore.Set([]byte("memkey1"), []byte("memvalue1")) // Memory data

	// Commit
	commitID := s.store.Commit()
	require.Equal(int64(1), commitID.Version)
	require.NotEmpty(commitID.Hash)

	lastCommitID := s.store.LastCommitID()
	require.Equal(commitID, lastCommitID)

	// Close the current store
	err = s.store.Close()
	require.NoError(err)

	// Create a new store instance and load the last version
	s.store = rootmulti.NewStore(s.dir, log.NewNopLogger(), false)
	s.store.MountStoreWithDB(key1, types.StoreTypeIAVL, nil)
	s.store.MountStoreWithDB(key2, types.StoreTypeIAVL, nil)
	s.store.MountStoreWithDB(tkey, types.StoreTypeTransient, nil) // Need to mount transient/memory again
	s.store.MountStoreWithDB(memKey, types.StoreTypeMemory, nil)

	err = s.store.LoadLatestVersion()
	require.NoError(err)
	require.Equal(int64(1), s.store.LastCommitID().Version)
	require.Equal(commitID.Hash, s.store.LastCommitID().Hash)

	// Check data persistence for IAVL stores
	kvStore1 = s.store.GetKVStore(key1)
	require.Equal([]byte("value1_1"), kvStore1.Get([]byte("key1_1")))
	kvStore2 = s.store.GetKVStore(key2)
	require.Equal([]byte("value2_1"), kvStore2.Get([]byte("key2_1")))

	// Check data non-persistence for Transient and Memory stores
	tStore = s.store.GetKVStore(tkey)
	require.Nil(tStore.Get([]byte("tkey1")))
	memStore = s.store.GetKVStore(memKey)
	require.Nil(memStore.Get([]byte("memkey1")))

	// --- Test loading specific version ---
	// Write more data and commit again
	kvStore1.Set([]byte("key1_2"), []byte("value1_2"))
	commitID2 := s.store.Commit()
	require.Equal(int64(2), commitID2.Version)

	err = s.store.Close()
	require.NoError(err)

	// Load version 1
	s.store = rootmulti.NewStore(s.dir, log.NewNopLogger(), false)
	s.store.MountStoreWithDB(key1, types.StoreTypeIAVL, nil)
	s.store.MountStoreWithDB(key2, types.StoreTypeIAVL, nil)
	err = s.store.LoadVersion(1)
	require.NoError(err)
	require.Equal(int64(1), s.store.LastCommitID().Version) // LastCommitID reflects the loaded version's commit

	// Check data at version 1
	kvStore1 = s.store.GetKVStore(key1)
	require.Equal([]byte("value1_1"), kvStore1.Get([]byte("key1_1")))
	require.Nil(kvStore1.Get([]byte("key1_2"))) // key1_2 should not exist at version 1
	kvStore2 = s.store.GetKVStore(key2)
	require.Equal([]byte("value2_1"), kvStore2.Get([]byte("key2_1")))
}

func (s *StoreTestSuite) TestWorkingHash() {
	require := s.Require()

	key1 := types.NewKVStoreKey("store1")
	s.store.MountStoreWithDB(key1, types.StoreTypeIAVL, nil)
	err := s.store.LoadLatestVersion()
	require.NoError(err)

	// Initial working hash
	initialHash := s.store.WorkingHash()
	require.NotEmpty(initialHash)

	// Write data, working hash should change
	kvStore1 := s.store.GetKVStore(key1)
	kvStore1.Set([]byte("key"), []byte("value"))
	workingHash1 := s.store.WorkingHash()
	require.NotEmpty(workingHash1)
	require.NotEqual(initialHash, workingHash1)

	// Commit, working hash should match commit hash
	commitID := s.store.Commit()
	require.Equal(int64(1), commitID.Version)
	workingHashAfterCommit := s.store.WorkingHash()
	require.Equal(commitID.Hash, workingHashAfterCommit)

	// Write more data, working hash changes again
	kvStore1.Set([]byte("key2"), []byte("value2"))
	workingHash2 := s.store.WorkingHash()
	require.NotEmpty(workingHash2)
	require.NotEqual(workingHashAfterCommit, workingHash2)
}

func (s *StoreTestSuite) TestGetStores() {
	require := s.Require()

	key1 := types.NewKVStoreKey("store1")
	tkey := types.NewTransientStoreKey("tstore")
	memKey := types.NewMemoryStoreKey("memstore")
	invalidKey := types.NewKVStoreKey("invalid")

	s.store.MountStoreWithDB(key1, types.StoreTypeIAVL, nil)
	s.store.MountStoreWithDB(tkey, types.StoreTypeTransient, nil)
	s.store.MountStoreWithDB(memKey, types.StoreTypeMemory, nil)

	err := s.store.LoadLatestVersion()
	require.NoError(err)

	// Get existing stores
	store1 := s.store.GetStore(key1)
	require.NotNil(store1)
	require.Equal(types.StoreTypeIAVL, store1.GetStoreType())

	kvStore1 := s.store.GetKVStore(key1)
	require.NotNil(kvStore1)

	commitStore1 := s.store.GetCommitStore(key1)
	require.NotNil(commitStore1)

	commitKVStore1 := s.store.GetCommitKVStore(key1)
	require.NotNil(commitKVStore1)

	tStore := s.store.GetStore(tkey)
	require.NotNil(tStore)
	require.Equal(types.StoreTypeTransient, tStore.GetStoreType())

	memStore := s.store.GetStore(memKey)
	require.NotNil(memStore)
	require.Equal(types.StoreTypeMemory, memStore.GetStoreType())

	// Get store by name
	storeByName := s.store.GetStoreByName("store1")
	require.NotNil(storeByName)
	require.Equal(key1.Name(), "store1") // GetStoreKey is on the Store interface

	storeByNameInvalid := s.store.GetStoreByName("invalid")
	require.Nil(storeByNameInvalid)

	// Panics for invalid gets
	require.Panics(func() { s.store.GetStore(invalidKey) })
	require.Panics(func() { s.store.GetKVStore(invalidKey) })
	require.Panics(func() { s.store.GetCommitStore(invalidKey) })
	require.Panics(func() { s.store.GetCommitKVStore(invalidKey) })
}

func (s *StoreTestSuite) TestCacheMultiStore() {
	require := s.Require()

	key1 := types.NewKVStoreKey("store1")
	s.store.MountStoreWithDB(key1, types.StoreTypeIAVL, nil)
	err := s.store.LoadLatestVersion()
	require.NoError(err)

	kvStore1 := s.store.GetKVStore(key1)
	kvStore1.Set([]byte("key"), []byte("value"))
	s.store.Commit() // v1

	// Get cache multi store at latest version (v1)
	cms, err := s.store.CacheMultiStoreWithVersion(1)
	require.NoError(err)
	require.NotNil(cms)

	cacheKVStore1 := cms.GetKVStore(key1)
	require.Equal([]byte("value"), cacheKVStore1.Get([]byte("key")))

	// Write to cache store
	cacheKVStore1.Set([]byte("key"), []byte("new_value_cache"))
	cacheKVStore1.Set([]byte("key_cache"), []byte("value_cache"))

	// Check cache store reflects changes
	require.Equal([]byte("new_value_cache"), cacheKVStore1.Get([]byte("key")))
	require.Equal([]byte("value_cache"), cacheKVStore1.Get([]byte("key_cache")))

	// Check original store is unchanged
	origKVStore1 := s.store.GetKVStore(key1)
	require.Equal([]byte("value"), origKVStore1.Get([]byte("key")))
	require.Nil(origKVStore1.Get([]byte("key_cache")))

	// Write changes from cache
	cms.Write()

	// Test CacheWrap
	cw := s.store.CacheWrap()
	require.NotNil(cw)
	// Further CacheWrap tests could involve checking trace/listeners if implemented fully
}

func (s *StoreTestSuite) TestQuery() {
	require := s.Require()

	key1 := types.NewKVStoreKey("store1")
	s.store.MountStoreWithDB(key1, types.StoreTypeIAVL, nil)
	err := s.store.LoadLatestVersion()
	require.NoError(err)

	// Commit v1
	kvStore1 := s.store.GetKVStore(key1)
	kvStore1.Set([]byte("querykey1"), []byte("value1"))
	commitID1 := s.store.Commit()
	require.Equal(int64(1), commitID1.Version)

	// Commit v2
	kvStore1.Set([]byte("querykey2"), []byte("value2"))
	commitID2 := s.store.Commit()
	require.Equal(int64(2), commitID2.Version)

	// Query at height 1 (non-proof)
	queryPath := "/store1/key" // using store name
	req := types.RequestQuery{
		Path:   queryPath,
		Data:   []byte("querykey1"), // Data field is often used for the key in KVStore queries
		Height: 1,
		Prove:  false,
	}
	res, err := s.store.Query(&req)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(uint32(0), res.Code) // Code 0 indicates success
	require.Equal(int64(1), res.Height)
	require.Equal([]byte("value1"), res.Value)
	require.Nil(res.ProofOps) // No proof requested

	// Query non-existent key at height 1
	req.Path = "/store1/key"
	req.Data = []byte("nonexistent")
	res, err = s.store.Query(&req)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(uint32(0), res.Code)
	require.Equal(int64(1), res.Height)
	require.Nil(res.Value) // Value should be nil for non-existent key

	// Query key added in v2 at height 1 (should not exist)
	req.Path = "/store1/key"
	req.Data = []byte("querykey2")
	res, err = s.store.Query(&req)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(uint32(0), res.Code)
	require.Equal(int64(1), res.Height)
	require.Nil(res.Value)
}

func (s *StoreTestSuite) TestUpgrades() {
	require := s.Require()

	// Initial stores
	keyA := types.NewKVStoreKey("storeA")
	keyB := types.NewKVStoreKey("storeB")

	s.store.MountStoreWithDB(keyA, types.StoreTypeIAVL, nil)
	s.store.MountStoreWithDB(keyB, types.StoreTypeIAVL, nil)

	err := s.store.LoadLatestVersion()
	require.NoError(err)

	// Write initial data and commit v1
	kvA := s.store.GetKVStore(keyA)
	kvA.Set([]byte("keyA1"), []byte("valueA1"))
	kvB := s.store.GetKVStore(keyB)
	kvB.Set([]byte("keyB1"), []byte("valueB1"))
	commitID1 := s.store.Commit()
	require.Equal(int64(1), commitID1.Version)

	// Close the store before upgrade
	err = s.store.Close()
	require.NoError(err)

	// --- Define Upgrades ---
	keyC := types.NewKVStoreKey("storeC")             // New store to add
	keyRenamed := types.NewKVStoreKey("storeRenamed") // New name for storeA

	upgrades := &types.StoreUpgrades{
		Added:   []string{keyC.Name()},
		Deleted: []string{keyB.Name()},
		Renamed: []types.StoreRename{
			{OldKey: keyA.Name(), NewKey: keyRenamed.Name()},
		},
	}

	// --- Load with Upgrades ---
	// Create a new store instance for loading with upgrades
	upgradedStore := rootmulti.NewStore(s.dir, log.NewNopLogger(), false)

	// Mount the *new* set of stores expected *after* the upgrade
	upgradedStore.MountStoreWithDB(keyRenamed, types.StoreTypeIAVL, nil)
	upgradedStore.MountStoreWithDB(keyC, types.StoreTypeIAVL, nil)
	// Do NOT mount keyA or keyB as they are involved in rename/delete

	// Load version 1 with upgrades applied
	err = upgradedStore.LoadVersionAndUpgrade(1, upgrades)
	require.NoError(err)
	require.Equal(int64(1), upgradedStore.LastCommitID().Version)
	require.Equal(commitID1.Hash, upgradedStore.LastCommitID().Hash) // Hash should match v1

	// Verify store structure after upgrade
	require.NotNil(upgradedStore.GetStoreByName(keyRenamed.Name()), "storeRenamed should exist")
	require.NotNil(upgradedStore.GetStoreByName(keyC.Name()), "storeC should exist")
	require.Nil(upgradedStore.GetStoreByName(keyA.Name()), "storeA should not exist")
	require.Nil(upgradedStore.GetStoreByName(keyB.Name()), "storeB should not exist")

	// Verify data migration (storeA -> storeRenamed)
	kvRenamed := upgradedStore.GetKVStore(keyRenamed)
	require.Equal([]byte("valueA1"), kvRenamed.Get([]byte("keyA1")))

	// Verify new store is empty
	kvC := upgradedStore.GetKVStore(keyC)
	require.Nil(kvC.Get([]byte("anykey")))

	// --- Commit after upgrade ---
	kvRenamed.Set([]byte("keyRenamed2"), []byte("valueRenamed2"))
	kvC.Set([]byte("keyC1"), []byte("valueC1"))
	commitID2 := upgradedStore.Commit()
	require.Equal(int64(2), commitID2.Version)
	require.NotEmpty(commitID2.Hash)

	// Close the upgraded store
	err = upgradedStore.Close()
	require.NoError(err)

	// --- Reload after upgrade and commit ---
	reloadedStore := rootmulti.NewStore(s.dir, log.NewNopLogger(), false)
	reloadedStore.MountStoreWithDB(keyRenamed, types.StoreTypeIAVL, nil)
	reloadedStore.MountStoreWithDB(keyC, types.StoreTypeIAVL, nil)
	err = reloadedStore.LoadLatestVersion() // Should load v2
	require.NoError(err)
	require.Equal(int64(2), reloadedStore.LastCommitID().Version)
	require.Equal(commitID2.Hash, reloadedStore.LastCommitID().Hash)

	// Verify data after reload
	kvRenamed = reloadedStore.GetKVStore(keyRenamed)
	require.Equal([]byte("valueA1"), kvRenamed.Get([]byte("keyA1")))
	require.Equal([]byte("valueRenamed2"), kvRenamed.Get([]byte("keyRenamed2")))
	kvC = reloadedStore.GetKVStore(keyC)
	require.Equal([]byte("valueC1"), kvC.Get([]byte("keyC1")))

	err = reloadedStore.Close()
	require.NoError(err)

	s.store = nil // Prevent teardown from closing the store again
}

func (s *StoreTestSuite) TestErrorConditions() {
	require := s.Require()

	// --- Test Mounting Errors ---
	key1 := types.NewKVStoreKey("store1")
	key1DupName := types.NewKVStoreKey("store1") // Same name, different key instance
	key2 := types.NewKVStoreKey("store2")

	s.store.MountStoreWithDB(key1, types.StoreTypeIAVL, nil)

	// Mount duplicate key instance
	require.Panics(func() {
		s.store.MountStoreWithDB(key1, types.StoreTypeIAVL, nil)
	}, "Panicked mounting duplicate key instance")

	// Mount duplicate key name
	require.Panics(func() {
		s.store.MountStoreWithDB(key1DupName, types.StoreTypeIAVL, nil)
	}, "Panicked mounting duplicate key name")

	// Mount nil key
	require.PanicsWithValue("MountIAVLStore() key cannot be nil", func() {
		s.store.MountStoreWithDB(nil, types.StoreTypeIAVL, nil)
	}, "Panicked mounting nil key")

	// --- Test Loading Errors ---
	s.store.MountStoreWithDB(key2, types.StoreTypeIAVL, nil) // Mount another store to proceed
	err := s.store.LoadLatestVersion()
	require.NoError(err)

	// Commit v1
	kv1 := s.store.GetKVStore(key1)
	kv1.Set([]byte("a"), []byte("1"))
	commitID1 := s.store.Commit()
	require.Equal(int64(1), commitID1.Version)

	// Load version > latest
	err = s.store.LoadVersion(2)
	require.Error(err) // memiavl Load returns error if target > latest

	// Load version with overflow
	err = s.store.LoadVersionAndUpgrade(int64(1)+int64(uint32(0xFFFFFFFF)), nil)
	require.Error(err)
	require.Contains(err.Error(), "version overflows uint32")

}
