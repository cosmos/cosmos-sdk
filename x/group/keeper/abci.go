package keeper

import (
	"context"
)

// EndBlocker called at every block, updates proposal's `FinalTallyResult` and
// prunes expired proposals.
func (k Keeper) EndBlocker(ctx context.Context) error {
	if err := k.TallyProposalsAtVPEnd(ctx, k.environment); err != nil {
		return err
	}

	return k.PruneProposals(ctx, k.environment)
}
