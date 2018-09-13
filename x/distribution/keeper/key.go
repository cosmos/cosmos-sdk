package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// keys/key-prefixes
var (
	FeePoolKey           = []byte{0x00} // key for global distribution state
	ValidatorDistInfoKey = []byte{0x01} // prefix for each key to a validator distribution
	DelegatorDistInfoKey = []byte{0x02} // prefix for each key to a delegation distribution

	// transient
	ProposerKey = []byte{0x00} // key for storing the proposer operator address
)

// gets the key for the validator distribution info from address
// VALUE: distribution/types.ValidatorDistInfo
func GetValidatorDistInfoKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorDistInfoKey, operatorAddr.Bytes()...)
}

// gets the key for delegator distribution for a validator
// VALUE: distribution/types.DelegatorDistInfo
func GetDelegationDistInfoKey(delAddr sdk.AccAddress, valOperatorAddr sdk.ValAddress) []byte {
	return append(GetDelegationDistInfosKey(delAddr), valAddr.Bytes()...)
}

// gets the prefix for a delegator's distributions across all validators
func GetDelegationDistInfosKey(delAddr sdk.AccAddress) []byte {
	return append(DelegatorDistInfoKey, delAddr.Bytes()...)
}
