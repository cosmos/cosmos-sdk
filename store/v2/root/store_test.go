package root

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/pruning"
)

const (
	testStoreKey  = "test_store_key"
	testStoreKey2 = "test_store_key2"
	testStoreKey3 = "test_store_key3"
)

var testStoreKeys = []string{testStoreKey, testStoreKey2, testStoreKey3}

var (
	testStoreKeyBytes  = []byte(testStoreKey)
	testStoreKey2Bytes = []byte(testStoreKey2)
	testStoreKey3Bytes = []byte(testStoreKey3)
)

type RootStoreTestSuite struct {
	suite.Suite

	rootStore store.RootStore
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, &RootStoreTestSuite{})
}

func (s *RootStoreTestSuite) SetupTest() {
	noopLog := coretesting.NewNopLogger()

	tree := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	tree2 := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	tree3 := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	sc, err := commitment.NewCommitStore(map[string]commitment.Tree{testStoreKey: tree, testStoreKey2: tree2, testStoreKey3: tree3}, nil, dbm.NewMemDB(), noopLog)
	s.Require().NoError(err)

	pm := pruning.NewManager(sc, nil)
	rs, err := New(dbm.NewMemDB(), noopLog, sc, pm, nil)
	s.Require().NoError(err)

	s.rootStore = rs
}

func (s *RootStoreTestSuite) newStoreWithPruneConfig(config *store.PruningOption) {
	noopLog := coretesting.NewNopLogger()

	mdb := dbm.NewMemDB()
	multiTrees := make(map[string]commitment.Tree)
	for _, storeKey := range testStoreKeys {
		prefixDB := dbm.NewPrefixDB(mdb, []byte(storeKey))
		multiTrees[storeKey] = iavl.NewIavlTree(prefixDB, noopLog, iavl.DefaultConfig())
	}

	sc, err := commitment.NewCommitStore(multiTrees, nil, dbm.NewMemDB(), noopLog)
	s.Require().NoError(err)

	pm := pruning.NewManager(sc, config)

	rs, err := New(dbm.NewMemDB(), noopLog, sc, pm, nil)
	s.Require().NoError(err)

	s.rootStore = rs
}

func (s *RootStoreTestSuite) newStoreWithBackendMount(sc store.Committer, pm *pruning.Manager) {
	noopLog := coretesting.NewNopLogger()

	rs, err := New(dbm.NewMemDB(), noopLog, sc, pm, nil)
	s.Require().NoError(err)

	s.rootStore = rs
}

func (s *RootStoreTestSuite) TearDownTest() {
	err := s.rootStore.Close()
	s.Require().NoError(err)
}

func (s *RootStoreTestSuite) TestGetStateCommitment() {
	s.Require().Equal(s.rootStore.GetStateCommitment(), s.rootStore.(*Store).stateCommitment)
}

func (s *RootStoreTestSuite) TestSetInitialVersion() {
	initialVersion := uint64(5)
	s.Require().NoError(s.rootStore.SetInitialVersion(initialVersion))

	// perform an initial, empty commit
	cs := corestore.NewChangeset(initialVersion)
	cs.Add(testStoreKeyBytes, []byte("foo"), []byte("bar"), false)
	_, err := s.rootStore.Commit(corestore.NewChangeset(initialVersion))
	s.Require().NoError(err)

	// check the latest version
	lVersion, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(initialVersion, lVersion)

	// set the initial version again
	rInitialVersion := uint64(100)
	s.Require().NoError(s.rootStore.SetInitialVersion(rInitialVersion))

	// TODO fix version munging here
	// perform the commit
	cs = corestore.NewChangeset(initialVersion + 1)
	cs.Add(testStoreKey2Bytes, []byte("foo"), []byte("bar"), false)
	_, err = s.rootStore.Commit(cs)
	s.Require().NoError(err)
	lVersion, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	// SetInitialVersion only works once
	s.Require().NotEqual(rInitialVersion, lVersion)
	s.Require().Equal(initialVersion+1, lVersion)
}

