package slashing

import (
	"context"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/x/slashing/keeper"
	"cosmossdk.io/x/slashing/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx context.Context, k keeper.Keeper, cometService comet.Service) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// Iterate over all the validators which *should* have signed this block
	// store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}
	ci := cometService.CometInfo(ctx)
	for _, vote := range ci.LastCommit.Votes {
		err := k.HandleValidatorSignatureWithParams(ctx, params, vote.Validator.Address, vote.Validator.Power, vote.BlockIDFlag)
		if err != nil {
			return err
		}
	}
	return nil
}
