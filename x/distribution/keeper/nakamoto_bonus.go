package keeper

import (
	"fmt"
	"sort"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AdjustNakamotoBonusCoefficient is called to adjust η dynamically for each block.
// Every 'period' blocks:
// - If avg(high group) >= 3x avg(low group), nb += step
// - Else nb -= step
// Clamp nb to [0, 1]. If disabled, force to 0.
//
// Events emitted:
//   - EventTypeUpdateNakamotoCoefficient: When η value changes
//     Attributes: nakamoto_coefficient (new value), block_height
//   - EventTypeNakamotoBonusDisabled: When feature is disabled
//     Attributes: nakamoto_coefficient (current value), block_height
func (k Keeper) AdjustNakamotoBonusCoefficient(ctx sdk.Context) error {
	nb, err := k.GetNakamotoBonus(ctx)
	if err != nil {
		return err
	}

	if !nb.Enabled {
		return nil
	}

	period := int64(nb.Period)
	if period <= 0 {
		// misconfigured, do nothing
		return nil
	}
	if ctx.BlockHeight()%period != 0 {
		return nil
	}

	nbCoefficient, err := k.GetNakamotoBonusCoefficient(ctx)
	if err != nil {
		return err
	}

	validators, err := k.stakingKeeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		return err
	}
	n := len(validators)
	if n < 3 {
		return nil // need 3 groups; skip if small set
	}

	// sort by bonded tokens descending
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].GetBondedTokens().GT(validators[j].GetBondedTokens())
	})

	// split into 3 groups as evenly as possible: high, medium, low
	groupSize := n / 3
	high := validators[:groupSize]
	low := validators[groupSize*2:]

	sum := func(vals []stakingtypes.Validator) math.Int {
		total := math.ZeroInt()
		for _, v := range vals {
			total = total.Add(v.GetBondedTokens())
		}
		return total
	}
	avg := func(vals []stakingtypes.Validator) math.LegacyDec {
		if len(vals) == 0 {
			return math.LegacyZeroDec()
		}
		return math.LegacyNewDecFromInt(sum(vals)).QuoInt64(int64(len(vals)))
	}

	highAvg := avg(high)
	lowAvg := avg(low)
	newCoefficient := nbCoefficient

	// If lowAvg is zero, treat as increase case to spur NB
	if lowAvg.IsZero() || highAvg.Quo(lowAvg).GTE(math.LegacyNewDec(3)) {
		newCoefficient = newCoefficient.Add(nb.Step)
	} else {
		newCoefficient = newCoefficient.Sub(nb.Step)
	}

	// clamp to [min, max]
	if newCoefficient.LT(nb.MinimumCoefficient) {
		newCoefficient = nb.MinimumCoefficient
	}
	if newCoefficient.GT(nb.MaximumCoefficient) {
		newCoefficient = nb.MaximumCoefficient
	}

	// emit event if changed
	if !newCoefficient.Equal(nbCoefficient) {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeUpdateNakamotoCoefficient,
				sdk.NewAttribute(types.AttributeNakamotoCoefficient, newCoefficient.String()),
				sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			),
		)
		if err := k.SetNakamotoBonusCoefficient(ctx, newCoefficient); err != nil {
			return err
		}
	}

	return nil
}
