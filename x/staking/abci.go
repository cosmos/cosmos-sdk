package staking

import (
	"fmt"
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
	EpochInterval := k.GetParams(ctx).EpochInterval
	if ctx.BlockHeight()%EpochInterval == 0 {
		defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

		// execute all epoch actions
		iterator := k.GetEpochActionsIteratorByEpochNumber(ctx, 0)

		for ; iterator.Valid(); iterator.Next() {
			msg := k.GetEpochActionByIterator(iterator)

			switch msg := msg.(type) {
			case *types.MsgEditValidator:
				// TODO what should we do if error happen for queued action?
				k.EpochEditValidator(ctx, msg)
			case *types.MsgDelegate:
				k.EpochDelegate(ctx, msg)
			case *types.MsgBeginRedelegate:
				k.EpochBeginRedelegate(ctx, msg)
			case *types.MsgUndelegate:
				k.EpochUndelegate(ctx, msg)
			default:
				panic(fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
			}
			// dequeue processed item
			k.DeleteByKey(ctx, iterator.Key())
		}

		// defer TODO should update epochNumber after epoch finish
		// This won't affect slashing module since slashing Endblocker run before staking module

		return k.EpochValidatorUpdates(ctx)
	}

	// run block validator updates for slashed, jailed validators
	return k.BlockValidatorUpdates(ctx)
}