func (s *RootStoreTestSuite) TestQuery() {
	_, err := s.rootStore.Query([]byte{}, 1, []byte("foo"), true)
	s.Require().Error(err)

	// write and commit a changeset
	cs := corestore.NewChangeset(1)
	cs.Add(testStoreKeyBytes, []byte("foo"), []byte("bar"), false)

	commitHash, err := s.rootStore.Commit(cs)
	s.Require().NoError(err)
	s.Require().NotNil(commitHash)

	// ensure the proof is non-nil for the corresponding version
	result, err := s.rootStore.Query([]byte(testStoreKey), 1, []byte("foo"), true)
	s.Require().NoError(err)
	s.Require().NotNil(result.ProofOps)
	s.Require().Equal([]byte("foo"), result.ProofOps[0].Key)
}

func (s *RootStoreTestSuite) TestGetFallback() {
	sc := s.rootStore.GetStateCommitment()

	// create a changeset and commit it to SC ONLY
	cs := corestore.NewChangeset(1)
	cs.Add(testStoreKeyBytes, []byte("foo"), []byte("bar"), false)

	err := sc.WriteChangeset(cs)
	s.Require().NoError(err)

	_, err = sc.Commit(cs.Version)
	s.Require().NoError(err)

	// ensure we can query for the key, which should fallback to SC
	qResult, err := s.rootStore.Query(testStoreKeyBytes, 1, []byte("foo"), false)
	s.Require().NoError(err)
	s.Require().Equal([]byte("bar"), qResult.Value)

	// non-existent key
	qResult, err = s.rootStore.Query(testStoreKeyBytes, 1, []byte("non_existent_key"), false)
	s.Require().NoError(err)
	s.Require().Nil(qResult.Value)
}

func (s *RootStoreTestSuite) TestQueryProof() {
	cs := corestore.NewChangeset(1)
	// testStoreKey
	cs.Add(testStoreKeyBytes, []byte("key1"), []byte("value1"), false)
	cs.Add(testStoreKeyBytes, []byte("key2"), []byte("value2"), false)
	// testStoreKey2
	cs.Add(testStoreKey2Bytes, []byte("key3"), []byte("value3"), false)
	// testStoreKey3
	cs.Add(testStoreKey3Bytes, []byte("key4"), []byte("value4"), false)

	// commit
	_, err := s.rootStore.Commit(cs)
	s.Require().NoError(err)

	// query proof for testStoreKey
	result, err := s.rootStore.Query(testStoreKeyBytes, 1, []byte("key1"), true)
	s.Require().NoError(err)
	s.Require().NotNil(result.ProofOps)
	cInfo, err := s.rootStore.GetStateCommitment().GetCommitInfo(1)
	s.Require().NoError(err)
	storeHash := cInfo.GetStoreCommitID(testStoreKeyBytes).Hash
	treeRoots, err := result.ProofOps[0].Run([][]byte{[]byte("value1")})
	s.Require().NoError(err)
	s.Require().Equal(treeRoots[0], storeHash)
	expRoots, err := result.ProofOps[1].Run([][]byte{storeHash})
	s.Require().NoError(err)
	s.Require().Equal(expRoots[0], cInfo.Hash())
}

