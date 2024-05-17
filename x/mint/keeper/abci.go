package keeper

import (
	"context"

	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

// BeginBlocker mints new tokens for the previous block.
func (k Keeper) BeginBlocker(ctx context.Context, ic types.MintFn) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter & params
	minter, err := k.Minter.Get(ctx)
	if err != nil {
		return err
	}

	err = ic(ctx, k.Environment, &minter)
	if err != nil {
		return err
	}

	return k.Minter.Set(ctx, minter)
}
