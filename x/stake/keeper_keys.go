package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

//nolint
var (
	// Keys for store prefixes
	CandidatesAddrKey = []byte{0x01} // key for all candidates' addresses
	ParamKey          = []byte{0x02} // key for global parameters relating to staking
	PoolKey           = []byte{0x03} // key for global parameters relating to staking

	// Key prefixes
	CandidateKeyPrefix        = []byte{0x04} // prefix for each key to a candidate
	ValidatorKeyPrefix        = []byte{0x05} // prefix for each key to a candidate
	ValidatorUpdatesKeyPrefix = []byte{0x06} // prefix for each key to a candidate
	DelegatorBondKeyPrefix    = []byte{0x07} // prefix for each key to a delegator's bond
)

// XXX remove beggining word get from all these keys
// GetCandidateKey - get the key for the candidate with address
func GetCandidateKey(addr sdk.Address) []byte {
	return append(CandidateKeyPrefix, addr.Bytes()...)
}

// GetValidatorKey - get the key for the validator used in the power-store
func GetValidatorKey(addr sdk.Address, power sdk.Rat, cdc *wire.Codec) []byte {
	b, _ := cdc.MarshalBinary(power)                                 // TODO need to handle error here?
	return append(ValidatorKeyPrefix, append(b, addr.Bytes()...)...) // TODO does this need prefix if its in its own store
}

// GetValidatorUpdatesKey - get the key for the validator used in the power-store
func GetValidatorUpdatesKey(addr sdk.Address) []byte {
	return append(ValidatorUpdatesKeyPrefix, addr.Bytes()...) // TODO does this need prefix if its in its own store
}

// GetDelegatorBondKey - get the key for delegator bond with candidate
func GetDelegatorBondKey(delegatorAddr, candidateAddr sdk.Address, cdc *wire.Codec) []byte {
	return append(GetDelegatorBondsKey(delegatorAddr, cdc), candidateAddr.Bytes()...)
}

// GetDelegatorBondKeyPrefix - get the prefix for a delegator for all candidates
func GetDelegatorBondsKey(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res, err := cdc.MarshalBinary(&delegatorAddr)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondKeyPrefix, res...)
}
