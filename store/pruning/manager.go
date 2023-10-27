package pruning

import (
	"sync"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
)

// Manager is an abstraction to handle the pruning logic.
type Manager struct {
	stateStore      store.VersionedDatabase
	stateCommitment store.Committer

	logger         log.Logger
	storeOpts      Option
	commitmentOpts Option

	chStore      chan struct{}
	chCommitment chan struct{}

	mtx sync.Mutex
}

// NewManager creates a new Manager instance.
func NewManager(
	stateStore store.VersionedDatabase,
	stateCommitment store.Committer,
	logger log.Logger,
) *Manager {
	return &Manager{
		stateStore:      stateStore,
		stateCommitment: stateCommitment,
		logger:          logger,
		storeOpts:       DefaultOptions(),
		commitmentOpts:  DefaultOptions(),
	}
}

// SetStoreOptions sets the store options.
func (m *Manager) SetStoreOptions(opts Option) {
	m.storeOpts = opts
}

// SetCommitmentOptions sets the commitment options.
func (m *Manager) SetCommitmentOptions(opts Option) {
	m.commitmentOpts = opts
}

// Start starts the manager.
func (m *Manager) Start() {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if !m.storeOpts.Sync {
		m.chStore = make(chan struct{}, 1)
		m.chStore <- struct{}{}
	}
	if !m.commitmentOpts.Sync {
		m.chCommitment = make(chan struct{}, 1)
		m.chCommitment <- struct{}{}
	}
}

// Stop stops the manager and waits for all goroutines to finish.
func (m *Manager) Stop() {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if !m.storeOpts.Sync {
		<-m.chStore
		close(m.chStore)
	}
	if !m.commitmentOpts.Sync {
		<-m.chCommitment
		close(m.chCommitment)
	}
}

// Prune prunes the state store and state commitment.
// It will not block the caller and just check if pruning is needed.
// If pruning is needed, it will spawn a goroutine to do the actual pruning.
func (m *Manager) Prune(height uint64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	// storage pruning
	if m.storeOpts.Interval > 0 && height > m.storeOpts.KeepRecent && height%m.storeOpts.Interval == 0 {
		pruneHeight := height - m.storeOpts.KeepRecent - 1
		if m.storeOpts.Sync {
			m.pruneStore(pruneHeight)
		} else {
			_, ok := <-m.chStore
			if ok {
				go func() {
					m.pruneStore(pruneHeight)
					m.chStore <- struct{}{}
				}()
			}
		}
	}

	// commitment pruning
	if m.commitmentOpts.Interval > 0 && height > m.commitmentOpts.KeepRecent && height%m.commitmentOpts.Interval == 0 {
		pruneHeight := height - m.commitmentOpts.KeepRecent - 1
		if m.commitmentOpts.Sync {
			m.pruneCommitment(pruneHeight)
		} else {

			_, ok := <-m.chCommitment
			if ok {
				go func() {
					m.pruneCommitment(height - m.commitmentOpts.KeepRecent - 1)
					m.chCommitment <- struct{}{}
				}()
			}
		}
	}
}

func (m *Manager) pruneStore(height uint64) {
	m.logger.Debug("pruning store", "height", height)

	err := m.stateStore.Prune(height)
	if err != nil {
		m.logger.Error("failed to prune store", "err", err)
	}
}

func (m *Manager) pruneCommitment(height uint64) {
	m.logger.Debug("pruning commitment", "height", height)

	err := m.stateCommitment.Prune(height)
	if err != nil {
		m.logger.Error("failed to prune commitment", "err", err)
	}
}
