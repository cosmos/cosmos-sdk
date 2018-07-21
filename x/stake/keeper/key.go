package keeper

import (
	"encoding/binary"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	IntraTxCounterKey                = []byte{0x09} // key for intra-block tx index
	DelegationKey                    = []byte{0x0A} // key for a delegation
	UnbondingDelegationKey           = []byte{0x0B} // key for an unbonding-delegation
	UnbondingDelegationByValIndexKey = []byte{0x0C} // prefix for each key for an unbonding-delegation, by validator owner
	RedelegationKey                  = []byte{0x0D} // key for a redelegation
	RedelegationByValSrcIndexKey     = []byte{0x0E} // prefix for each key for an redelegation, by source validator owner
	RedelegationByValDstIndexKey     = []byte{0x0F} // prefix for each key for an redelegation, by destination validator owner
)

const maxDigitsForAccount = 12 // ~220,000,000 atoms created at launch

// get the key for the validator with address.
// VALUE: stake/types.Validator
func GetValidatorKey(ownerAddr sdk.AccAddress) []byte {
	return append(ValidatorsKey, ownerAddr.Bytes()...)
}

// get the key for the validator with pubkey.
// VALUE: validator owner address ([]byte)
func GetValidatorByPubKeyIndexKey(pubkey crypto.PubKey) []byte {
	return append(ValidatorsByPubKeyIndexKey, pubkey.Bytes()...)
}

// get the key for the current validator group
// VALUE: none (key rearrangement with GetValKeyFromValBondedIndexKey)
func GetValidatorsBondedIndexKey(ownerAddr sdk.AccAddress) []byte {
	return append(ValidatorsBondedIndexKey, ownerAddr.Bytes()...)
}

// Get the validator owner address from ValBondedIndexKey
func GetAddressFromValBondedIndexKey(IndexKey []byte) []byte {
	return IndexKey[1:] // remove prefix bytes
}

// get the validator by power index. power index is the key used in the power-store,
// and represents the relative power ranking of the validator.
// VALUE: validator owner address ([]byte)
func GetValidatorsByPowerIndexKey(validator types.Validator, pool types.Pool) []byte {
	// NOTE the address doesn't need to be stored because counter bytes must always be different
	return getValidatorPowerRank(validator, pool)
}

// get the power ranking of a validator
// NOTE the larger values are of higher value
func getValidatorPowerRank(validator types.Validator, pool types.Pool) []byte {

	potentialPower := validator.Tokens
	powerBytes := []byte(potentialPower.ToLeftPadded(maxDigitsForAccount)) // power big-endian (more powerful validators first)

	revokedBytes := make([]byte, 1)
	if validator.Revoked {
		revokedBytes[0] = byte(0x00)
	} else {
		revokedBytes[0] = byte(0x01)
	}

	// heightBytes and counterBytes represent strings like powerBytes does
	heightBytes := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(heightBytes, ^uint64(validator.BondHeight)) // invert height (older validators first)
	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, ^uint16(validator.BondIntraTxCounter)) // invert counter (first txns have priority)

	return append(append(append(append(
		ValidatorsByPowerIndexKey,
		revokedBytes...),
		powerBytes...),
		heightBytes...),
		counterBytes...)
}

// get the key for the accumulated update validators.
// VALUE: abci.Validator
// note records using these keys should never persist between blocks
func GetTendermintUpdatesKey(ownerAddr sdk.AccAddress) []byte {
	return append(TendermintUpdatesKey, ownerAddr.Bytes()...)
}

//________________________________________________________________________________

// get the key for delegator bond with validator.
// VALUE: stake/types.Delegation
func GetDelegationKey(delegatorAddr, validatorAddr sdk.AccAddress) []byte {
	return append(GetDelegationsKey(delegatorAddr), validatorAddr.Bytes()...)
}

// get the prefix for a delegator for all validators
func GetDelegationsKey(delegatorAddr sdk.AccAddress) []byte {
	return append(DelegationKey, delegatorAddr.Bytes()...)
}

//________________________________________________________________________________

// get the key for an unbonding delegation by delegator and validator addr.
// VALUE: stake/types.UnbondingDelegation
func GetUBDKey(delegatorAddr, validatorAddr sdk.AccAddress) []byte {
	return append(
		GetUBDsKey(delegatorAddr.Bytes()),
		validatorAddr.Bytes()...)
}

// get the index-key for an unbonding delegation, stored by validator-index
// VALUE: none (key rearrangement used)
func GetUBDByValIndexKey(delegatorAddr, validatorAddr sdk.AccAddress) []byte {
	return append(GetUBDsByValIndexKey(validatorAddr), delegatorAddr.Bytes()...)
}

