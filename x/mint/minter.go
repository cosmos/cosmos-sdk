package mint

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Minter represents the minting state
type Minter struct {
	LastUpdate       time.Time `json:"last_update"`       // time which the last update was made to the minter
	Inflation        sdk.Dec   `json:"inflation"`         // current annual inflation rate
	AnnualProvisions sdk.Dec   `json:"annual_provisions"` // current annual expected provisions
}

// Create a new minter object
func NewMinter(lastUpdate time.Time, inflation,
	annualProvisions sdk.Dec) Minter {

	return Minter{
		LastUpdate:       lastUpdate,
		Inflation:        inflation,
		AnnualProvisions: annualProvisions,
	}
}

// minter object for a new chain
func InitialMinter(inflation sdk.Dec) Minter {
	return NewMinter(
		time.Unix(0, 0),
		inflation,
		sdk.NewDec(0),
	)
}

// default initial minter object for a new chain
// which uses an inflation rate of 13%
func DefaultInitialMinter() Minter {
	return InitialMinter(
		sdk.NewDecWithPrec(13, 2),
	)
}

func validateMinter(minter Minter) error {
	if minter.Inflation.LT(sdk.ZeroDec()) {
		return fmt.Errorf("mint parameter Inflation should be positive, is %s",
			minter.Inflation.String())
	}
	return nil
}

var hrsPerYr = sdk.NewDec(8766) // as defined by a julian year of 365.25 days

// get the new inflation rate for the next hour
func (m Minter) NextInflationRate(params Params, bondedRatio sdk.Dec) (
	inflation sdk.Dec) {

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

// calculate the annual provisions based on current total supply and inflation rate
func (m Minter) NextAnnualProvisions(params Params, totalSupply sdk.Dec) (
	provisions sdk.Dec) {

	return m.Inflation.Mul(totalSupply)
}

// get the provisions for a block based on the annual provisions rate
func (m Minter) BlockProvision(params Params) sdk.Coin {
	provisionAmt := m.AnnualProvisions.QuoInt(sdk.NewInt(int64(params.BlocksPerYear)))
	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}
