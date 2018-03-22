package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// TODO remove some of these prefixes once have working multistore

//nolint
var (
	// Keys for store prefixes
	ParamKey               = []byte{0x00} // key for global parameters relating to staking
	PoolKey                = []byte{0x01} // key for global parameters relating to staking
	CandidatesKey          = []byte{0x02} // prefix for each key to a candidate
	ValidatorsKey          = []byte{0x03} // prefix for each key to a validator
	AccUpdateValidatorsKey = []byte{0x04} // prefix for each key to a validator which is being updated
	RecentValidatorsKey    = []byte{0x04} // prefix for each key to the last updated validator group

	DelegatorBondKeyPrefix = []byte{0x05} // prefix for each key to a delegator's bond
)

// get the key for the candidate with address
func GetCandidateKey(addr sdk.Address) []byte {
	return append(CandidatesKey, addr.Bytes()...)
}

// get the key for the validator used in the power-store
func GetValidatorKey(addr sdk.Address, power sdk.Rat, cdc *wire.Codec) []byte {
	b, err := cdc.MarshalBinary(power)
	if err != nil {
		panic(err)
	}
	return append(ValidatorsKey, append(b, addr.Bytes()...)...)
}

// get the key for the accumulated update validators
func GetAccUpdateValidatorKey(addr sdk.Address) []byte {
	return append(AccUpdateValidatorsKey, addr.Bytes()...)
}

// get the key for the accumulated update validators
func GetRecentValidatorKey(addr sdk.Address) []byte {
	return append(RecentValidatorsKey, addr.Bytes()...)
}

// get the key for delegator bond with candidate
func GetDelegatorBondKey(delegatorAddr, candidateAddr sdk.Address, cdc *wire.Codec) []byte {
	return append(GetDelegatorBondsKey(delegatorAddr, cdc), candidateAddr.Bytes()...)
}

// get the prefix for a delegator for all candidates
func GetDelegatorBondsKey(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res, err := cdc.MarshalBinary(&delegatorAddr)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondKeyPrefix, res...)
}
