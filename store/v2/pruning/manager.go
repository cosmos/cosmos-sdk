package pruning

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

const batchFlushThreshold = 1 << 16 // 64KB

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

// PruneKVStores prunes KVStores which are needed to be pruned by upgrading the
// store key.
func (m *Manager) PruneKVStores(kvStores []corestore.KVStoreWithBatch) error {
	for _, kvStore := range kvStores {
		if kvStore == nil {
			continue
		}
		if err := func() error {
			batch := kvStore.NewBatch()
			iter, err := kvStore.Iterator(nil, nil)
			if err != nil {
				return err
			}
			defer func() {
				_ = iter.Close()
				_ = batch.Close()
			}()

			for ; iter.Valid(); iter.Next() {
				if err := batch.Delete(iter.Key()); err != nil {
					return err
				}
				bs, err := batch.GetByteSize()
				if err != nil {
					return err
				}
				if bs > batchFlushThreshold {
					if err := batch.Write(); err != nil {
						return err
					}
					if err := batch.Close(); err != nil {
						return err
					}
					batch = kvStore.NewBatch()
				}
			}
			if err := batch.Write(); err != nil {
				return err
			}
			if err := batch.Close(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}
