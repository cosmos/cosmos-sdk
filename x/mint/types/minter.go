package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values. Previous block time is initially nil and will be set during the execution.
func NewMinter(
	inflation math.LegacyDec,
	annualProvisions math.LegacyDec,
	genesisTime *time.Time,
	mintDenom string,
) Minter {
	return Minter{
		Inflation:         inflation,
		AnnualProvisions:  annualProvisions,
		GenesisTime:       genesisTime,
		PreviousBlockTime: nil, // is nil here
		MintDenom:         mintDenom,
	}
}

// InitialMinter returns an initial Minter object with a given inflation value.
func InitialMinter(inflation math.LegacyDec) Minter {
	genesisTime := time.Unix(0, 0).UTC()
	return NewMinter(
		inflation,
		math.LegacyZeroDec(),
		&genesisTime,
		sdk.DefaultBondDenom,
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain
// which uses an inflation rate of 13%.
func DefaultInitialMinter() Minter {
	return InitialMinter(
		initialInflationRate,
	)
}

// ValidateMinter does a basic validation on minter.
func ValidateMinter(minter Minter) error {
	if minter.Inflation.IsNegative() {
		return fmt.Errorf("inflation %v should be positive", minter.Inflation.String())
	}
	if minter.AnnualProvisions.IsNegative() {
		return fmt.Errorf("annual provisions %v should be positive", minter.AnnualProvisions.String())
	}
	return nil
}

// AnnualProvisions returns the annual provisions based on the total supply and the inflation rate.
func AnnualProvisions(inflation math.LegacyDec, totalSupply math.Int) math.LegacyDec {
	return inflation.MulInt(totalSupply)
}

// BlockProvision returns the provisions for a block based on the annual provisions rate.
func BlockProvision(
	currentBlockTime time.Time,
	previousBlockTime time.Time,
	annualProvisions math.LegacyDec,
) math.Int {
	// nanosecs in year / diff == blocks - the number of blocks if the rate keeps the same
	// provision / blocks == provision per block
	// -> provision per block = provision * diff / year

	diff := currentBlockTime.Sub(previousBlockTime).Nanoseconds() // nanosecs between the blocks
	provisionAmt := annualProvisions.Mul(math.LegacyNewDec(diff)).Quo(nanosecondsPerYearDec)

	return provisionAmt.TruncateInt()
}

// InflationRate returns the new inflation rate for the given block.
// The algorithm is implemented according to the ADR 019 of Celestia.
func InflationRate(genesisTime, blockTime time.Time) math.LegacyDec {
	years := yearsSinceGenesis(genesisTime, blockTime)
	inflationRate := initialInflationRate.Mul(math.LegacyOneDec().Sub(disinflationRate).Power(uint64(years)))

	if inflationRate.LT(targetInflationRate) {
		return targetInflationRate
	}
	return inflationRate
}

// yearsSinceGenesis returns the number of years that have passed between
// genesis and current (rounded down).
func yearsSinceGenesis(genesis time.Time, current time.Time) (years int64) {
	if current.Before(genesis) {
		return 0
	}
	return current.Sub(genesis).Nanoseconds() / nanosecondsPerYear
}