// rearrange the ValIndexKey to get the UBDKey
func GetUBDKeyFromValIndexKey(IndexKey []byte) []byte {
	addrs := IndexKey[1:] // remove prefix bytes
	if len(addrs) != 2*sdk.AddrLen {
		panic("unexpected key length")
	}
	valAddr := addrs[:sdk.AddrLen]
	delAddr := addrs[sdk.AddrLen:]
	return GetUBDKey(delAddr, valAddr)
}

//______________

// get the prefix for all unbonding delegations from a delegator
func GetUBDsKey(delegatorAddr sdk.AccAddress) []byte {
	return append(UnbondingDelegationKey, delegatorAddr.Bytes()...)
}

// get the prefix keyspace for the indexes of unbonding delegations for a validator
func GetUBDsByValIndexKey(validatorAddr sdk.AccAddress) []byte {
	return append(UnbondingDelegationByValIndexKey, validatorAddr.Bytes()...)
}

//________________________________________________________________________________

// get the key for a redelegation
// VALUE: stake/types.RedelegationKey
func GetREDKey(delegatorAddr, validatorSrcAddr, validatorDstAddr sdk.AccAddress) []byte {
	return append(append(
		GetREDsKey(delegatorAddr.Bytes()),
		validatorSrcAddr.Bytes()...),
		validatorDstAddr.Bytes()...)
}

// get the index-key for a redelegation, stored by source-validator-index
// VALUE: none (key rearrangement used)
func GetREDByValSrcIndexKey(delegatorAddr, validatorSrcAddr, validatorDstAddr sdk.AccAddress) []byte {
	return append(append(
		GetREDsFromValSrcIndexKey(validatorSrcAddr),
		delegatorAddr.Bytes()...),
		validatorDstAddr.Bytes()...)
}

// get the index-key for a redelegation, stored by destination-validator-index
// VALUE: none (key rearrangement used)
func GetREDByValDstIndexKey(delegatorAddr, validatorSrcAddr, validatorDstAddr sdk.AccAddress) []byte {
	return append(append(
		GetREDsToValDstIndexKey(validatorDstAddr),
		delegatorAddr.Bytes()...),
		validatorSrcAddr.Bytes()...)
}

// rearrange the ValSrcIndexKey to get the REDKey
func GetREDKeyFromValSrcIndexKey(IndexKey []byte) []byte {
	addrs := IndexKey[1:] // remove prefix bytes
	if len(addrs) != 3*sdk.AddrLen {
		panic("unexpected key length")
	}
	valSrcAddr := addrs[:sdk.AddrLen]
	delAddr := addrs[sdk.AddrLen : 2*sdk.AddrLen]
	valDstAddr := addrs[2*sdk.AddrLen:]

	return GetREDKey(delAddr, valSrcAddr, valDstAddr)
}

// rearrange the ValDstIndexKey to get the REDKey
func GetREDKeyFromValDstIndexKey(IndexKey []byte) []byte {
	addrs := IndexKey[1:] // remove prefix bytes
	if len(addrs) != 3*sdk.AddrLen {
		panic("unexpected key length")
	}
	valDstAddr := addrs[:sdk.AddrLen]
	delAddr := addrs[sdk.AddrLen : 2*sdk.AddrLen]
	valSrcAddr := addrs[2*sdk.AddrLen:]
	return GetREDKey(delAddr, valSrcAddr, valDstAddr)
}

//______________

// get the prefix keyspace for redelegations from a delegator
func GetREDsKey(delegatorAddr sdk.AccAddress) []byte {
	return append(RedelegationKey, delegatorAddr.Bytes()...)
}

// get the prefix keyspace for all redelegations redelegating away from a source validator
func GetREDsFromValSrcIndexKey(validatorSrcAddr sdk.AccAddress) []byte {
	return append(RedelegationByValSrcIndexKey, validatorSrcAddr.Bytes()...)
}

// get the prefix keyspace for all redelegations redelegating towards a destination validator
func GetREDsToValDstIndexKey(validatorDstAddr sdk.AccAddress) []byte {
	return append(RedelegationByValDstIndexKey, validatorDstAddr.Bytes()...)
}

// get the prefix keyspace for all redelegations redelegating towards a destination validator
// from a particular delegator
func GetREDsByDelToValDstIndexKey(delegatorAddr, validatorDstAddr sdk.AccAddress) []byte {
	return append(
		GetREDsToValDstIndexKey(validatorDstAddr),
		delegatorAddr.Bytes()...)
}