func (s *RootStoreTestSuite) TestLoadVersion() {
	// write and commit a few changesets
	for v := uint64(1); v <= 5; v++ {
		val := fmt.Sprintf("val%03d", v) // val001, val002, ..., val005

		cs := corestore.NewChangeset(v)
		cs.Add(testStoreKeyBytes, []byte("key"), []byte(val), false)

		commitHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(commitHash)
	}

	// ensure the latest version is correct
	latest, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(5), latest)

	// attempt to load a non-existent version
	err = s.rootStore.LoadVersion(6)
	s.Require().Error(err)

	// attempt to load a previously committed version
	err = s.rootStore.LoadVersion(3)
	s.Require().NoError(err)

	// ensure the latest version is correct
	latest, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), latest)

	// query state and ensure values returned are based on the loaded version
	_, ro, err := s.rootStore.StateLatest()
	s.Require().NoError(err)

	reader, err := ro.GetReader(testStoreKeyBytes)
	s.Require().NoError(err)
	val, err := reader.Get([]byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("val003"), val)

	// attempt to write and commit a few changesets
	for v := 4; v <= 5; v++ {
		val := fmt.Sprintf("overwritten_val%03d", v) // overwritten_val004, overwritten_val005

		cs := corestore.NewChangeset(uint64(v))
		cs.Add(testStoreKeyBytes, []byte("key"), []byte(val), false)

		_, err := s.rootStore.Commit(cs)
		s.Require().Error(err)
	}

	// ensure the latest version is correct
	latest, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), latest) // should have stayed at 3 after failed commits

	// query state and ensure values returned are based on the loaded version
	_, ro, err = s.rootStore.StateLatest()
	s.Require().NoError(err)

	reader, err = ro.GetReader(testStoreKeyBytes)
	s.Require().NoError(err)
	val, err = reader.Get([]byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("val003"), val)
}

func (s *RootStoreTestSuite) TestLoadVersionForOverwriting() {
	// write and commit a few changesets
	for v := uint64(1); v <= 5; v++ {
		val := fmt.Sprintf("val%03d", v) // val001, val002, ..., val005

		cs := corestore.NewChangeset(v)
		cs.Add(testStoreKeyBytes, []byte("key"), []byte(val), false)

		commitHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(commitHash)
	}

	// ensure the latest version is correct
	latest, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(5), latest)

	// attempt to load a non-existent version
	err = s.rootStore.LoadVersionForOverwriting(6)
	s.Require().Error(err)

	// attempt to load a previously committed version
	err = s.rootStore.LoadVersionForOverwriting(3)
	s.Require().NoError(err)

	// ensure the latest version is correct
	latest, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), latest)

	// query state and ensure values returned are based on the loaded version
	_, ro, err := s.rootStore.StateLatest()
	s.Require().NoError(err)

	reader, err := ro.GetReader(testStoreKeyBytes)
	s.Require().NoError(err)
	val, err := reader.Get([]byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("val003"), val)

	// attempt to write and commit a few changesets
	for v := 4; v <= 5; v++ {
		val := fmt.Sprintf("overwritten_val%03d", v) // overwritten_val004, overwritten_val005

		cs := corestore.NewChangeset(uint64(v))
		cs.Add(testStoreKeyBytes, []byte("key"), []byte(val), false)

		commitHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(commitHash)
	}

	// ensure the latest version is correct
	latest, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(5), latest)

	// query state and ensure values returned are based on the loaded version
	_, ro, err = s.rootStore.StateLatest()
	s.Require().NoError(err)

	reader, err = ro.GetReader(testStoreKeyBytes)
	s.Require().NoError(err)
	val, err = reader.Get([]byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("overwritten_val005"), val)
}

func (s *RootStoreTestSuite) TestCommit() {
	lv, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Zero(lv)

	// perform changes
	cs := corestore.NewChangeset(1)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		cs.Add(testStoreKeyBytes, []byte(key), []byte(val), false)
	}

	cHash, err := s.rootStore.Commit(cs)
	s.Require().NoError(err)
	s.Require().NotNil(cHash)

	// ensure latest version is updated
	lv, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), lv)

	// perform reads on the updated root store
	_, ro, err := s.rootStore.StateLatest()
	s.Require().NoError(err)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		reader, err := ro.GetReader(testStoreKeyBytes)
		s.Require().NoError(err)
		result, err := reader.Get([]byte(key))
		s.Require().NoError(err)

		s.Require().Equal([]byte(val), result)
	}
}

