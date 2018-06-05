package keeper

import (
	"encoding/binary"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// TODO remove some of these prefixes once have working multistore

//nolint
var (
	// Keys for store prefixes
	ParamKey                         = []byte{0x00} // key for parameters relating to staking
	PoolKey                          = []byte{0x01} // key for the staking pools
	ValidatorsKey                    = []byte{0x02} // prefix for each key to a validator
	ValidatorsByPubKeyIndexKey       = []byte{0x03} // prefix for each key to a validator index, by pubkey
	ValidatorsBondedIndexKey         = []byte{0x04} // prefix for each key to a validator index, for bonded validators
	ValidatorsByPowerIndexKey        = []byte{0x05} // prefix for each key to a validator index, sorted by power
	ValidatorCliffIndexKey           = []byte{0x06} // key for the validator index of the cliff validator
	ValidatorPowerCliffKey           = []byte{0x07} // key for the power of the validator on the cliff
	TendermintUpdatesKey             = []byte{0x08} // prefix for each key to a validator which is being updated
	DelegationKey                    = []byte{0x09} // key for a delegation
	UnbondingDelegationKey           = []byte{0x0A} // key for an unbonding-delegation
	UnbondingDelegationByValIndexKey = []byte{0x0B} // prefix for each key for an unbonding-delegation, by validator owner
	RedelegationKey                  = []byte{0x0C} // key for a redelegation
	RedelegationByValIndexKey        = []byte{0x0D} // prefix for each key for an redelegation, by validator owner
	IntraTxCounterKey                = []byte{0x0E} // key for intra-block tx index
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
func GetValidatorsBondedIndexKey(pk crypto.PubKey) []byte {
	addr := pk.Address()
	return append(ValidatorsBondedIndexKey, addr.Bytes()...)
}

// get the power which is the key for the validator used in the power-store
func GetValidatorsByPowerIndexKey(validator types.Validator, pool types.Pool) []byte {
	return GetValidatorsByPower(validator, pool)
}

// get the power of a validator
func GetValidatorsByPower(validator types.Validator, pool types.Pool) []byte {

	power := validator.EquivalentBondedShares(pool)
	powerBytes := []byte(power.ToLeftPadded(maxDigitsForAccount)) // power big-endian (more powerful validators first)

	// TODO ensure that the key will be a readable string.. probably should add seperators and have
	// heightBytes and counterBytes represent strings like powerBytes does
	heightBytes := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(heightBytes, ^uint64(validator.BondHeight)) // invert height (older validators first)
	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, ^uint16(validator.BondIntraTxCounter)) // invert counter (first txns have priority)

	// NOTE the address doesn't need to be stored because counter bytes must always be different
	return append(ValidatorsByPowerIndexKey,
		append(powerBytes,
			append(heightBytes, counterBytes...)...)...)
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
