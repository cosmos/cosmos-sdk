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

// GetMinDeposit returns the (dynamic) minimum deposit currently required for a proposal
func (keeper Keeper) GetMinDeposit(ctx context.Context) sdk.Coins {
	lastMinDeposit, err := keeper.getMinDeposit(ctx)
	if err != nil {
		keeper.Logger(ctx).Error("failed to get last min deposit", "error", err)
		return sdk.NewCoins()
	}

	return lastMinDeposit.Value
}

func (keeper Keeper) getMinDeposit(ctx context.Context) (v1.LastMinDeposit, error) {
	lastMinDeposit, err := keeper.LastMinDeposit.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		// if LastMinDeposit is empty it means it was never set,
		// so we return the floor value
		params, err := keeper.Params.Get(ctx)
		if err != nil {
			return v1.LastMinDeposit{}, fmt.Errorf("failed to get params: %w", err)
		}

		lastMinDeposit.Value = params.MinDepositThrottler.GetFloorValue()
		lastMinDeposit.Time = &time.Time{}
	} else if err != nil {
		return v1.LastMinDeposit{}, fmt.Errorf("failed to get last min deposit: %w", err)
	}

	return lastMinDeposit, nil
}

// UpdateMinDeposit updates the minimum deposit required for a proposal
func (keeper Keeper) UpdateMinDeposit(ctx context.Context, checkElapsedTime bool) {
	logger := keeper.Logger(ctx)

	params, err := keeper.Params.Get(ctx)
	if err != nil {
		keeper.Logger(ctx).Error("failed to get params", "error", err)
		return
	}

	tick := params.MinDepositThrottler.UpdatePeriod
	lastMinDeposit, err := keeper.getMinDeposit(ctx)
	if err != nil {
		keeper.Logger(ctx).Error("failed to get last min deposit", "error", err)
		return
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if checkElapsedTime && sdkCtx.BlockTime().Sub(*lastMinDeposit.Time).Nanoseconds() < tick.Nanoseconds() {
		return
	}

	minDepositFloor := sdk.Coins(params.MinDepositThrottler.FloorValue)
	targetActiveProposals := math.NewIntFromUint64(params.MinDepositThrottler.TargetActiveProposals)
	k := params.MinDepositThrottler.DecreaseSensitivityTargetDistance
	var alpha math.LegacyDec

	var countActiveProposals uint64
	err = keeper.ActiveProposalsQueue.Walk(ctx, nil, func(collections.Pair[time.Time, uint64], uint64) (bool, error) {
		countActiveProposals++
		return false, nil
	})
	if err != nil {
		keeper.Logger(ctx).Error("failed to update last min deposit", "error", err)
		return
	}

	numActiveProposals := math.NewIntFromUint64(countActiveProposals)
	if numActiveProposals.GTE(targetActiveProposals) {
		if checkElapsedTime {
			// no time-based increases
			return
		}
		alpha = math.LegacyMustNewDecFromStr(params.MinDepositThrottler.IncreaseRatio)
	} else {
		distance := numActiveProposals.Sub(targetActiveProposals)
		if !checkElapsedTime {
			// decreases can only happen due to time-based updates
			// and if the number of active proposals is below the target
			return
		}
		alpha = math.LegacyMustNewDecFromStr(params.MinDepositThrottler.DecreaseRatio)
		// ApproxRoot is here being called on a relatively small positive
		// integer (when distance < 0, ApproxRoot will return
		// `|distance|.ApproxRoot(k) * -1`) with a value of k expected to also
		// be relatively small (<= 100).
		// This is a safe operation and should not error.
		b, err := distance.ToLegacyDec().ApproxRoot(k)
		if err != nil {
			// in case of error bypass the sensitivity, i.e. assume k = 1
			b = distance.ToLegacyDec()
			logger.Error("failed to calculate ApproxRoot for min deposit",
				"error", err,
				"distance", distance.String(),
				"k", k,
				"fallback", "using k=1")
		}
		alpha = alpha.Mul(b)
	}
	percChange := math.LegacyOneDec().Add(alpha)
	newMinDeposit := v1.GetNewMinDeposit(minDepositFloor, lastMinDeposit.Value, percChange)

	time := sdkCtx.BlockTime()
	if err := keeper.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
		Value: newMinDeposit,
		Time:  &time,
	}); err != nil {
		logger.Error("failed to set last min deposit",
			"error", err,
			"newMinDeposit", newMinDeposit.String(),
			"time", time.String())
		return
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMinDepositChange,
			sdk.NewAttribute(types.AttributeKeyNewMinDeposit, newMinDeposit.String()),
			sdk.NewAttribute(types.AttributeKeyLastMinDeposit, lastMinDeposit.String()),
		),
	)
}