func (s *RootStoreTestSuite) TestStateAt() {
	// write keys over multiple versions
	for v := uint64(1); v <= 5; v++ {
		// perform changes
		cs := corestore.NewChangeset(v)
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
			val := fmt.Sprintf("val%03d_%03d", i, v) // val000_1, val001_1, ..., val099_1

			cs.Add(testStoreKeyBytes, []byte(key), []byte(val), false)
		}

		// execute Commit
		cHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(cHash)
	}

	lv, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(5), lv)

	// ensure we can read state correctly at each version
	for v := uint64(1); v <= 5; v++ {
		ro, err := s.rootStore.StateAt(v)
		s.Require().NoError(err)

		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
			val := fmt.Sprintf("val%03d_%03d", i, v) // val000_1, val001_1, ..., val099_1

			reader, err := ro.GetReader(testStoreKeyBytes)
			s.Require().NoError(err)
			isExist, err := reader.Has([]byte(key))
			s.Require().NoError(err)
			s.Require().True(isExist)
			result, err := reader.Get([]byte(key))
			s.Require().NoError(err)
			s.Require().Equal([]byte(val), result)
		}

		// non-existent key
		reader, err := ro.GetReader(testStoreKey2Bytes)
		s.Require().NoError(err)
		isExist, err := reader.Has([]byte("key"))
		s.Require().NoError(err)
		s.Require().False(isExist)
		v, err := reader.Get([]byte("key"))
		s.Require().NoError(err)
		s.Require().Nil(v)
	}
}

func (s *RootStoreTestSuite) TestPrune() {
	// perform changes
	cs := corestore.NewChangeset(1)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		cs.Add(testStoreKeyBytes, []byte(key), []byte(val), false)
	}

	testCases := []struct {
		name        string
		numVersions int64
		po          store.PruningOption
		deleted     []uint64
		saved       []uint64
	}{
		{"prune nothing", 10, store.PruningOption{
			KeepRecent: 0,
			Interval:   0,
		}, nil, []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"prune everything", 12, store.PruningOption{
			KeepRecent: 1,
			Interval:   10,
		}, []uint64{1, 2, 3, 4, 5, 6, 7, 8}, []uint64{9, 10, 11, 12}},
		{"prune some; no batch", 10, store.PruningOption{
			KeepRecent: 2,
			Interval:   1,
		}, []uint64{1, 2, 3, 4, 6, 5, 7}, []uint64{8, 9, 10}},
		{"prune some; small batch", 10, store.PruningOption{
			KeepRecent: 2,
			Interval:   3,
		}, []uint64{1, 2, 3, 4, 5, 6}, []uint64{7, 8, 9, 10}},
		{"prune some; large batch", 10, store.PruningOption{
			KeepRecent: 2,
			Interval:   11,
		}, nil, []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}

	for _, tc := range testCases {

		s.newStoreWithPruneConfig(&tc.po)

		// write keys over multiple versions
		for i := int64(0); i < tc.numVersions; i++ {
			// execute Commit
			cs.Version = uint64(i + 1)
			cHash, err := s.rootStore.Commit(cs)
			s.Require().NoError(err)
			s.Require().NotNil(cHash)
		}

		for _, v := range tc.saved {
			ro, err := s.rootStore.StateAt(v)
			s.Require().NoError(err, "expected no error when loading height %d at test %s", v, tc.name)

			for i := 0; i < 10; i++ {
				key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
				val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

				reader, err := ro.GetReader(testStoreKeyBytes)
				s.Require().NoError(err)
				result, err := reader.Get([]byte(key))
				s.Require().NoError(err)
				s.Require().Equal([]byte(val), result, "value should be equal for test: %s", tc.name)
			}
		}

		for _, v := range tc.deleted {
			var err error
			checkErr := func() bool {
				if _, err = s.rootStore.StateAt(v); err != nil {
					return true
				}
				return false
			}
			// wait for async pruning process to finish
			s.Require().Eventually(checkErr, 2*time.Second, 100*time.Millisecond)
			s.Require().Error(err, "expected error when loading height %d at test %s", v, tc.name)
		}
	}
}

