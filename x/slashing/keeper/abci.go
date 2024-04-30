package keeper

import (
	"context"

	consensusv1 "cosmossdk.io/api/cosmos/consensus/v1"
	"cosmossdk.io/x/slashing/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func (k Keeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// Iterate over all the validators which *should* have signed this block
	// store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	res := consensusv1.QueryCometInfoResponse{}
	if err := k.RouterService.QueryRouterService().InvokeTyped(ctx, &consensusv1.MsgCometInfoRequest{}, &res); err != nil {
		return err
	}
	for _, vote := range res.CometInfo.LastCommit.Votes {
		err := k.HandleValidatorSignatureWithParams(ctx, params, vote.Validator.Address, vote.Validator.Power, vote.BlockIdFlag)
		if err != nil {
			return err
		}
	}
	return nil
}
