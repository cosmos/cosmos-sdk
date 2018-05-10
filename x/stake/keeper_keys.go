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
	ParamKey                       = []byte{0x00} // key for global parameters relating to staking
	PoolKey                        = []byte{0x01} // key for global parameters relating to staking
	ValidatorsKey                  = []byte{0x02} // prefix for each key to a validator
	ValidatorsByPowerKey           = []byte{0x03} // prefix for each key to a validator
	ValidatorsTendermintUpdatesKey = []byte{0x04} // prefix for each key to a validator which is being updated
	ValidatorsBondedKey            = []byte{0x05} // prefix for each key to bonded/actively validating validators
	DelegationKey                  = []byte{0x06} // prefix for each key to a delegator's bond
	IntraTxCounterKey              = []byte{0x07} // key for block-local tx index
)

const maxDigitsForAccount = 12 // ~220,000,000 atoms created at launch

// get the key for the validator with address
func GetValidatorKey(addr sdk.Address) []byte {
	return append(ValidatorsKey, addr.Bytes()...)
}

// get the key for the validator used in the power-store
func GetValidatorsBondedByPowerKey(validator Validator) []byte {
	powerBytes := []byte(validator.Power.ToLeftPadded(maxDigitsForAccount)) // power big-endian (more powerful validators first)

	// TODO ensure that the key will be a readable string.. probably should add seperators and have
	// heightBytes and counterBytes represent strings like powerBytes does
	heightBytes := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(heightBytes, ^uint64(validator.Height)) // invert height (older validators first)
	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, ^uint16(validator.Counter)) // invert counter (first txns have priority)
	return append(ValidatorsByPowerKey,
		append(powerBytes,
			append(heightBytes,
				append(counterBytes, validator.Address.Bytes()...)...)...)...)
}

// get the key for the accumulated update validators
func GetValidatorsTendermintUpdatesKey(addr sdk.Address) []byte {
	return append(ValidatorsTendermintUpdatesKey, addr.Bytes()...)
}

// get the key for the current validator group, ordered like tendermint
func GetValidatorsBondedBondedKey(pk crypto.PubKey) []byte {
	addr := pk.Address()
	return append(ValidatorsBondedKey, addr.Bytes()...)
}

// get the key for delegator bond with validator
func GetDelegationKey(delegatorAddr, validatorAddr sdk.Address, cdc *wire.Codec) []byte {
	return append(GetDelegationsKey(delegatorAddr, cdc), validatorAddr.Bytes()...)
}

// get the prefix for a delegator for all validators
func GetDelegationsKey(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res, err := cdc.MarshalBinary(&delegatorAddr)
	if err != nil {
		panic(err)
	}
	return append(DelegationKey, res...)
}

//______________________________________________________________

// remove the prefix byte from a key, possibly revealing and address
func AddrFromKey(key []byte) sdk.Address {
	return key[1:]
}