func (s *RootStoreTestSuite) TestMultiStore_Pruning_SameHeightsTwice() {
	// perform changes
	cs := corestore.NewChangeset(1)
	cs.Add(testStoreKeyBytes, []byte("key"), []byte("val"), false)

	const (
		numVersions uint64 = 10
		keepRecent  uint64 = 1
		interval    uint64 = 10
	)

	s.newStoreWithPruneConfig(&store.PruningOption{
		KeepRecent: keepRecent,
		Interval:   interval,
	})
	s.Require().NoError(s.rootStore.LoadLatestVersion())

	for i := uint64(0); i < numVersions; i++ {
		// execute Commit
		cs.Version = i + 1
		cHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(cHash)
	}

	latestVer, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(numVersions, latestVer)

	for v := uint64(1); v < numVersions-keepRecent; v++ {
		var err error
		checkErr := func() bool {
			if _, err = s.rootStore.StateAt(v); err != nil {
				return true
			}
			return false
		}
		// wait for async pruning process to finish
		s.Require().Eventually(checkErr, 2*time.Second, 100*time.Millisecond, "expected no error when loading height: %d", v)
	}

	for v := (numVersions - keepRecent); v < numVersions; v++ {
		_, err := s.rootStore.StateAt(v)
		s.Require().NoError(err, "expected no error when loading height: %d", v)
	}

	// Get latest
	err = s.rootStore.LoadVersion(numVersions)
	s.Require().NoError(err)

	// Test pruning the same heights again
	cs.Version++
	_, err = s.rootStore.Commit(cs)
	s.Require().NoError(err)

	// Ensure that can commit one more height with no panic
	cs.Version++
	_, err = s.rootStore.Commit(cs)
	s.Require().NoError(err)
}

func (s *RootStoreTestSuite) TestMultiStore_PruningRestart() {
	// perform changes
	cs := corestore.NewChangeset(1)
	cs.Add(testStoreKeyBytes, []byte("key"), []byte("val"), false)

	pruneOpt := &store.PruningOption{
		KeepRecent: 2,
		Interval:   11,
	}

	noopLog := coretesting.NewNopLogger()

	mdb1 := dbm.NewMemDB()
	mdb2 := dbm.NewMemDB()

	tree := iavl.NewIavlTree(mdb1, noopLog, iavl.DefaultConfig())
	sc, err := commitment.NewCommitStore(map[string]commitment.Tree{testStoreKey: tree}, nil, mdb2, noopLog)
	s.Require().NoError(err)

	pm := pruning.NewManager(sc, pruneOpt)

	s.newStoreWithBackendMount(sc, pm)
	s.Require().NoError(s.rootStore.LoadLatestVersion())

	// Commit enough to build up heights to prune, where on the next block we should
	// batch delete.
	for i := uint64(1); i <= 10; i++ {
		// execute Commit
		cs.Version = i
		cHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(cHash)
	}

	latestVer, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)

	ok, actualHeightToPrune := pruneOpt.ShouldPrune(latestVer)
	s.Require().False(ok)
	s.Require().Equal(uint64(0), actualHeightToPrune)

	tree = iavl.NewIavlTree(mdb1, noopLog, iavl.DefaultConfig())
	sc, err = commitment.NewCommitStore(map[string]commitment.Tree{testStoreKey: tree}, nil, mdb2, noopLog)
	s.Require().NoError(err)

	pm = pruning.NewManager(sc, pruneOpt)

	s.newStoreWithBackendMount(sc, pm)
	err = s.rootStore.LoadLatestVersion()
	s.Require().NoError(err)

	latestVer, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)

	ok, actualHeightToPrune = pruneOpt.ShouldPrune(latestVer)
	s.Require().False(ok)
	s.Require().Equal(uint64(0), actualHeightToPrune)

	// commit one more block and ensure the heights have been pruned
	// execute Commit
	cs.Version++
	cHash, err := s.rootStore.Commit(cs)
	s.Require().NoError(err)
	s.Require().NotNil(cHash)

	latestVer, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)

	ok, actualHeightToPrune = pruneOpt.ShouldPrune(latestVer)
	s.Require().True(ok)
	s.Require().Equal(uint64(8), actualHeightToPrune)

	for v := uint64(1); v <= actualHeightToPrune; v++ {
		checkErr := func() bool {
			if _, err = s.rootStore.StateAt(v); err != nil {
				return true
			}
			return false
		}
		// wait for async pruning process to finish
		s.Require().Eventually(checkErr, 10*time.Second, 1*time.Second, "expected error when loading height: %d", v)
	}
}

