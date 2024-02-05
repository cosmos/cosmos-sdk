package pruning

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
)

type PruningTestSuite struct {
	suite.Suite

	manager *Manager
	ss      *busyPruningStore
	sc      *busyPruningStore
}

func (s *PruningTestSuite) SetupTest() {
	logger := log.NewNopLogger()
	if testing.Verbose() {
		logger = log.NewTestLogger(s.T())
	}

	s.ss = newBusyPruningStore()
	s.sc = newBusyPruningStore()
	s.manager = NewManager(logger, s.ss, s.sc, Options{3, 3}, Options{4, 2})
}

// waitForPruning waits for the pruning to finish, test only.
// It returns true if the pruning is finished, false otherwise.
func waitForPruning(m *Manager) bool {
	for i := 0; i < 10; i++ {
		if m.storageVersion.Load() == 0 && m.commitmentVersion.Load() == 0 {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func (s *PruningTestSuite) TestPruning() {
	// Prune at version 7, nothing should happen because the version 7 is not a multiple of 3 or 2.
	s.manager.Prune(7)
	s.Require().True(waitForPruning(s.manager), "pruning did not finish")
	s.Equal(s.ss.count, 0)
	s.Equal(s.sc.count, 0)

	// Prune at version 8, only the commitment should be pruned.
	s.manager.Prune(8)
	// The pruning is not finished yet, so the count should be 0.
	s.Require().False(waitForPruning(s.manager))
	s.Equal(s.sc.count, 0)
	// Finish the pruning and check the count.
	s.sc.finishPruning()
	s.Require().True(waitForPruning(s.manager))
	s.Equal(s.ss.count, 0)
	s.Equal(s.sc.count, 1)

	// Prune at version 9, only the storage should be pruned.
	s.manager.Prune(9)
	s.Require().False(waitForPruning(s.manager))
	s.Equal(s.ss.count, 0)
	s.ss.finishPruning()
	s.Require().True(waitForPruning(s.manager))
	s.Equal(s.ss.count, 1)
	s.Equal(s.sc.count, 1)

	// Prune at version 12, both the storage and the commitment should be pruned.
	s.manager.Prune(12)
	s.ss.finishPruning()
	s.sc.finishPruning()
	s.Require().True(waitForPruning(s.manager))
	s.Equal(s.ss.count, 2)
	s.Equal(s.sc.count, 2)
}

func TestPruningTestSuite(t *testing.T) {
	suite.Run(t, &PruningTestSuite{})
}
