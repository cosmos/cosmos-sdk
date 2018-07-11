package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

const (
	hrsPerYr  = 8766         // as defined by a julian year of 365.25 days
	precision = 100000000000 // increased to this precision for accuracy
)

var hrsPerYrRat = sdk.NewRat(hrsPerYr)

// process provisions for an hour period
func (k Keeper) ProcessProvisions(ctx sdk.Context) types.Pool {

	pool := k.GetPool(ctx)
	pool.Inflation = k.NextInflation(ctx)

	provisions := pool.Inflation.Mul(sdk.NewRat(pool.TokenSupply())).Quo(hrsPerYrRat).RoundInt64()

	// TODO add to the fees provisions
	pool.LooseTokens += provisions
	return pool
}

// get the next inflation rate for the hour
func (k Keeper) NextInflation(ctx sdk.Context) (inflation sdk.Rat) {

	params := k.GetParams(ctx)
	pool := k.GetPool(ctx)
	// The target annual inflation rate is recalculated for each previsions cycle. The
	// inflation is also subject to a rate change (positive or negative) depending on
	// the distance from the desired ratio (67%). The maximum rate change possible is
	// defined to be 13% per year, however the annual inflation is capped as between
	// 7% and 20%.

	// (1 - bondedRatio/GoalBonded) * InflationRateChange
	inflationRateChangePerYear := sdk.OneRat().Sub(pool.BondedRatio().Quo(params.GoalBonded)).Mul(params.InflationRateChange)
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
