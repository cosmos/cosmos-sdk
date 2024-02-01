package pruning

import (
	"sync/atomic"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
)

// Manager is an abstraction to handle pruning of SS and SC backends.
type Manager struct {
	isStarted *atomic.Bool

	scPruner *pruner
	ssPruner *pruner
}

// NewManager creates a new Manager instance.
func NewManager(
	logger log.Logger,
	ss store.VersionedDatabase,
	sc store.Committer,
) *Manager {
	m := &Manager{
		isStarted: new(atomic.Bool),
		scPruner:  newPruner(logger, DefaultOptions(), sc, "commitment"),
		ssPruner:  newPruner(logger, DefaultOptions(), ss, "storage"),
	}

	return m
}

// SetStorageOptions sets the state storage options.
func (m *Manager) SetStorageOptions(opts Options) { m.ssPruner.setOptions(opts) }

// SetCommitmentOptions sets the state commitment options.
func (m *Manager) SetCommitmentOptions(opts Options) { m.scPruner.setOptions(opts) }

// Start starts the manager.
func (m *Manager) Start() {
	// if the CAS returns true then it means that the manager already started.
	if !m.isStarted.CompareAndSwap(false, true) {
		return
	}
}

// Stop stops the manager and waits for all goroutines to finish.
func (m *Manager) Stop() {
	// if the CAS returns false then it means the manager did never start.
	if !m.isStarted.CompareAndSwap(true, false) {
		return
	}
	for {
		if m.ssPruner.donePruning() && m.scPruner.donePruning() {
			return
		}
	}
}

// Prune prunes the state storage and state commitment.
// It will check the pruning conditions and prune if necessary.
func (m *Manager) Prune(height uint64) {
	if !m.isStarted.Load() {
		return
	}
	m.ssPruner.maybePrune(height)
	m.scPruner.maybePrune(height)
}

func newPruner(
	log log.Logger,
	opts Options,
	state interface{ Prune(height uint64) error },
	name string,
) *pruner {
	pruner := &pruner{
		logger:    log,
		opts:      atomic.Pointer[Options]{},
		isPruning: atomic.Bool{},
		state:     state,
		name:      name,
	}
	pruner.opts.Store(&opts)

	return pruner
}

type pruner struct {
	logger log.Logger

	opts      atomic.Pointer[Options]
	isPruning atomic.Bool

	state interface{ Prune(height uint64) error }

	name string // the name of the component being pruned (eg: state storage, commitment)
}

func (p *pruner) setOptions(opt Options) {
	p.opts.Store(&opt)
}

func (p *pruner) maybePrune(currentHeight uint64) {
	opts := p.opts.Load()
	if !opts.ShouldPrune(currentHeight) {
		return
	}

	heightToPrune := currentHeight - opts.KeepRecent - 1
	// check if we can prune in sync or not
	if opts.Sync {
		p.prune(heightToPrune)
		return
	}

	// if we cannot prune in sync we'll prune async, if  the sync goroutine is done pruning. This will attempt
	// to set isPruning from false to true, if it can't (because it's still true, hence blocked) then it will exit.
	if !p.isPruning.CompareAndSwap(false, true) {
		p.logger.Debug("storage pruning is still running; skipping", "name", p.name, "height", heightToPrune)
		return
	}

	go func() {
		p.prune(heightToPrune)
		p.isPruning.Store(false)
	}()
}

func (p *pruner) prune(height uint64) {
	p.logger.Debug("pruning state", "name", p.name, "height", height)
	err := p.state.Prune(height)
	if err != nil {
		p.logger.Error("pruning state", "name", p.name, "height", height, "err", err)
	}
}

func (p *pruner) donePruning() bool {
	return p.isPruning.Load() == false
}
