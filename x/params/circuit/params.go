package circuit

import (
	params "github.com/cosmos/cosmos-sdk/x/params/space"
)

// Default parameter namespace
const (
	DefaultParamSpace = "circuit"
)

// CircuitBrakeKey - returns key for boolean flag indicating circuit brake
func CircuitBrakeKey(msgtype string) params.Key { return params.NewKey(msgtype) }
