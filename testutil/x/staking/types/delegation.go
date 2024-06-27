package types

import (
	"cosmossdk.io/math"
)

// NewDelegation creates a new delegation object
func NewDelegation(delegatorAddr, validatorAddr string, shares math.LegacyDec) Delegation {
	return Delegation{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Shares:           shares,
	}
}
