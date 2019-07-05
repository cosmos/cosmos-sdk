package types

import (
	"encoding/binary"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the staking module
	ModuleName = "staking"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + ModuleName

	// QuerierRoute is the querier route for the staking module
	QuerierRoute = ModuleName

	// RouterKey is the msg router key for the staking module
	RouterKey = ModuleName
)

//nolint
var (
	// Keys for store prefixes
	// Last* values are constant during a block.
	LastValidatorPowerKey = []byte{0x11} // prefix for each key to a validator index, for bonded validators
	LastTotalPowerKey     = []byte{0x12} // prefix for the total power

	ValidatorsKey             = []byte{0x21} // prefix for each key to a validator
	ValidatorsByConsAddrKey   = []byte{0x22} // prefix for each key to a validator index, by pubkey
	ValidatorsByPowerIndexKey = []byte{0x23} // prefix for each key to a validator index, sorted by power

	DelegationKey                    = []byte{0x31} // key for a delegation
	UnbondingDelegationKey           = []byte{0x32} // key for an unbonding-delegation
	UnbondingDelegationByValIndexKey = []byte{0x33} // prefix for each key for an unbonding-delegation, by validator operator
	RedelegationKey                  = []byte{0x34} // key for a redelegation
	RedelegationByValSrcIndexKey     = []byte{0x35} // prefix for each key for an redelegation, by source validator operator
	RedelegationByValDstIndexKey     = []byte{0x36} // prefix for each key for an redelegation, by destination validator operator

	UnbondingQueueKey    = []byte{0x41} // prefix for the timestamps in unbonding queue
	RedelegationQueueKey = []byte{0x42} // prefix for the timestamps in redelegations queue
	ValidatorQueueKey    = []byte{0x43} // prefix for the timestamps in validator queue
)

// gets the key for the validator with address
// VALUE: staking/Validator
func GetValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorsKey, operatorAddr.Bytes()...)
}

// gets the key for the validator with pubkey
// VALUE: validator operator address ([]byte)
func GetValidatorByConsAddrKey(addr sdk.ConsAddress) []byte {
	return append(ValidatorsByConsAddrKey, addr.Bytes()...)
}

// Get the validator operator address from LastValidatorPowerKey
func AddressFromLastValidatorPowerKey(key []byte) []byte {
	return key[1:] // remove prefix bytes
}

// get the validator by power index.
// Power index is the key used in the power-store, and represents the relative
// power ranking of the validator.
// VALUE: validator operator address ([]byte)
func GetValidatorsByPowerIndexKey(validator Validator) []byte {
	// NOTE the address doesn't need to be stored because counter bytes must always be different
	return getValidatorPowerRank(validator)
}

// get the bonded validator index key for an operator address
func GetLastValidatorPowerKey(operator sdk.ValAddress) []byte {
	return append(LastValidatorPowerKey, operator...)
}

// get the power ranking of a validator
// NOTE the larger values are of higher value
func getValidatorPowerRank(validator Validator) []byte {

	consensusPower := sdk.TokensToConsensusPower(validator.Tokens)
	consensusPowerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(consensusPowerBytes[:], uint64(consensusPower))

	powerBytes := consensusPowerBytes
	powerBytesLen := len(powerBytes) // 8

	// key is of format prefix || powerbytes || addrBytes
	key := make([]byte, 1+powerBytesLen+sdk.AddrLen)

	key[0] = ValidatorsByPowerIndexKey[0]
	copy(key[1:powerBytesLen+1], powerBytes)
	operAddrInvr := sdk.CopyBytes(validator.OperatorAddress)
	for i, b := range operAddrInvr {
		operAddrInvr[i] = ^b
	}
	copy(key[powerBytesLen+1:], operAddrInvr)

	return key
}

// parse the validators operator address from power rank key
func ParseValidatorPowerRankKey(key []byte) (operAddr []byte) {
	powerBytesLen := 8
	if len(key) != 1+powerBytesLen+sdk.AddrLen {
		panic("Invalid validator power rank key length")
	}
	operAddr = sdk.CopyBytes(key[powerBytesLen+1:])
	for i, b := range operAddr {
		operAddr[i] = ^b
	}
	return operAddr
}

// gets the prefix for all unbonding delegations from a delegator
func GetValidatorQueueTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(ValidatorQueueKey, bz...)
}

//______________________________________________________________________________

// gets the key for delegator bond with validator
// VALUE: staking/Delegation
func GetDelegationKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(GetDelegationsKey(delAddr), valAddr.Bytes()...)
}

// gets the prefix for a delegator for all validators
func GetDelegationsKey(delAddr sdk.AccAddress) []byte {
	return append(DelegationKey, delAddr.Bytes()...)
}

//______________________________________________________________________________

// gets the key for an unbonding delegation by delegator and validator addr
// VALUE: staking/UnbondingDelegation
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

// gets the prefix for all unbonding delegations from a delegator
func GetUnbondingDelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(UnbondingQueueKey, bz...)
}

//________________________________________________________________________________

// gets the key for a redelegation
// VALUE: staking/RedelegationKey
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

// GetREDKeyFromValSrcIndexKey rearranges the ValSrcIndexKey to get the REDKey
func GetREDKeyFromValSrcIndexKey(indexKey []byte) []byte {
	// note that first byte is prefix byte
	if len(indexKey) != 3*sdk.AddrLen+1 {
		panic("unexpected key length")
	}
	valSrcAddr := indexKey[1 : sdk.AddrLen+1]
	delAddr := indexKey[sdk.AddrLen+1 : 2*sdk.AddrLen+1]
	valDstAddr := indexKey[2*sdk.AddrLen+1 : 3*sdk.AddrLen+1]

	return GetREDKey(delAddr, valSrcAddr, valDstAddr)
}

// GetREDKeyFromValDstIndexKey rearranges the ValDstIndexKey to get the REDKey
func GetREDKeyFromValDstIndexKey(indexKey []byte) []byte {
	// note that first byte is prefix byte
	if len(indexKey) != 3*sdk.AddrLen+1 {
		panic("unexpected key length")
	}
	valDstAddr := indexKey[1 : sdk.AddrLen+1]
	delAddr := indexKey[sdk.AddrLen+1 : 2*sdk.AddrLen+1]
	valSrcAddr := indexKey[2*sdk.AddrLen+1 : 3*sdk.AddrLen+1]
	return GetREDKey(delAddr, valSrcAddr, valDstAddr)
}

// gets the prefix for all unbonding delegations from a delegator
func GetRedelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(RedelegationQueueKey, bz...)
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
