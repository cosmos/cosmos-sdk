package types

import (
	sdkmath "cosmossdk.io/math"
)

// Type aliases to the SDK's math sub-module
//
// Deprecated: Functionality of this package has been moved to it's own module:
// github.com/cosmos/cosmos-sdk/math
//
// Please use the above module instead of this package.
type (
	Int  = sdkmath.Int
	Uint = sdkmath.Uint
	Dec  = sdkmath.LegacyDec
)

const (
	Precision = sdkmath.LegacyPrecision
)

var (
	// Int
	NewIntFromBigInt = sdkmath.NewIntFromBigInt
	OneInt           = sdkmath.OneInt
	NewInt           = sdkmath.NewInt
	ZeroInt          = sdkmath.ZeroInt
	IntEq            = sdkmath.IntEq
	NewIntFromString = sdkmath.NewIntFromString
	NewUint          = sdkmath.NewUint
	NewIntFromUint64 = sdkmath.NewIntFromUint64
	MaxInt           = sdkmath.MaxInt
	MinInt           = sdkmath.MinInt

	// Dec
	ToDec = sdkmath.LegacyNewDecFromInt

	NewDec                   = sdkmath.LegacyNewDec
	NewDecWithPrec           = sdkmath.LegacyNewDecWithPrec
	NewDecFromBigInt         = sdkmath.LegacyNewDecFromBigInt
	NewDecFromBigIntWithPrec = sdkmath.LegacyNewDecFromBigIntWithPrec
	NewDecFromInt            = sdkmath.LegacyNewDecFromInt
	NewDecFromIntWithPrec    = sdkmath.LegacyNewDecFromIntWithPrec
	NewDecFromStr            = sdkmath.LegacyNewDecFromStr
	MustNewDecFromStr        = sdkmath.LegacyMustNewDecFromStr
	MaxSortableDec           = sdkmath.LegacyMaxSortableDec
	MaxDec                   = sdkmath.LegacyMaxDec
	DecEq                    = sdkmath.LegacyDecEq
	DecApproxEq              = sdkmath.LegacyDecApproxEq
	FormatDec                = sdkmath.FormatDec
	MinDec                   = sdkmath.LegacyMinDec
	ZeroDec                  = sdkmath.LegacyZeroDec
	OneDec                   = sdkmath.LegacyOneDec
	SmallestDec              = sdkmath.LegacySmallestDec
)

const (
	MaxBitLen = sdkmath.MaxBitLen
)

func (ip IntProto) String() string {
	return ip.Int.String()
}

func (dp DecProto) String() string {
	return dp.Dec.String()
}
