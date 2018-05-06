package stake

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	crypto "github.com/tendermint/go-crypto"
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
	IntraTxCounterKey      = []byte{0x08} // key for block-local tx index
	PowerChangeKey         = []byte{0x09} // prefix for power change object
)

const maxDigitsForAccount = 12 // ~220,000,000 atoms created at launch

// get the key for the candidate with address
func GetCandidateKey(addr sdk.Address) []byte {
	return append(CandidatesKey, addr.Bytes()...)
}

// get the key for the validator used in the power-store
func GetValidatorKey(validator Validator) []byte {
	powerBytes := []byte(validator.Power.ToLeftPadded(maxDigitsForAccount)) // power big-endian (more powerful validators first)

	// TODO ensure that the key will be a readable string.. probably should add seperators and have
	// heightBytes and counterBytes represent strings like powerBytes does
	heightBytes := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(heightBytes, ^uint64(validator.Height)) // invert height (older validators first)
	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, ^uint16(validator.Counter)) // invert counter (first txns have priority)
	return append(ValidatorsKey,
		append(powerBytes,
			append(heightBytes,
				append(counterBytes, validator.Address.Bytes()...)...)...)...)
}

// get the key for the accumulated update validators
func GetAccUpdateValidatorKey(addr sdk.Address) []byte {
	return append(AccUpdateValidatorsKey, addr.Bytes()...)
}

// get the key for the recent validator group, ordered like tendermint
func GetRecentValidatorKey(pk crypto.PubKey) []byte {
	addr := pk.Address()
	return append(RecentValidatorsKey, addr.Bytes()...)
}

// remove the prefix byte from a key
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

// get the key for the accumulated update validators
func GetPowerChangeKey(height int64) []byte {
	heightBytes := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(heightBytes, ^uint64(height)) // invert height (older validators first)
	return append(PowerChangeKey, heightBytes...)
}
