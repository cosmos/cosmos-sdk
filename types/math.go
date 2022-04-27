package types

import (
	mathmod "github.com/cosmos/cosmos-sdk/math"
)

// Type aliases to the SDK's math sub-module
//
// Deprecated: Functionality of this package has been moved to it's own module:
// github.com/cosmos/cosmos-sdk/math
//
// Please use the above module instead of this package.
var (
	Int = mathmod.Int
)

func (ip IntProto) String() string {
	return ip.Int.String()
}
