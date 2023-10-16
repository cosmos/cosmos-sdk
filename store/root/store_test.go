package root

import (
	"fmt"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	"cosmossdk.io/store/v2/storage/sqlite"
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

	ss, err := sqlite.New(s.T().TempDir())
	s.Require().NoError(err)

	tree := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	sc := commitment.NewDatabase(tree)

	rootStore, err := New(noopLog, 1, ss, sc)
	s.Require().NoError(err)

	s.rootStore = rootStore
}

func (s *RootStoreTestSuite) TestClose() {
	s.Require().NoError(s.rootStore.Close())
}

func (s *RootStoreTestSuite) TestMountSCStore() {
	s.Require().Error(s.rootStore.MountSCStore("", nil))
}

func (s *RootStoreTestSuite) TestGetSCStore() {
	s.Require().Equal(s.rootStore.GetSCStore(""), s.rootStore.(*Store).stateCommitment)
}

func (s *RootStoreTestSuite) TestGetKVStore() {
	s.Require().Equal(s.rootStore.GetKVStore(""), s.rootStore.(*Store).rootKVStore)
}

func (s *RootStoreTestSuite) TestGetBranchedKVStore() {
	bs := s.rootStore.GetBranchedKVStore("")
	s.Require().NotNil(bs)
	s.Require().Equal(store.StoreTypeBranch, bs.GetStoreType())
	s.Require().Empty(bs.GetChangeset().Pairs)
}

func (s *RootStoreTestSuite) TestGetProof() {
	p, err := s.rootStore.GetProof("", 1, []byte("foo"))
	s.Require().Error(err)
	s.Require().Nil(p)

	// write and commit a changeset
	bs := s.rootStore.GetBranchedKVStore("")
	bs.Set([]byte("foo"), []byte("bar"))

	workingHash, err := s.rootStore.WorkingHash()
	s.Require().NoError(err)
	s.Require().NotNil(workingHash)

	commitHash, err := s.rootStore.Commit()
	s.Require().NoError(err)
	s.Require().NotNil(commitHash)
	s.Require().Equal(workingHash, commitHash)

	// ensure the proof is non-nil for the corresponding version
	p, err = s.rootStore.GetProof("", 1, []byte("foo"))
	s.Require().NoError(err)
	s.Require().NotNil(p)
	s.Require().Equal([]byte("foo"), p.GetExist().Key)
	s.Require().Equal([]byte("bar"), p.GetExist().Value)
}

func (s *RootStoreTestSuite) TestBranch() {
	// write and commit a changeset
	bs := s.rootStore.GetKVStore("")
	bs.Set([]byte("foo"), []byte("bar"))

	workingHash, err := s.rootStore.WorkingHash()
	s.Require().NoError(err)
	s.Require().NotNil(workingHash)

	commitHash, err := s.rootStore.Commit()
	s.Require().NoError(err)
	s.Require().NotNil(commitHash)
	s.Require().Equal(workingHash, commitHash)

	// branch the root store
	rs2 := s.rootStore.Branch()

	// ensure we can perform reads which pass through to the original root store
	bs2 := rs2.GetKVStore("")
	s.Require().Equal([]byte("bar"), bs2.Get([]byte("foo")))

	// make a change to the branched root store
	bs2.Set([]byte("foo"), []byte("updated_bar"))

	// ensure the original root store is not modified
	s.Require().Equal([]byte("bar"), bs.Get([]byte("foo")))

	// write changes
	rs2.Write()

	// ensure changes are reflected in the original root store
	s.Require().Equal([]byte("updated_bar"), bs.Get([]byte("foo")))
}

func (s *RootStoreTestSuite) TestMultiBranch() {
	// write and commit a changeset
	bs := s.rootStore.GetKVStore("")
	bs.Set([]byte("foo"), []byte("bar"))

	workingHash, err := s.rootStore.WorkingHash()
	s.Require().NoError(err)
	s.Require().NotNil(workingHash)

	commitHash, err := s.rootStore.Commit()
	s.Require().NoError(err)
	s.Require().NotNil(commitHash)
	s.Require().Equal(workingHash, commitHash)

	// create multiple branches of the root store
	var branchedRootStores []store.BranchedRootStore
	for i := 0; i < 5; i++ {
		branchedRootStores = append(branchedRootStores, s.rootStore.Branch())
	}

	// get the last branched root store
	rs2 := branchedRootStores[4]

	// ensure we can perform reads which pass through to the original root store
	bs2 := rs2.GetKVStore("")
	s.Require().Equal([]byte("bar"), bs2.Get([]byte("foo")))

	// make a change to the branched root store
	bs2.Set([]byte("foo"), []byte("updated_bar"))

	// ensure the original root store is not modified
	s.Require().Equal([]byte("bar"), bs.Get([]byte("foo")))

	// write changes
	rs2.Write()

	// ensure changes are reflected in the original root store
	s.Require().Equal([]byte("updated_bar"), bs.Get([]byte("foo")))
}

func (s *RootStoreTestSuite) TestCommit() {
	lv, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Zero(lv)

	// branch the root store
	rs2 := s.rootStore.Branch()

	// perform changes
	bs2 := rs2.GetKVStore("")
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		bs2.Set([]byte(key), []byte(val))
	}

	// write to the branched root store, which will flush to the parent root store
	rs2.Write()

	// committing w/o calling WorkingHash should error
	_, err = s.rootStore.Commit()
	s.Require().Error(err)

	// execute WorkingHash and Commit
	wHash, err := s.rootStore.WorkingHash()
	s.Require().NoError(err)

	cHash, err := s.rootStore.Commit()
	s.Require().NoError(err)
	s.Require().Equal(wHash, cHash)

	// ensure latest version is updated
	lv, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), lv)

	// ensure the root KVStore is cleared
	s.Require().Empty(s.rootStore.(*Store).rootKVStore.GetChangeset().Pairs)

	// perform reads on the updated root store
	bs := s.rootStore.GetKVStore("")
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		s.Require().Equal([]byte(val), bs.Get([]byte(key)))
	}
}
