package keeper

import "context"

// Migrator is a struct for handling in-place state migrations.
type Migrator struct {
	keeper *Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(k *Keeper) Migrator {
	return Migrator{
		keeper: k,
	}
}

// Migrate1to2 migrates the x/crisis module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/crisis
// module state.
func (m Migrator) Migrate1to2(ctx context.Context) error {
	return nil
}
