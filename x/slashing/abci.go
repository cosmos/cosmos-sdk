package slashing

import (
	"context"
	"time"

	"cosmossdk.io/x/slashing/keeper"
	"cosmossdk.io/x/slashing/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// Iterate over all the validators which *should* have signed this block
	// store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx) // TODO remove by passing the comet service
	for _, vote := range sdkCtx.CometInfo().LastCommit.Votes {
		err := k.HandleValidatorSignatureWithParams(ctx, params, vote.Validator.Address, vote.Validator.Power, vote.BlockIDFlag)
		if err != nil {
			return err
		}
	}
	return nil
}
