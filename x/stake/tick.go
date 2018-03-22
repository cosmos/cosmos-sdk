package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
)

// Tick - called at the end of every block
func Tick(ctx sdk.Context, k Keeper) (change []*abci.Validator, err error) {

	// retrieve params
	params := k.getParams(ctx)
	p := k.getPool(ctx)
	height := ctx.BlockHeight()

	// Process Validator Provisions
	// XXX right now just process every 5 blocks, in new SDK make hourly
	if p.InflationLastTime+5 <= height {
		p.InflationLastTime = height
		processProvisions(ctx, k, p, params)
	}

	newVals := k.getValidators(ctx, params.MaxValidators)
	// XXX determine change from old validators, set to change
	_ = newVals
	return change, nil
}

var hrsPerYr = sdk.NewRat(8766) // as defined by a julian year of 365.25 days

// process provisions for an hour period
func processProvisions(ctx sdk.Context, k Keeper, p Pool, params Params) {

	p.Inflation = nextInflation(p, params).Round(1000000000)

	// Because the validators hold a relative bonded share (`GlobalStakeShare`), when
	// more bonded tokens are added proportionally to all validators the only term
	// which needs to be updated is the `BondedPool`. So for each previsions cycle:

	provisions := p.Inflation.Mul(sdk.NewRat(p.TotalSupply)).Quo(hrsPerYr).Evaluate()
	p.BondedPool += provisions
	p.TotalSupply += provisions

	// XXX XXX XXX XXX XXX XXX XXX XXX XXX
	// XXX Mint them to the hold account
	// XXX XXX XXX XXX XXX XXX XXX XXX XXX

	// save the params
	k.setPool(ctx, p)
}

// get the next inflation rate for the hour
func nextInflation(p Pool, params Params) (inflation sdk.Rat) {

	// The target annual inflation rate is recalculated for each previsions cycle. The
	// inflation is also subject to a rate change (positive of negative) depending or
	// the distance from the desired ratio (67%). The maximum rate change possible is
	// defined to be 13% per year, however the annual inflation is capped as between
	// 7% and 20%.

	// (1 - bondedRatio/GoalBonded) * InflationRateChange
	inflationRateChangePerYear := sdk.OneRat.Sub(p.bondedRatio().Quo(params.GoalBonded)).Mul(params.InflationRateChange)
	inflationRateChange := inflationRateChangePerYear.Quo(hrsPerYr)

	// increase the new annual inflation for this next cycle
	inflation = p.Inflation.Add(inflationRateChange)
	if inflation.GT(params.InflationMax) {
		inflation = params.InflationMax
	}
	if inflation.LT(params.InflationMin) {
		inflation = params.InflationMin
	}

	return
}
