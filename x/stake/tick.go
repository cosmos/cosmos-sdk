package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
)

const (
	hrsPerYr  = 8766 // as defined by a julian year of 365.25 days
	precision = 1000000000
)

var hrsPerYrRat = sdk.NewRat(hrsPerYr) // as defined by a julian year of 365.25 days

// Tick - called at the end of every block
func (k Keeper) Tick(ctx sdk.Context) (change []abci.Validator) {
	p := k.GetPool(ctx)

	// Process Validator Provisions
	blockTime := ctx.BlockHeader().Time // XXX assuming in seconds, confirm
	if p.InflationLastTime+blockTime >= 3600 {
		p.InflationLastTime = blockTime
		p = k.processProvisions(ctx)
	}

	// save the params
	k.setPool(ctx, p)

	// reset the counter
	k.setCounter(ctx, 0)

	change = k.getAccUpdateValidators(ctx)

	return
}

// process provisions for an hour period
func (k Keeper) processProvisions(ctx sdk.Context) Pool {

	pool := k.GetPool(ctx)
	pool.Inflation = k.nextInflation(ctx)

	// Because the validators hold a relative bonded share (`GlobalStakeShare`), when
	// more bonded tokens are added proportionally to all validators the only term
	// which needs to be updated is the `BondedPool`. So for each previsions cycle:

	provisions := pool.Inflation.Mul(sdk.NewRat(pool.TotalSupply)).Quo(hrsPerYrRat).Evaluate()
	pool.BondedPool += provisions
	pool.TotalSupply += provisions
	return pool
}

// get the next inflation rate for the hour
func (k Keeper) nextInflation(ctx sdk.Context) (inflation sdk.Rat) {

	params := k.GetParams(ctx)
	pool := k.GetPool(ctx)
	// The target annual inflation rate is recalculated for each previsions cycle. The
	// inflation is also subject to a rate change (positive of negative) depending or
	// the distance from the desired ratio (67%). The maximum rate change possible is
	// defined to be 13% per year, however the annual inflation is capped as between
	// 7% and 20%.

	// (1 - bondedRatio/GoalBonded) * InflationRateChange
	inflationRateChangePerYear := sdk.OneRat().Sub(pool.bondedRatio().Quo(params.GoalBonded)).Mul(params.InflationRateChange)
	inflationRateChange := inflationRateChangePerYear.Quo(hrsPerYrRat)

	// increase the new annual inflation for this next cycle
	inflation = pool.Inflation.Add(inflationRateChange)
	if inflation.GT(params.InflationMax) {
		inflation = params.InflationMax
	}
	if inflation.LT(params.InflationMin) {
		inflation = params.InflationMin
	}

	return inflation.Round(precision)
}
