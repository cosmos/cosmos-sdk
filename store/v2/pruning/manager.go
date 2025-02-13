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
}

// NewManager creates a new Pruning Manager.
func NewManager(scPruner store.Pruner, scPruningOption *store.PruningOption) *Manager {
	return &Manager{
		scPruner:        scPruner,
		scPruningOption: scPruningOption,
	}
}

// Prune prunes the SC and SS to the provided version.
//
// NOTE: It can be called outside the store manually.
func (m *Manager) Prune(version uint64) error {
	// Prune the SC.
	if m.scPruningOption != nil {
		if prune, pruneTo := m.scPruningOption.ShouldPrune(version); prune {
			if err := m.scPruner.Prune(pruneTo); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) signalPruning(pause bool) {
	if scPausablePruner, ok := m.scPruner.(store.PausablePruner); ok {
		scPausablePruner.PausePruning(pause)
	}
}

func (m *Manager) PausePruning() {
	m.signalPruning(true)
}

func (m *Manager) ResumePruning(version uint64) error {
	m.signalPruning(false)
	return m.Prune(version)
}
