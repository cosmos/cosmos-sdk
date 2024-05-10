package root

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	coreheader "cosmossdk.io/core/header"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/sqlite"
)

const (
	testStoreKey  = "test_store_key"
	testStoreKey2 = "test_store_key2"
	testStoreKey3 = "test_store_key3"
)

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
	noopLog := log.NewNopLogger()

	sqliteDB, err := sqlite.New(s.T().TempDir())
	s.Require().NoError(err)
	ss := storage.NewStorageStore(sqliteDB, nil, noopLog)

	tree := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	tree2 := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	tree3 := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	sc, err := commitment.NewCommitStore(map[string]commitment.Tree{testStoreKey: tree, testStoreKey2: tree2, testStoreKey3: tree3}, dbm.NewMemDB(), nil, noopLog)
	s.Require().NoError(err)

	rs, err := New(noopLog, ss, sc, nil, nil)
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

func (s *RootStoreTestSuite) TestGetStateStorage() {
	s.Require().Equal(s.rootStore.GetStateStorage(), s.rootStore.(*Store).stateStorage)
}

func (s *RootStoreTestSuite) TestSetInitialVersion() {
	s.Require().NoError(s.rootStore.SetInitialVersion(100))
}

func (s *RootStoreTestSuite) TestSetCommitHeader() {
	h := &coreheader.Info{
		Height:  100,
		Hash:    []byte("foo"),
		ChainID: "test",
	}
	s.rootStore.SetCommitHeader(h)

	s.Require().Equal(h, s.rootStore.(*Store).commitHeader)
}

func (s *RootStoreTestSuite) TestQuery() {
	_, err := s.rootStore.Query([]byte{}, 1, []byte("foo"), true)
	s.Require().Error(err)

	// write and commit a changeset
	cs := corestore.NewChangeset()
	cs.Add(testStoreKeyBytes, []byte("foo"), []byte("bar"), false)

	workingHash, err := s.rootStore.WorkingHash(cs)
	s.Require().NoError(err)
	s.Require().NotNil(workingHash)

	commitHash, err := s.rootStore.Commit(cs)
	s.Require().NoError(err)
	s.Require().NotNil(commitHash)
	s.Require().Equal(workingHash, commitHash)

	// ensure the proof is non-nil for the corresponding version
	result, err := s.rootStore.Query([]byte(testStoreKey), 1, []byte("foo"), true)
	s.Require().NoError(err)
	s.Require().NotNil(result.ProofOps)
	s.Require().Equal([]byte("foo"), result.ProofOps[0].Key)
}

func (s *RootStoreTestSuite) TestGetFallback() {
	sc := s.rootStore.GetStateCommitment()

	// create a changeset and commit it to SC ONLY
	cs := corestore.NewChangeset()
	cs.Add(testStoreKeyBytes, []byte("foo"), []byte("bar"), false)

	err := sc.WriteBatch(cs)
	s.Require().NoError(err)

	ci := sc.WorkingCommitInfo(1)
	_, err = sc.Commit(ci.Version)
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
	cs := corestore.NewChangeset()
	// testStoreKey
	cs.Add(testStoreKeyBytes, []byte("key1"), []byte("value1"), false)
	cs.Add(testStoreKeyBytes, []byte("key2"), []byte("value2"), false)
	// testStoreKey2
	cs.Add(testStoreKey2Bytes, []byte("key3"), []byte("value3"), false)
	// testStoreKey3
	cs.Add(testStoreKey3Bytes, []byte("key4"), []byte("value4"), false)

	// commit
	_, err := s.rootStore.WorkingHash(cs)
	s.Require().NoError(err)
	_, err = s.rootStore.Commit(cs)
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
	for v := 1; v <= 5; v++ {
		val := fmt.Sprintf("val%03d", v) // val001, val002, ..., val005

		cs := corestore.NewChangeset()
		cs.Add(testStoreKeyBytes, []byte("key"), []byte(val), false)

		workingHash, err := s.rootStore.WorkingHash(cs)
		s.Require().NoError(err)
		s.Require().NotNil(workingHash)

		commitHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(commitHash)
		s.Require().Equal(workingHash, commitHash)
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

		cs := corestore.NewChangeset()
		cs.Add(testStoreKeyBytes, []byte("key"), []byte(val), false)

		workingHash, err := s.rootStore.WorkingHash(cs)
		s.Require().NoError(err)
		s.Require().NotNil(workingHash)

		commitHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(commitHash)
		s.Require().Equal(workingHash, commitHash)
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
	cs := corestore.NewChangeset()
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		cs.Add(testStoreKeyBytes, []byte(key), []byte(val), false)
	}

	// committing w/o calling WorkingHash should error
	_, err = s.rootStore.Commit(cs)
	s.Require().Error(err)

	// execute WorkingHash and Commit
	wHash, err := s.rootStore.WorkingHash(cs)
	s.Require().NoError(err)

	cHash, err := s.rootStore.Commit(cs)
	s.Require().NoError(err)
	s.Require().Equal(wHash, cHash)

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
		cs := corestore.NewChangeset()
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
			val := fmt.Sprintf("val%03d_%03d", i, v) // val000_1, val001_1, ..., val099_1

			cs.Add(testStoreKeyBytes, []byte(key), []byte(val), false)
		}

		// execute WorkingHash and Commit
		wHash, err := s.rootStore.WorkingHash(cs)
		s.Require().NoError(err)

		cHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().Equal(wHash, cHash)
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
			result, err := reader.Get([]byte(key))
			s.Require().NoError(err)
			s.Require().Equal([]byte(val), result)
		}
	}
}
