package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter(blockHeader tmproto.Header, inflation, annualProvisions sdk.Dec) Minter {
	return Minter{
		BlockHeader:      blockHeader,
		Inflation:        inflation,
		AnnualProvisions: annualProvisions,
	}
}

// InitialMinter returns an initial Minter object with a given inflation value.
func InitialMinter(blockHeader tmproto.Header, inflation sdk.Dec) Minter {
	return NewMinter(
		blockHeader,
		inflation,
		sdk.NewDec(0),
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain
// which uses an inflation rate of 13%.
func DefaultInitialMinter() Minter {
	return InitialMinter(
		tmproto.Header{},
		sdk.NewDecWithPrec(13, 2),
	)
}

// validate minter
func ValidateMinter(minter Minter) error {
	if minter.Inflation.IsNegative() {
		return fmt.Errorf("mint parameter Inflation should be positive, is %s",
			minter.Inflation.String())
	}
	return nil
}

// NextInflationRate returns the new inflation rate for the next hour.
func (m Minter) NextInflationRate(params Params, bondedRatio sdk.Dec) sdk.Dec {
	// The target annual inflation rate is recalculated for each previsions cycle. The
	// inflation is also subject to a rate change (positive or negative) depending on
	// the distance from the desired ratio (67%). The maximum rate change possible is
	// defined to be 13% per year, however the annual inflation is capped as between
	// 7% and 20%.

	// (1 - bondedRatio/GoalBonded) * InflationRateChange
	inflationRateChangePerYear := sdk.OneDec().
		Sub(bondedRatio.Quo(params.GoalBonded)).
		Mul(params.InflationRateChange)
	inflationRateChange := inflationRateChangePerYear.Quo(sdk.NewDec(int64(params.BlocksPerYear)))

	// adjust the new annual inflation for this next cycle
	inflation := m.Inflation.Add(inflationRateChange) // note inflationRateChange may be negative
	if inflation.GT(params.InflationMax) {
		inflation = params.InflationMax
	}
	if inflation.LT(params.InflationMin) {
		inflation = params.InflationMin
	}

	return inflation
}

// NextAnnualProvisions returns the annual provisions based on current total
// supply and inflation rate.
func (m Minter) NextAnnualProvisions(_ Params, totalSupply sdk.Int) sdk.Dec {
	return m.Inflation.MulInt(totalSupply)
}

// BlockProvision returns the provisions for a block based on the annual
// provisions rate.
func (m Minter) BlockProvision(params Params) sdk.Coin {
	blocksPerYear := sdk.NewDec(int64(params.BlocksPerYear))
	currentHeight := sdk.NewDec(m.BlockHeader.GetHeight())
	//currentHeight := sdk.NewDec(6311520 * 2)

	currentYear := currentHeight.Quo(blocksPerYear).TruncateDec()
	mintedAmountPerBlock := params.MintedAmountPerBlock

	for i := 0; i < int(currentYear.RoundInt64()); i++ {
		reductionAmount := mintedAmountPerBlock.Mul(params.YearlyReduction)
		mintedAmountPerBlock = mintedAmountPerBlock.Sub(reductionAmount)
	}
	provisionAmt := mintedAmountPerBlock
	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}

// BlockPeriodProvision returns the provisions for a block based on the period
func (m Minter) BlockPeriodProvision(params Params) sdk.Coin {
	fmt.Println("BlockPeriodProvision ", params.BlocksPerYear)
	fmt.Println("params.BlocksPerYear: ", params.BlocksPerYear)
	blocksPerPeriod := sdk.NewDec(int64(params.BlocksPerYear))
	currentHeight := sdk.NewDec(m.BlockHeader.GetHeight())
	//currentHeight := sdk.NewDec(6311520 * 2)

	fmt.Println("currentHeight: ", currentHeight)

	currentPeriod := currentHeight.Quo(blocksPerPeriod).TruncateDec()
	fmt.Println("currentPeriod: ", currentPeriod)
	fmt.Println("currentYearlyReduction: ", params.YearlyReduction)

	mintedAmountPerBlock := params.MintedAmountPerBlock

	for i := 0; i < int(currentPeriod.RoundInt64()); i++ {
		reductionAmount := mintedAmountPerBlock.Mul(params.YearlyReduction)
		mintedAmountPerBlock = mintedAmountPerBlock.Sub(reductionAmount)
	}
	provisionAmt := mintedAmountPerBlock
	fmt.Println("BlockPeriodProvision: ", sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt()))

	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}
