package keeper

import (
	"context"

	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

// BeginBlocker mints new tokens for the previous block.
func (k Keeper) BeginBlocker(ctx context.Context, mintFn types.MintFn) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter & params
	minter, err := k.Minter.Get(ctx)
	if err != nil {
		return err
	}

	// we pass -1 as epoch number to indicate that this is not an epoch minting,
	// but a regular block minting.
	err = mintFn(ctx, k.Environment, &minter, -1)
	if err != nil {
		return err
	}

	return k.Minter.Set(ctx, minter)
}
