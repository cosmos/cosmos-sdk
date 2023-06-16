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
	Int = sdkmath.Int
)

var (
	NewInt  = sdkmath.NewInt
	ZeroInt = sdkmath.ZeroInt
)

func (ip IntProto) String() string {
	return ip.Int.String()
}

type (
	Dec = sdkmath.LegacyDec
)

var (
	NewDecWithPrec    = sdkmath.LegacyNewDecWithPrec
	NewDecFromInt     = sdkmath.LegacyNewDecFromInt
	NewDecFromStr     = sdkmath.LegacyNewDecFromStr
	MustNewDecFromStr = sdkmath.LegacyMustNewDecFromStr
)

var _ CustomProtobufType = (*Dec)(nil)

func (dp DecProto) String() string {
	return dp.Dec.String()
}
