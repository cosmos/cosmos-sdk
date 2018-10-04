package keeper

import (
	"encoding/binary"

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
	ValidatorsByConsAddrKey          = []byte{0x03} // prefix for each key to a validator index, by pubkey
	ValidatorsBondedIndexKey         = []byte{0x04} // prefix for each key to a validator index, for bonded validators
	ValidatorsByPowerIndexKey        = []byte{0x05} // prefix for each key to a validator index, sorted by power
	IntraTxCounterKey                = []byte{0x06} // key for intra-block tx index
	DelegationKey                    = []byte{0x07} // key for a delegation
	UnbondingDelegationKey           = []byte{0x08} // key for an unbonding-delegation
	UnbondingDelegationByValIndexKey = []byte{0x09} // prefix for each key for an unbonding-delegation, by validator operator
	RedelegationKey                  = []byte{0x0A} // key for a redelegation
	RedelegationByValSrcIndexKey     = []byte{0x0B} // prefix for each key for an redelegation, by source validator operator
	RedelegationByValDstIndexKey     = []byte{0x0C} // prefix for each key for an redelegation, by destination validator operator
)

const maxDigitsForAccount = 12 // ~220,000,000 atoms created at launch

// gets the key for the validator with address
// VALUE: stake/types.Validator
func GetValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorsKey, operatorAddr.Bytes()...)
}

// gets the key for the validator with pubkey
// VALUE: validator operator address ([]byte)
func GetValidatorByConsAddrKey(addr sdk.ConsAddress) []byte {
	return append(ValidatorsByConsAddrKey, addr.Bytes()...)
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
	return getValidatorPowerRank(validator)
}

// get the bonded validator index key for an operator address
func GetBondedValidatorIndexKey(operator sdk.ValAddress) []byte {
	return append(ValidatorsBondedIndexKey, operator...)
}

// get the power ranking of a validator
// NOTE the larger values are of higher value
// nolint: unparam
func getValidatorPowerRank(validator types.Validator) []byte {

	potentialPower := validator.Tokens
	powerBytes := []byte(potentialPower.ToLeftPadded(maxDigitsForAccount)) // power big-endian (more powerful validators first)
	powerBytesLen := len(powerBytes)
	// key is of format prefix || powerbytes || heightBytes || counterBytes
	key := make([]byte, 1+powerBytesLen+8+2)

	key[0] = ValidatorsByPowerIndexKey[0]
	copy(key[1:powerBytesLen+1], powerBytes)

	// include heightBytes height is inverted (older validators first)
	binary.BigEndian.PutUint64(key[powerBytesLen+1:powerBytesLen+9], ^uint64(validator.BondHeight))
	// include counterBytes, counter is inverted (first txns have priority)
	binary.BigEndian.PutUint16(key[powerBytesLen+9:powerBytesLen+11], ^uint16(validator.BondIntraTxCounter))

	return key
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
	key := make([]byte, 1+sdk.AddrLen*3)

	copy(key[0:sdk.AddrLen+1], GetREDsKey(delAddr.Bytes()))
	copy(key[sdk.AddrLen+1:2*sdk.AddrLen+1], valSrcAddr.Bytes())
	copy(key[2*sdk.AddrLen+1:3*sdk.AddrLen+1], valDstAddr.Bytes())

	return key
}

// gets the index-key for a redelegation, stored by source-validator-index
// VALUE: none (key rearrangement used)
func GetREDByValSrcIndexKey(delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) []byte {
	REDSFromValsSrcKey := GetREDsFromValSrcIndexKey(valSrcAddr)
	offset := len(REDSFromValsSrcKey)

	// key is of the form REDSFromValsSrcKey || delAddr || valDstAddr
	key := make([]byte, len(REDSFromValsSrcKey)+2*sdk.AddrLen)
	copy(key[0:offset], REDSFromValsSrcKey)
	copy(key[offset:offset+sdk.AddrLen], delAddr.Bytes())
	copy(key[offset+sdk.AddrLen:offset+2*sdk.AddrLen], valDstAddr.Bytes())
	return key
}

// gets the index-key for a redelegation, stored by destination-validator-index
// VALUE: none (key rearrangement used)
func GetREDByValDstIndexKey(delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) []byte {
	REDSToValsDstKey := GetREDsToValDstIndexKey(valDstAddr)
	offset := len(REDSToValsDstKey)

	// key is of the form REDSToValsDstKey || delAddr || valSrcAddr
	key := make([]byte, len(REDSToValsDstKey)+2*sdk.AddrLen)
	copy(key[0:offset], REDSToValsDstKey)
	copy(key[offset:offset+sdk.AddrLen], delAddr.Bytes())
	copy(key[offset+sdk.AddrLen:offset+2*sdk.AddrLen], valSrcAddr.Bytes())

	return key
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
