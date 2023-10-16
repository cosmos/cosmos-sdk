package root

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	"cosmossdk.io/store/v2/storage/sqlite"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"
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
	s.Require().Nil(p)
	s.Require().Error(err)
}
