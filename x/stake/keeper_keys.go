package stake

import (
	"encoding/binary"

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
	RecentValidatorsKey    = []byte{0x05} // prefix for each key to the last updated validator group

	ToKickOutValidatorsKey = []byte{0x06} // prefix for each key to the last updated validator group

	DelegatorBondKeyPrefix = []byte{0x07} // prefix for each key to a delegator's bond

	CounterKey = []byte{0x08} // key for block-local tx index
)

const maxDigitsForAccount = 12 // ~220,000,000 atoms created at launch

// get the key for the candidate with address
func GetCandidateKey(addr sdk.Address) []byte {
	return append(CandidatesKey, addr.Bytes()...)
}

// get the key for the validator used in the power-store
func GetValidatorKey(addr sdk.Address, power sdk.Rat, height int64, counter int16, cdc *wire.Codec) []byte {
	powerBytes := []byte(power.ToLeftPadded(maxDigitsForAccount)) // power big-endian (more powerful validators first)
	heightBytes := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(heightBytes, ^uint64(height)) // invert height (older validators first)
	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, ^uint16(counter)) // invert counter (first txns have priority)
	return append(ValidatorsKey, append(powerBytes, append(heightBytes, append(counterBytes, addr.Bytes()...)...)...)...)
}

// get the key for the accumulated update validators
func GetAccUpdateValidatorKey(addr sdk.Address) []byte {
	return append(AccUpdateValidatorsKey, addr.Bytes()...)
}

// get the key for the accumulated update validators
func GetRecentValidatorKey(addr sdk.Address) []byte {
	return append(RecentValidatorsKey, addr.Bytes()...)
}

// reverse operation of GetRecentValidatorKey
func AddrFromKey(key []byte) sdk.Address {
	return key[1:]
}

// get the key for the accumulated update validators
func GetToKickOutValidatorKey(addr sdk.Address) []byte {
	return append(ToKickOutValidatorsKey, addr.Bytes()...)
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