func (s *RootStoreTestSuite) TestMultiStoreRestart() {
	noopLog := coretesting.NewNopLogger()

	mdb1 := dbm.NewMemDB()
	mdb2 := dbm.NewMemDB()
	multiTrees := make(map[string]commitment.Tree)
	for _, storeKey := range testStoreKeys {
		prefixDB := dbm.NewPrefixDB(mdb1, []byte(storeKey))
		multiTrees[storeKey] = iavl.NewIavlTree(prefixDB, noopLog, iavl.DefaultConfig())
	}

	sc, err := commitment.NewCommitStore(multiTrees, nil, mdb2, noopLog)
	s.Require().NoError(err)

	pm := pruning.NewManager(sc, nil)

	s.newStoreWithBackendMount(sc, pm)
	s.Require().NoError(s.rootStore.LoadLatestVersion())

	// perform changes
	for i := 1; i < 3; i++ {
		cs := corestore.NewChangeset(uint64(i))
		key := fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d_%03d", i, 1) // val000_1, val001_1, ..., val099_1

		cs.Add(testStoreKeyBytes, []byte(key), []byte(val), false)

		key = fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
		val = fmt.Sprintf("val%03d_%03d", i, 2) // val000_1, val001_1, ..., val099_1

		cs.Add(testStoreKey2Bytes, []byte(key), []byte(val), false)

		key = fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
		val = fmt.Sprintf("val%03d_%03d", i, 3) // val000_1, val001_1, ..., val099_1

		cs.Add(testStoreKey3Bytes, []byte(key), []byte(val), false)

		// execute Commit
		cHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(cHash)

		latestVer, err := s.rootStore.GetLatestVersion()
		s.Require().NoError(err)
		s.Require().Equal(uint64(i), latestVer)
	}

	// more changes
	cs1 := corestore.NewChangeset(3)
	key := fmt.Sprintf("key%03d", 3)         // key000, key001, ..., key099
	val := fmt.Sprintf("val%03d_%03d", 3, 1) // val000_1, val001_1, ..., val099_1

	cs1.Add(testStoreKeyBytes, []byte(key), []byte(val), false)

	key = fmt.Sprintf("key%03d", 3)         // key000, key001, ..., key099
	val = fmt.Sprintf("val%03d_%03d", 3, 2) // val000_1, val001_1, ..., val099_1

	cs1.Add(testStoreKey2Bytes, []byte(key), []byte(val), false)

	// execute Commit
	cHash, err := s.rootStore.Commit(cs1)
	s.Require().NoError(err)
	s.Require().NotNil(cHash)

	latestVer, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), latestVer)

	cs2 := corestore.NewChangeset(4)
	key = fmt.Sprintf("key%03d", 4)         // key000, key001, ..., key099
	val = fmt.Sprintf("val%03d_%03d", 4, 3) // val000_1, val001_1, ..., val099_1

	cs2.Add(testStoreKey3Bytes, []byte(key), []byte(val), false)

	// execute Commit
	cHash, err = s.rootStore.Commit(cs2)
	s.Require().NoError(err)
	s.Require().NotNil(cHash)

	latestVer, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(4), latestVer)

	_, ro1, err := s.rootStore.StateLatest()
	s.Require().Nil(err)
	reader1, err := ro1.GetReader(testStoreKeyBytes)
	s.Require().NoError(err)
	result1, err := reader1.Get([]byte(fmt.Sprintf("key%03d", 3)))
	s.Require().NoError(err)
	s.Require().Equal([]byte(fmt.Sprintf("val%03d_%03d", 3, 1)), result1, "value should be equal")

	// "restart"
	multiTrees = make(map[string]commitment.Tree)
	for _, storeKey := range testStoreKeys {
		prefixDB := dbm.NewPrefixDB(mdb1, []byte(storeKey))
		multiTrees[storeKey] = iavl.NewIavlTree(prefixDB, noopLog, iavl.DefaultConfig())
	}

	sc, err = commitment.NewCommitStore(multiTrees, nil, mdb2, noopLog)
	s.Require().NoError(err)

	pm = pruning.NewManager(sc, nil)

	s.newStoreWithBackendMount(sc, pm)
	err = s.rootStore.LoadLatestVersion()
	s.Require().Nil(err)

	latestVer, ro, err := s.rootStore.StateLatest()
	s.Require().Nil(err)
	s.Require().Equal(uint64(4), latestVer)
	reader, err := ro.GetReader(testStoreKeyBytes)
	s.Require().NoError(err)
	result, err := reader.Get([]byte(fmt.Sprintf("key%03d", 3)))
	s.Require().NoError(err)
	s.Require().Equal([]byte(fmt.Sprintf("val%03d_%03d", 3, 1)), result, "value should be equal")

	reader, err = ro.GetReader(testStoreKey2Bytes)
	s.Require().NoError(err)
	result, err = reader.Get([]byte(fmt.Sprintf("key%03d", 2)))
	s.Require().NoError(err)
	s.Require().Equal([]byte(fmt.Sprintf("val%03d_%03d", 2, 2)), result, "value should be equal")

	reader, err = ro.GetReader(testStoreKey3Bytes)
	s.Require().NoError(err)
	result, err = reader.Get([]byte(fmt.Sprintf("key%03d", 4)))
	s.Require().NoError(err)
	s.Require().Equal([]byte(fmt.Sprintf("val%03d_%03d", 4, 3)), result, "value should be equal")
}

