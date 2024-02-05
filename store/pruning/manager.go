package pruning

import (
	"sync/atomic"

	"cosmossdk.io/log"
)

// Manager is an abstraction to handle pruning of SS and SC backends.
type Manager struct {
	logger            log.Logger
	storageOpts       Options
	commitmentOpts    Options
	storageVersion    atomic.Uint64 // the pruning version of the storage in progress
	commitmentVersion atomic.Uint64 // the pruning version of the commitment in progress

	stateStorage    PruningStore
	stateCommitment PruningStore
}

// NewManager creates a new Manager instance.
func NewManager(
	logger log.Logger,
	ss PruningStore,
	sc PruningStore,
	storageOpts, commitmentOpts Options,
) *Manager {
	return &Manager{
		stateStorage:    ss,
		stateCommitment: sc,
		logger:          logger,
		storageOpts:     storageOpts,
		commitmentOpts:  commitmentOpts,
	}
}

// Prune prunes the state storage and state commitment.
// It will check the pruning conditions and prune if necessary.
func (m *Manager) Prune(version uint64) {
	// storage pruning
	if m.storageOpts.Interval > 0 && version > m.storageOpts.KeepRecent && version%m.storageOpts.Interval == 0 {
		pruneVersion := version - m.storageOpts.KeepRecent - 1
		m.pruneStorage(pruneVersion)
	}

	// commitment pruning
	if m.commitmentOpts.Interval > 0 && version > m.commitmentOpts.KeepRecent && version%m.commitmentOpts.Interval == 0 {
		pruneVersion := version - m.commitmentOpts.KeepRecent - 1
		m.pruneCommitment(pruneVersion)
	}
}

func (m *Manager) pruneStorage(version uint64) {
	currentVersion := m.storageVersion.Load()
	if 0 < currentVersion {
		// pruning is already in progress
		m.logger.Info("pruning storage already in progress", "version", currentVersion)
		return
	}

	m.storageVersion.Store(version)
	m.logger.Debug("pruning storage store", "version", version)

	go func() {
		defer m.storageVersion.Store(0)
		if err := m.stateStorage.Prune(version); err != nil {
			m.logger.Error("failed to prune storage store", "err", err)
		}
	}()
}

func (m *Manager) pruneCommitment(version uint64) {
	currentVersion := m.commitmentVersion.Load()
	if 0 < currentVersion {
		// pruning is already in progress
		m.logger.Info("pruning commitment already in progress", "version", currentVersion)
		return
	}

	m.commitmentVersion.Store(version)
	m.logger.Debug("pruning commitment store", "version", version)

	go func() {
		defer m.commitmentVersion.Store(0)
		if err := m.stateCommitment.Prune(version); err != nil {
			m.logger.Error("failed to prune commitment store", "err", err)
		}
	}()
}
