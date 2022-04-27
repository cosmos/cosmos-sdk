package types

import (
	sdkmath "github.com/cosmos/cosmos-sdk/math"
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

func (ip IntProto) String() string {
	return ip.Int.String()
}

// ToDec converts an Int type to a Dec type.
func ToDec(i Int) Dec {
	return NewDecFromInt(i)
}
