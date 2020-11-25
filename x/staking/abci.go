package staking

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	if ctx.BlockHeight()%10 == 1 {
		defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
		k.TrackHistoricalInfo(ctx)
	}
}

// Called every block, update validator set
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	if ctx.BlockHeight()%10 == 0 { // TODO should update hardcoded 10 to params.EpochInterval (epoch_interval)
		defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

		// execute all epoch actions
		iterator := k.GetEpochActionsIterator()
	
		for ; iterator.Valid(); iterator.Next() {
			var msg sdk.Msg
			bz := iterator.Value()
			k.cdc.MustUnmarshalBinaryBare(bz, &msg)

			switch msg := msg.(type) {
			case *types.MsgEditValidator:
				res, err := k.EpochEditValidator(sdk.WrapSDKContext(ctx), msg)

			default:
			}
		}
		// dequeue all epoch actions after run
		k.DequeueEpochActions()

		return k.BlockValidatorUpdates(ctx)
	}
	return []abci.ValidatorUpdate{}
}
