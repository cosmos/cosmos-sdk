package mint

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Minter - dynamic parameters of the current state
type Params struct {
	MintDenom           string  `json:"mint_denom"`            // type of coin to mint
	InflationRateChange sdk.Dec `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Dec `json:"inflation_max"`         // maximum inflation rate
	InflationMin        sdk.Dec `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          sdk.Dec `json:"goal_bonded"`           // goal of percent bonded atoms
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:           "steak",
		InflationRateChange: sdk.NewDecWithPrec(13, 2),
		InflationMax:        sdk.NewDecWithPrec(20, 2),
		InflationMin:        sdk.NewDecWithPrec(7, 2),
		GoalBonded:          sdk.NewDecWithPrec(67, 2),
	}
}

func validateParams(params Params) error {
	if params.GoalBonded.LT(sdk.ZeroDec()) {
		return fmt.Errorf("mint parameter GoalBonded should be positive, is %s ", params.GoalBonded.String())
	}
	if params.GoalBonded.GT(sdk.OneDec()) {
		return fmt.Errorf("mint parameter GoalBonded must be <= 1, is %s", params.GoalBonded.String())
	}
	if params.InflationMax.LT(params.InflationMin) {
		return fmt.Errorf("mint parameter Max inflation must be greater than or equal to min inflation")
	}
	if params.MintDenom == "" {
		return fmt.Errorf("mint parameter MintDenom can't be an empty string")
	}
	return nil
}

//______________________________________________________________

// Minter - dynamic parameters of the current state
type Minter struct {
	InflationLastTime time.Time `json:"inflation_last_time"` // block time which the last inflation was processed
	Inflation         sdk.Dec   `json:"inflation"`           // current annual inflation rate
}

// minter object for a new minter
func InitialMinter() Minter {
	return Minter{
		InflationLastTime: time.Unix(0, 0),
		Inflation:         sdk.NewDecWithPrec(13, 2),
	}
}

func validateMinter(minter Minter) error {
	if minter.Inflation.LT(sdk.ZeroDec()) {
		return fmt.Errorf("mint parameter Inflation should be positive, is %s ", minter.Inflation.String())
	}
	if minter.Inflation.GT(sdk.OneDec()) {
		return fmt.Errorf("mint parameter Inflation must be <= 1, is %s", minter.Inflation.String())
	}
	return nil
}

var hrsPerYr = sdk.NewDec(8766) // as defined by a julian year of 365.25 days

// process provisions for an hour period
func (m Minter) ProcessProvisions(params Params, totalSupply, bondedRatio sdk.Dec) (
	minter Minter, provisions sdk.Coin) {

	m.Inflation = m.NextInflation(params, bondedRatio)
	provisionsDec := m.Inflation.Mul(totalSupply).Quo(hrsPerYr)
	provisions = sdk.NewCoin(params.MintDenom, provisionsDec.TruncateInt())
	return m, provisions
}

// get the next inflation rate for the hour
func (m Minter) NextInflation(params Params, bondedRatio sdk.Dec) (inflation sdk.Dec) {

	// The target annual inflation rate is recalculated for each previsions cycle. The
	// inflation is also subject to a rate change (positive or negative) depending on
	// the distance from the desired ratio (67%). The maximum rate change possible is
	// defined to be 13% per year, however the annual inflation is capped as between
	// 7% and 20%.

	// (1 - bondedRatio/GoalBonded) * InflationRateChange
	inflationRateChangePerYear := sdk.OneDec().
		Sub(bondedRatio.Quo(params.GoalBonded)).
		Mul(params.InflationRateChange)
	inflationRateChange := inflationRateChangePerYear.Quo(hrsPerYr)

	// increase the new annual inflation for this next cycle
	inflation = m.Inflation.Add(inflationRateChange)
	if inflation.GT(params.InflationMax) {
		inflation = params.InflationMax
	}
	if inflation.LT(params.InflationMin) {
		inflation = params.InflationMin
	}

	return inflation
}
