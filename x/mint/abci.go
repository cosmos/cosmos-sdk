package mint

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, ic types.InflationCalculationFn) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// fetch collected fees
	collectedFeeCoin := k.CountCollectedFees(ctx, params.MintDenom)

	// recalculate inflation rate
	bondedRatio := k.BondedRatio(ctx)
	minter.Inflation = ic(ctx, minter, params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, k.BondedTokenSupply(ctx))

	k.SetMinter(ctx, minter)

	// mint coins, update supply
	neededCoin := minter.BlockProvision(params)
	mintedCoin := sdk.NewCoin(params.MintDenom, sdk.ZeroInt())
	burnedCoin := sdk.NewCoin(params.MintDenom, sdk.ZeroInt())

	if collectedFeeCoin.IsLT(neededCoin) {
		// if the fee collector has not collected enough fees to meet the
		// staking incentive goals, mint enough to meet.
		mintedCoin = neededCoin.Sub(collectedFeeCoin)
		mintedCoins := sdk.NewCoins(mintedCoin)

		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		// send the minted coins to the fee collector account
		err = k.AddCollectedFees(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		if mintedCoin.Amount.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
		}

	} else {
		// if the fee collector has collected more fees than are needed to meet the
		// staking incentive goals, burn the rest.
		burnedCoin = collectedFeeCoin.Sub(neededCoin)
		burnedCoins := sdk.NewCoins(burnedCoin)
		err := k.BurnFees(ctx, burnedCoins)
		if err != nil {
			panic(err)
		}
	}

	mintEvent := types.MintIncentiveTokens{
		BondedRatio:      bondedRatio,
		Inflation:        minter.Inflation,
		AnnualProvisions: minter.AnnualProvisions,
		NeededAmount:     neededCoin.Amount.Uint64(),
		CollectedAmount:  collectedFeeCoin.Amount.Uint64(),
		MintedAmount:     mintedCoin.Amount.Uint64(),
		BurnedAmount:     burnedCoin.Amount.Uint64(),
	}
	if err := ctx.EventManager().EmitTypedEvent(&mintEvent); err != nil {
		k.Logger(ctx).Error("error emitting event",
			"error", err,
			"event", mintEvent)
	}
}
