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
	UnbondingDelegationByValIndexKey = []byte{0x0C} // prefix for each key for an unbonding-delegation, by validator operator
	RedelegationKey                  = []byte{0x0D} // key for a redelegation
	RedelegationByValSrcIndexKey     = []byte{0x0E} // prefix for each key for an redelegation, by source validator operator
	RedelegationByValDstIndexKey     = []byte{0x0F} // prefix for each key for an redelegation, by destination validator operator
)

const maxDigitsForAccount = 12 // ~220,000,000 atoms created at launch

// gets the key for the validator with address
// VALUE: stake/types.Validator
func GetValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorsKey, operatorAddr.Bytes()...)
}

// gets the key for the validator with pubkey
// VALUE: validator operator address ([]byte)
func GetValidatorByPubKeyIndexKey(pubkey crypto.PubKey) []byte {
	return append(ValidatorsByPubKeyIndexKey, pubkey.Bytes()...)
}

// gets the key for the current validator group
// VALUE: none (key rearrangement with GetValKeyFromValBondedIndexKey)
func GetValidatorsBondedIndexKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorsBondedIndexKey, operatorAddr.Bytes()...)
}

// Get the validator operator address from ValBondedIndexKey
func GetAddressFromValBondedIndexKey(IndexKey []byte) []byte {
	return IndexKey[1:] // remove prefix bytes
}

// get the validator by power index.
// Power index is the key used in the power-store, and represents the relative
// power ranking of the validator.
// VALUE: validator operator address ([]byte)
func GetValidatorsByPowerIndexKey(validator types.Validator, pool types.Pool) []byte {
	// NOTE the address doesn't need to be stored because counter bytes must always be different
	return getValidatorPowerRank(validator, pool)
}

// get the power ranking of a validator
// NOTE the larger values are of higher value
// nolint: unparam
func getValidatorPowerRank(validator types.Validator, pool types.Pool) []byte {

	potentialPower := validator.Tokens
	powerBytes := []byte(potentialPower.ToLeftPadded(maxDigitsForAccount)) // power big-endian (more powerful validators first)

	jailedBytes := make([]byte, 1)
	if validator.Jailed {
		jailedBytes[0] = byte(0x00)
	} else {
		jailedBytes[0] = byte(0x01)
	}

	// heightBytes and counterBytes represent strings like powerBytes does
	heightBytes := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(heightBytes, ^uint64(validator.BondHeight)) // invert height (older validators first)
	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, ^uint16(validator.BondIntraTxCounter)) // invert counter (first txns have priority)

	return append(append(append(append(
		ValidatorsByPowerIndexKey,
		jailedBytes...),
		powerBytes...),
		heightBytes...),
		counterBytes...)
}

// get the key for the accumulated update validators
// VALUE: abci.Validator
// note records using these keys should never persist between blocks
func GetTendermintUpdatesKey(operatorAddr sdk.ValAddress) []byte {
	return append(TendermintUpdatesKey, operatorAddr.Bytes()...)
}

//______________________________________________________________________________

// gets the key for delegator bond with validator
// VALUE: stake/types.Delegation
func GetDelegationKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(GetDelegationsKey(delAddr), valAddr.Bytes()...)
}

// gets the prefix for a delegator for all validators
func GetDelegationsKey(delAddr sdk.AccAddress) []byte {
	return append(DelegationKey, delAddr.Bytes()...)
}

//______________________________________________________________________________

// gets the key for an unbonding delegation by delegator and validator addr
// VALUE: stake/types.UnbondingDelegation
func GetUBDKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(
		GetUBDsKey(delAddr.Bytes()),
		valAddr.Bytes()...)
}

// gets the index-key for an unbonding delegation, stored by validator-index
// VALUE: none (key rearrangement used)
func GetUBDByValIndexKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(GetUBDsByValIndexKey(valAddr), delAddr.Bytes()...)
}

// rearranges the ValIndexKey to get the UBDKey
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

// gets the prefix for all unbonding delegations from a delegator
func GetUBDsKey(delAddr sdk.AccAddress) []byte {
	return append(UnbondingDelegationKey, delAddr.Bytes()...)
}

// gets the prefix keyspace for the indexes of unbonding delegations for a validator
func GetUBDsByValIndexKey(valAddr sdk.ValAddress) []byte {
	return append(UnbondingDelegationByValIndexKey, valAddr.Bytes()...)
}

//________________________________________________________________________________

// gets the key for a redelegation
// VALUE: stake/types.RedelegationKey
func GetREDKey(delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) []byte {
	return append(append(
		GetREDsKey(delAddr.Bytes()),
		valSrcAddr.Bytes()...),
		valDstAddr.Bytes()...)
}

// gets the index-key for a redelegation, stored by source-validator-index
// VALUE: none (key rearrangement used)
func GetREDByValSrcIndexKey(delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) []byte {
	return append(append(
		GetREDsFromValSrcIndexKey(valSrcAddr),
		delAddr.Bytes()...),
		valDstAddr.Bytes()...)
}

// gets the index-key for a redelegation, stored by destination-validator-index
// VALUE: none (key rearrangement used)
func GetREDByValDstIndexKey(delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) []byte {
	return append(append(
		GetREDsToValDstIndexKey(valDstAddr),
		delAddr.Bytes()...),
		valSrcAddr.Bytes()...)
}

// rearranges the ValSrcIndexKey to get the REDKey
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

// rearranges the ValDstIndexKey to get the REDKey
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

// gets the prefix keyspace for redelegations from a delegator
func GetREDsKey(delAddr sdk.AccAddress) []byte {
	return append(RedelegationKey, delAddr.Bytes()...)
}

// gets the prefix keyspace for all redelegations redelegating away from a source validator
func GetREDsFromValSrcIndexKey(valSrcAddr sdk.ValAddress) []byte {
	return append(RedelegationByValSrcIndexKey, valSrcAddr.Bytes()...)
}

// gets the prefix keyspace for all redelegations redelegating towards a destination validator
func GetREDsToValDstIndexKey(valDstAddr sdk.ValAddress) []byte {
	return append(RedelegationByValDstIndexKey, valDstAddr.Bytes()...)
}

// gets the prefix keyspace for all redelegations redelegating towards a destination validator
// from a particular delegator
func GetREDsByDelToValDstIndexKey(delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) []byte {
	return append(
		GetREDsToValDstIndexKey(valDstAddr),
		delAddr.Bytes()...)
}
