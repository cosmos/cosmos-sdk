package types

import (
	sdkmath "cosmossdk.io/math"
)

// Type aliases to the SDK's math sub-module
//
// Deprecated: Functionality of this package has been moved to it's own module:
// cosmossdk.io/math
//
// Please use the above module instead of this package.
type (
	Int  = sdkmath.Int
	Uint = sdkmath.Uint
)

var (
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
)

const (
	MaxBitLen = sdkmath.MaxBitLen
)

func (ip IntProto) String() string {
	return ip.Int.String()
}

var _ CustomProtobufType = (*Dec)(nil)

type (
	Dec = sdkmath.Dec
)

const (
	Precision            = sdkmath.Precision
	DecimalPrecisionBits = sdkmath.DecimalPrecisionBits
)

var (
	ZeroDec                  = sdkmath.ZeroDec
	OneDec                   = sdkmath.OneDec
	SmallestDec              = sdkmath.SmallestDec
	NewDec                   = sdkmath.NewDec
	NewDecWithPrec           = sdkmath.NewDecWithPrec
	NewDecFromBigInt         = sdkmath.NewDecFromBigInt
	NewDecFromBigIntWithPrec = sdkmath.NewDecFromBigIntWithPrec
	NewDecFromInt            = sdkmath.NewDecFromInt
	NewDecFromIntWithPrec    = sdkmath.NewDecFromIntWithPrec
	NewDecFromStr            = sdkmath.NewDecFromStr
	MustNewDecFromStr        = sdkmath.MustNewDecFromStr
	MaxSortableDec           = sdkmath.MaxSortableDec
	ValidSortableDec         = sdkmath.ValidSortableDec
	SortableDecBytes         = sdkmath.SortableDecBytes
	DecsEqual                = sdkmath.DecsEqual
	MinDec                   = sdkmath.MinDec
	MaxDec                   = sdkmath.MaxDec
	DecEq                    = sdkmath.DecEq
	DecApproxEq              = sdkmath.DecApproxEq
)