func (s *RootStoreTestSuite) TestHashStableWithEmptyCommitAndRestart() {
	err := s.rootStore.LoadLatestVersion()
	s.Require().NoError(err)

	emptyHash := sha256.Sum256([]byte{})
	appHash := emptyHash[:]
	commitID := proof.CommitID{Hash: appHash}
	lastCommitID, err := s.rootStore.LastCommitID()
	s.Require().Nil(err)

	// the hash of a store with no commits is the root hash of a tree with empty hashes as leaves.
	// it should not be equal an empty hash.
	s.Require().NotEqual(commitID, lastCommitID)

	cs := corestore.NewChangeset(1)
	cs.Add(testStoreKeyBytes, []byte("key"), []byte("val"), false)

	cHash, err := s.rootStore.Commit(cs)
	s.Require().Nil(err)
	s.Require().NotNil(cHash)
	latestVersion, err := s.rootStore.GetLatestVersion()
	hash := cHash
	s.Require().Nil(err)
	s.Require().Equal(uint64(1), latestVersion)

	// make an empty commit, it should update version, but not affect hash
	cHash, err = s.rootStore.Commit(corestore.NewChangeset(2))
	s.Require().Nil(err)
	s.Require().NotNil(cHash)
	latestVersion, err = s.rootStore.GetLatestVersion()
	s.Require().Nil(err)
	s.Require().Equal(uint64(2), latestVersion)
	s.Require().Equal(hash, cHash)

	// reload the store
	s.Require().NoError(s.rootStore.LoadLatestVersion())
	lastCommitID, err = s.rootStore.LastCommitID()
	s.Require().NoError(err)
	s.Require().Equal(lastCommitID.Hash, hash)
}
