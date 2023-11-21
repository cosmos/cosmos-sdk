package pruning

import (
	"fmt"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment/iavl"
	"cosmossdk.io/store/v2/storage/sqlite"
)

type PruningTestSuite struct {
	suite.Suite

	manager *Manager
	ss      store.VersionedDatabase
	sc      store.Committer
}

func TestPruningTestSuite(t *testing.T) {
	suite.Run(t, &PruningTestSuite{})
}

func (s *PruningTestSuite) SetupTest() {
	logger := log.NewNopLogger()
	if testing.Verbose() {
		logger = log.NewTestLogger(s.T())
	}

	ss, err := sqlite.New(s.T().TempDir())
	s.Require().NoError(err)

	sc := iavl.NewIavlTree(dbm.NewMemDB(), log.NewNopLogger(), iavl.DefaultConfig())

	s.manager = NewManager(logger, ss, sc)
	s.ss = ss
	s.sc = sc
}

func (s *PruningTestSuite) TearDownTest() {
	s.manager.Start()
	s.manager.Stop()
}

func (s *PruningTestSuite) TestPruning() {
	s.manager.SetCommitmentOptions(Options{4, 2, true})
	s.manager.SetStorageOptions(Options{3, 3, true})
	s.manager.Start()

	latestVersion := uint64(100)

	// write 10 batches
	for i := uint64(0); i < latestVersion; i++ {
		version := i + 1

		cs := store.NewChangeset()
		cs.Add([]byte("key"), []byte(fmt.Sprintf("value%d", version)))

		err := s.sc.WriteBatch(cs)
		s.Require().NoError(err)

		_, err = s.sc.Commit()
		s.Require().NoError(err)

		err = s.ss.ApplyChangeset(version, cs)
		s.Require().NoError(err)
		s.manager.Prune(version)
	}

	// wait for pruning to finish
	s.manager.Stop()

	// check the store for the version 96
	val, err := s.ss.Get("", latestVersion-4, []byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("value96"), val)

	// check the store for the version 50
	val, err = s.ss.Get("", 50, []byte("key"))
	s.Require().Error(err)
	s.Require().Nil(val)

	// check the commitment for the version 96
	proof, err := s.sc.GetProof(latestVersion-4, []byte("key"))
	s.Require().NoError(err)
	s.Require().NotNil(proof.GetExist())

	// check the commitment for the version 95
	proof, err = s.sc.GetProof(latestVersion-5, []byte("key"))
	s.Require().Error(err)
	s.Require().Nil(proof)
}
