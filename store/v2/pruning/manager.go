package pruning

import (
	"cosmossdk.io/store/v2"
)

// Manager is a struct that manages the pruning of old versions of the SC and SS.
type Manager struct {
	// scPruner is the pruner for the SC.
	scPruner store.Pruner
	// scPruningOption are the pruning options for the SC.
	scPruningOption *store.PruningOption
	// ssPruner is the pruner for the SS.
	ssPruner store.Pruner
	// ssPruningOption are the pruning options for the SS.
	ssPruningOption *store.PruningOption
}

// NewManager creates a new Pruning Manager.
func NewManager(scPruner, ssPruner store.Pruner, scPruningOption, ssPruningOption *store.PruningOption) *Manager {
	return &Manager{
		scPruner:        scPruner,
		scPruningOption: scPruningOption,
		ssPruner:        ssPruner,
		ssPruningOption: ssPruningOption,
	}
}

// Prune prunes the SC and SS to the provided version.
//
// NOTE: It can be called outside of the store manually.
func (m *Manager) Prune(version uint64) error {
	// Prune the SC.
	if m.scPruningOption != nil {
		if prune, pruneTo := m.scPruningOption.ShouldPrune(version); prune {
			if err := m.scPruner.Prune(pruneTo); err != nil {
				return err
			}
		}
	}

	// Prune the SS.
	if m.ssPruningOption != nil {
		if prune, pruneTo := m.ssPruningOption.ShouldPrune(version); prune {
			if err := m.ssPruner.Prune(pruneTo); err != nil {
				return err
			}
		}
	}

	return nil
}

// SignalCommit signals to the manager that a commit has started or finished.
// It is used to trigger the pruning of the SC and SS.
// It pauses or resumes the pruning of the SC and SS if the pruner implements
// the PausablePruner interface.
func (m *Manager) SignalCommit(start bool, version uint64) error {
	if scPausablePruner, ok := m.scPruner.(store.PausablePruner); ok {
		scPausablePruner.PausePruning(start)
	}
	if ssPausablePruner, ok := m.ssPruner.(store.PausablePruner); ok {
		ssPausablePruner.PausePruning(start)
	}

	if !start {
		return m.Prune(version)
	}

	return nil
}
