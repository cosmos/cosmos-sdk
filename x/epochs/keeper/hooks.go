package keeper

import (
	"context"
)

// AfterEpochEnd gets called at the end of the epoch, end of epoch is the timestamp of first block produced after epoch duration.
func (k Keeper) AfterEpochEnd(ctx context.Context, identifier string, epochNumber int64) error {
	return k.hooks.AfterEpochEnd(ctx, identifier, epochNumber, k.environment)
}

// BeforeEpochStart new epoch is next block of epoch end block
func (k Keeper) BeforeEpochStart(ctx context.Context, identifier string, epochNumber int64) error {
	return k.hooks.BeforeEpochStart(ctx, identifier, epochNumber, k.environment)
}
