package pruning

import (
	"sync"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
)

// Manager is an abstraction to handle pruning of SS and SC backends.
type Manager struct {
	mtx       sync.Mutex
	isStarted bool

	stateStorage    store.VersionedDatabase
	stateCommitment store.Committer

	logger         log.Logger
	storageOpts    Options
	commitmentOpts Options

	chStorage    chan struct{}
	chCommitment chan struct{}
}

// NewManager creates a new Manager instance.
func NewManager(
	logger log.Logger,
	ss store.VersionedDatabase,
	sc store.Committer,
) *Manager {
	return &Manager{
		stateStorage:    ss,
		stateCommitment: sc,
		logger:          logger,
		storageOpts:     DefaultOptions(),
		commitmentOpts:  DefaultOptions(),
	}
}

// SetStorageOptions sets the state storage options.
func (m *Manager) SetStorageOptions(opts Options) {
	m.storageOpts = opts
}

// SetCommitmentOptions sets the state commitment options.
func (m *Manager) SetCommitmentOptions(opts Options) {
	m.commitmentOpts = opts
}

// Start starts the manager.
func (m *Manager) Start() {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.isStarted {
		return
	}
	m.isStarted = true

	if !m.storageOpts.Sync {
		m.chStorage = make(chan struct{}, 1)
		m.chStorage <- struct{}{}
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

	if !m.isStarted {
		return
	}
	m.isStarted = false

	if !m.storageOpts.Sync {
		<-m.chStorage
		close(m.chStorage)
	}
	if !m.commitmentOpts.Sync {
		<-m.chCommitment
		close(m.chCommitment)
	}
}

// Prune prunes the state storage and state commitment.
// It will not block the caller and just check if pruning is needed.
// If pruning is needed, it will spawn a goroutine to do the actual pruning.
func (m *Manager) Prune(height uint64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if !m.isStarted {
		return
	}

	// storage pruning
	if m.storageOpts.Interval > 0 && height > m.storageOpts.KeepRecent && height%m.storageOpts.Interval == 0 {
		pruneHeight := height - m.storageOpts.KeepRecent - 1
		if m.storageOpts.Sync {
			m.pruneStorage(pruneHeight)
		} else {
			_, ok := <-m.chStorage
			if ok {
				go func() {
					m.pruneStorage(pruneHeight)
					m.chStorage <- struct{}{}
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

func (m *Manager) pruneStorage(height uint64) {
	m.logger.Debug("pruning state storage", "height", height)

	err := m.stateStorage.Prune(height)
	if err != nil {
		m.logger.Error("failed to prune state storage", "err", err)
	}
}

func (m *Manager) pruneCommitment(height uint64) {
	m.logger.Debug("pruning commitment", "height", height)

	err := m.stateCommitment.Prune(height)
	if err != nil {
		m.logger.Error("failed to prune commitment", "err", err)
	}
}
