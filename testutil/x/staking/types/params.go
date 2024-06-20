package types

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Staking params default values
const (
	// DefaultUnbondingTime reflects three weeks in seconds as the default
	// unbonding time.
	// TODO: Justify our choice of default here.
	DefaultUnbondingTime time.Duration = time.Hour * 24 * 7 * 3

	// Default maximum number of bonded validators
	DefaultMaxValidators uint32 = 100

	// Default maximum entries in a UBD/RED pair
	DefaultMaxEntries uint32 = 7

	// DefaultHistorical entries is 10000. Apps that don't use IBC can ignore this
	// value by not adding the staking module to the application module manager's
	// SetOrderBeginBlockers.
	DefaultHistoricalEntries uint32 = 10000
)

var (
	// DefaultMinCommissionRate is set to 0%
	DefaultMinCommissionRate = math.LegacyZeroDec()

	// DefaultKeyRotationFee is fees used to rotate the ConsPubkey or Operator key
	DefaultKeyRotationFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)
)

// NewParams creates a new Params instance
func NewParams(unbondingTime time.Duration,
	maxValidators, maxEntries, historicalEntries uint32,
	bondDenom string, minCommissionRate math.LegacyDec,
	keyRotationFee sdk.Coin,
) Params {
	return Params{
		UnbondingTime:     unbondingTime,
		MaxValidators:     maxValidators,
		MaxEntries:        maxEntries,
		HistoricalEntries: historicalEntries,
		BondDenom:         bondDenom,
		MinCommissionRate: minCommissionRate,
		KeyRotationFee:    keyRotationFee,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultUnbondingTime,
		DefaultMaxValidators,
		DefaultMaxEntries,
		DefaultHistoricalEntries,
		sdk.DefaultBondDenom,
		DefaultMinCommissionRate,
		DefaultKeyRotationFee,
	)
}
