package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// GetMinInitialDeposit returns the (dynamic) minimum initial deposit currently required for
// proposal submission
func (keeper Keeper) GetMinInitialDeposit(ctx sdk.Context) sdk.Coins {
	lastMinInitialDeposit, err := keeper.getMinInitialDeposit(ctx)
	if err != nil {
		keeper.Logger(ctx).Error("failed to get last min initial deposit", "error", err)
		return sdk.NewCoins()
	}

	return lastMinInitialDeposit.Value
}

func (keeper Keeper) getMinInitialDeposit(ctx context.Context) (v1.LastMinDeposit, error) {
	lastMinDeposit, err := keeper.LastMinInitialDeposit.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		// if LastMinInitialDeposit is empty it means it was never set,
		// so we return the floor value
		params, err := keeper.Params.Get(ctx)
		if err != nil {
			return v1.LastMinDeposit{}, fmt.Errorf("failed to get params: %w", err)
		}

		lastMinDeposit.Value = params.MinInitialDepositThrottler.GetFloorValue()
		lastMinDeposit.Time = &time.Time{}
	} else if err != nil {
		return v1.LastMinDeposit{}, fmt.Errorf("failed to get min deposit: %w", err)
	}

	return lastMinDeposit, nil
}

// UpdateMinInitialDeposit updates the min initial deposit required for proposal submission
func (keeper Keeper) UpdateMinInitialDeposit(ctx context.Context, checkElapsedTime bool) {
	logger := keeper.Logger(ctx)

	params, err := keeper.Params.Get(ctx)
	if err != nil {
		logger.Error("failed to get params", "error", err)
		return
	}

	lastMinInitialDeposit, err := keeper.getMinInitialDeposit(ctx)
	if err != nil {
		logger.Error("failed to get last min initial deposit", "error", err)
		return
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	tick := params.MinInitialDepositThrottler.UpdatePeriod

	if checkElapsedTime && sdkCtx.BlockTime().Sub(*lastMinInitialDeposit.Time).Nanoseconds() < tick.Nanoseconds() {
		return
	}

	minInitialDepositFloor := sdk.Coins(params.MinInitialDepositThrottler.FloorValue)
	targetInactiveProposals := math.NewIntFromUint64(params.MinInitialDepositThrottler.TargetProposals)
	k := params.MinInitialDepositThrottler.DecreaseSensitivityTargetDistance
	var alpha math.LegacyDec

	var countInactiveProposals uint64
	err = keeper.InactiveProposalsQueue.Walk(ctx, nil, func(key collections.Pair[time.Time, uint64], _ uint64) (bool, error) {
		countInactiveProposals++
		return false, nil
	})
	if err != nil {
		keeper.Logger(ctx).Error("failed to update last min initial deposit", "error", err)
		return
	}

	numInactiveProposals := math.NewIntFromUint64(countInactiveProposals)
	if numInactiveProposals.GTE(targetInactiveProposals) {
		if checkElapsedTime {
			// no time-based increases
			return
		}
		alpha = math.LegacyMustNewDecFromStr(params.MinInitialDepositThrottler.IncreaseRatio)
	} else {
		distance := numInactiveProposals.Sub(targetInactiveProposals)
		if !checkElapsedTime {
			// decreases can only happen due to time-based updates
			// and if the number of active proposals is below the target
			return
		}
		alpha = math.LegacyMustNewDecFromStr(params.MinInitialDepositThrottler.DecreaseRatio)
		// ApproxRoot is here being called on a relatively small positive
		// integer (when distance < 0, ApproxRoot will return
		// `|distance|.ApproxRoot(k) * -1`) with a value of k expected to also
		// be relatively small (<= 100).
		// This is a safe operation and should not error.
		b, err := distance.ToLegacyDec().ApproxRoot(k)
		if err != nil {
			// in case of error bypass the sensitivity, i.e. assume k = 1
			b = distance.ToLegacyDec()
			logger.Error("failed to calculate ApproxRoot for min initial deposit",
				"error", err,
				"distance", distance.String(),
				"k", k,
				"fallback", "using k=1")
		}
		alpha = alpha.Mul(b)
	}
	percChange := math.LegacyOneDec().Add(alpha)
	newMinInitialDeposit := v1.GetNewMinDeposit(minInitialDepositFloor, lastMinInitialDeposit.Value, percChange)

	time := sdkCtx.BlockTime()
	if err := keeper.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
		Value: newMinInitialDeposit,
		Time:  &time,
	}); err != nil {
		logger.Error("failed to set last min initial deposit",
			"error", err,
			"newMinInitialDeposit", newMinInitialDeposit.String(),
			"lastMinInitialDeposit", lastMinInitialDeposit.String())
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMinInitialDepositChange,
			sdk.NewAttribute(types.AttributeKeyNewMinInitialDeposit, newMinInitialDeposit.String()),
			sdk.NewAttribute(types.AttributeKeyLastMinInitialDeposit, lastMinInitialDeposit.String()),
		),
	)
}
