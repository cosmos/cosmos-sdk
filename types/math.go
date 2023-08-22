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

	ToDec = sdkmath.LegacyNewDecFromInt
)

const (
	MaxBitLen = sdkmath.MaxBitLen
)

func (ip IntProto) String() string {
	return ip.Int.String()
}
