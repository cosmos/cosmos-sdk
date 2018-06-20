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
	ParamKey                   = []byte{0x00} // key for parameters relating to staking
	PoolKey                    = []byte{0x01} // key for the staking pools
	ValidatorsKey              = []byte{0x02} // prefix for each key to a validator
	ValidatorsByPubKeyIndexKey = []byte{0x03} // prefix for each key to a validator by pubkey
	ValidatorsBondedKey        = []byte{0x04} // prefix for each key to bonded/actively validating validators
	ValidatorsByPowerKey       = []byte{0x05} // prefix for each key to a validator sorted by power
	ValidatorCliffKey          = []byte{0x06} // key for block-local tx index
	ValidatorPowerCliffKey     = []byte{0x07} // key for block-local tx index
	TendermintUpdatesKey       = []byte{0x08} // prefix for each key to a validator which is being updated
	DelegationKey              = []byte{0x09} // prefix for each key to a delegator's bond
	IntraTxCounterKey          = []byte{0x10} // key for block-local tx index
)

const maxDigitsForAccount = 12 // ~220,000,000 atoms created at launch

// get the key for the validator with address
func GetValidatorKey(ownerAddr sdk.Address) []byte {
	return append(ValidatorsKey, ownerAddr.Bytes()...)
}

// get the key for the validator with pubkey
func GetValidatorByPubKeyIndexKey(pubkey crypto.PubKey) []byte {
	return append(ValidatorsByPubKeyIndexKey, pubkey.Bytes()...)
}

// get the key for the current validator group, ordered like tendermint
func GetValidatorsBondedKey(pk crypto.PubKey) []byte {
	addr := pk.Address()
	return append(ValidatorsBondedKey, addr.Bytes()...)
}

// get the key for the validator used in the power-store
func GetValidatorsByPowerKey(validator Validator, pool Pool) []byte {

	power := validator.EquivalentBondedShares(pool)
	powerBytes := []byte(power.ToLeftPadded(maxDigitsForAccount)) // power big-endian (more powerful validators first)

	// TODO ensure that the key will be a readable string.. probably should add seperators and have
	revokedBytes := make([]byte, 1)
	if validator.Revoked {
		revokedBytes[0] = byte(0x01)
	} else {
		revokedBytes[0] = byte(0x00)
	}
	// heightBytes and counterBytes represent strings like powerBytes does
	heightBytes := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(heightBytes, ^uint64(validator.BondHeight)) // invert height (older validators first)
	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, ^uint16(validator.BondIntraTxCounter)) // invert counter (first txns have priority)
	return append(ValidatorsByPowerKey,
		append(revokedBytes,
			append(powerBytes,
				append(heightBytes,
					append(counterBytes, validator.Owner.Bytes()...)...)...)...)...) // TODO don't technically need to store owner
}

// get the key for the accumulated update validators
func GetTendermintUpdatesKey(ownerAddr sdk.Address) []byte {
	return append(TendermintUpdatesKey, ownerAddr.Bytes()...)
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
