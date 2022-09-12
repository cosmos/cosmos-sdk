package types

import (
	"encoding/binary"

	"cosmossdk.io/collections"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/math"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const (
	// ModuleName is the name of the staking module
	ModuleName = "staking"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the staking module
	RouterKey = ModuleName

	// GovModuleName is the name of the gov module
	GovModuleName = "gov"

	// PoolModuleName duplicates the Protocolpool module's name to avoid a cyclic dependency with x/protocolpool.
	// It should be synced with the distribution module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/912390d5fc4a32113ea1aacc98b77b2649aea4c2/x/distribution/types/keys.go#L15
	PoolModuleName = "protocolpool"
)

var (
	// Keys for store prefixes
	// Last* values are constant during a block.
	LastValidatorPowerKey = collections.NewPrefix(17) // prefix for each key to a validator index, for bonded validators
	LastTotalPowerKey     = collections.NewPrefix(18) // prefix for the total power

	ValidatorsKey             = collections.NewPrefix(33) // prefix for each key to a validator
	ValidatorsByConsAddrKey   = collections.NewPrefix(34) // prefix for each key to a validator index, by pubkey
	ValidatorsByPowerIndexKey = []byte{0x23}              // prefix for each key to a validator index, sorted by power

	DelegationKey                    = collections.NewPrefix(49) // key for a delegation
	UnbondingDelegationKey           = collections.NewPrefix(50) // key for an unbonding-delegation
	UnbondingDelegationByValIndexKey = collections.NewPrefix(51) // prefix for each key for an unbonding-delegation, by validator operator

	RedelegationKey              = collections.NewPrefix(52) // key for a redelegation
	RedelegationByValSrcIndexKey = collections.NewPrefix(53) // prefix for each key for a redelegation, by source validator operator
	RedelegationByValDstIndexKey = collections.NewPrefix(54) // prefix for each key for a redelegation, by destination validator operator

	UnbondingIDKey    = collections.NewPrefix(55) // key for the counter for the incrementing id for UnbondingOperations
	UnbondingIndexKey = collections.NewPrefix(56) // prefix for an index for looking up unbonding operations by their IDs
	UnbondingTypeKey  = collections.NewPrefix(57) // prefix for an index containing the type of unbonding operations

	UnbondingQueueKey    = collections.NewPrefix(65) // prefix for the timestamps in unbonding queue
	RedelegationQueueKey = collections.NewPrefix(66) // prefix for the timestamps in redelegations queue
	ValidatorQueueKey    = collections.NewPrefix(67) // prefix for the timestamps in validator queue

	ParamsKey = collections.NewPrefix(81) // prefix for parameters for module x/staking

	DelegationByValIndexKey = collections.NewPrefix(113) // key for delegations by a validator

	ValidatorConsPubKeyRotationHistoryKey       = collections.NewPrefix(101) // prefix for consPubkey rotation history by validator
	BlockConsPubKeyRotationHistoryKey           = collections.NewPrefix(102) // prefix for consPubkey rotation history by height
	ValidatorConsensusKeyRotationRecordQueueKey = collections.NewPrefix(103) // this key is used to set the unbonding period time on each rotation
	ValidatorConsensusKeyRotationRecordIndexKey = collections.NewPrefix(104) // this key is used to restrict the validator next rotation within waiting (unbonding) period
	ConsAddrToValidatorIdentifierMapPrefix      = collections.NewPrefix(105) // prefix for rotated cons address to new cons address
	OldToNewConsAddrMap                         = collections.NewPrefix(106) // prefix for rotated cons address to new cons address
)

// Reserved kvstore keys
var (
	_ = collections.NewPrefix(97)
)

// UnbondingType defines the type of unbonding operation
type UnbondingType int

const (
	UnbondingType_Undefined UnbondingType = iota
	UnbondingType_UnbondingDelegation
	UnbondingType_Redelegation
	UnbondingType_ValidatorUnbonding
)

// GetValidatorKey creates the key for the validator with address
// VALUE: staking/Validator
func GetValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(ValidatorsKey, address.MustLengthPrefix(operatorAddr)...)
}

// AddressFromValidatorsKey creates the validator operator address from ValidatorsKey
func AddressFromValidatorsKey(key []byte) []byte {
	kv.AssertKeyAtLeastLength(key, 3)
	return key[2:] // remove prefix bytes and address length
}

// AddressFromLastValidatorPowerKey creates the validator operator address from LastValidatorPowerKey
func AddressFromLastValidatorPowerKey(key []byte) []byte {
	kv.AssertKeyAtLeastLength(key, 3)
	return key[2:] // remove prefix bytes and address length
}

// GetValidatorsByPowerIndexKey creates the validator by power index.
// Power index is the key used in the power-store, and represents the relative
// power ranking of the validator.
// VALUE: validator operator address ([]byte)
func GetValidatorsByPowerIndexKey(validator Validator, powerReduction math.Int) []byte {
	// NOTE the address doesn't need to be stored because counter bytes must always be different
	// NOTE the larger values are of higher value

	consensusPower := sdk.TokensToConsensusPower(validator.Tokens, powerReduction)
	consensusPowerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(consensusPowerBytes, uint64(consensusPower))

	powerBytes := consensusPowerBytes
	powerBytesLen := len(powerBytes) // 8

	addr, err := valAc.StringToBytes(validator.OperatorAddress)
	if err != nil {
		panic(err)
	}
	operAddrInvr := sdk.CopyBytes(addr)
	addrLen := len(operAddrInvr)

	for i, b := range operAddrInvr {
		operAddrInvr[i] = ^b
	}

	// key is of format prefix || powerbytes || addrLen (1byte) || addrBytes
	key := make([]byte, 1+powerBytesLen+1+addrLen)

	key[0] = ValidatorsByPowerIndexKey[0]
	copy(key[1:powerBytesLen+1], powerBytes)
	key[powerBytesLen+1] = byte(addrLen)
	copy(key[powerBytesLen+2:], operAddrInvr)

	return key
}

// ParseValidatorPowerRankKey parses the validators operator address from power rank key
func ParseValidatorPowerRankKey(key []byte) (operAddr []byte) {
	powerBytesLen := 8

	// key is of format prefix (1 byte) || powerbytes || addrLen (1byte) || addrBytes
	operAddr = sdk.CopyBytes(key[powerBytesLen+2:])

	for i, b := range operAddr {
		operAddr[i] = ^b
	}

	return operAddr
}

// GetUBDKey creates the key for an unbonding delegation by delegator and validator addr
// VALUE: staking/UnbondingDelegation
func GetUBDKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(GetUBDsKey(delAddr.Bytes()), address.MustLengthPrefix(valAddr)...)
}

// GetUBDByValIndexKey creates the index-key for an unbonding delegation, stored by validator-index
// VALUE: none (key rearrangement used)
func GetUBDByValIndexKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(GetUBDsByValIndexKey(valAddr), address.MustLengthPrefix(delAddr)...)
}

// GetUBDKeyFromValIndexKey rearranges the ValIndexKey to get the UBDKey
func GetUBDKeyFromValIndexKey(indexKey []byte) []byte {
	kv.AssertKeyAtLeastLength(indexKey, 2)
	addrs := indexKey[1:] // remove prefix bytes

	valAddrLen := addrs[0]
	kv.AssertKeyAtLeastLength(addrs, 2+int(valAddrLen))
	valAddr := addrs[1 : 1+valAddrLen]
	kv.AssertKeyAtLeastLength(addrs, 3+int(valAddrLen))
	delAddr := addrs[valAddrLen+2:]

	return GetUBDKey(delAddr, valAddr)
}

// GetUBDsKey creates the prefix for all unbonding delegations from a delegator
func GetUBDsKey(delAddr sdk.AccAddress) []byte {
	return append(UnbondingDelegationKey, address.MustLengthPrefix(delAddr)...)
}

// GetUBDsByValIndexKey creates the prefix keyspace for the indexes of unbonding delegations for a validator
func GetUBDsByValIndexKey(valAddr sdk.ValAddress) []byte {
	return append(UnbondingDelegationByValIndexKey, address.MustLengthPrefix(valAddr)...)
}

// GetUnbondingDelegationTimeKey creates the prefix for all unbonding delegations from a delegator
func GetUnbondingDelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(UnbondingQueueKey, bz...)
}

// GetREDKey returns a key prefix for indexing a redelegation from a delegator
// and source validator to a destination validator.
func GetREDKey(delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) []byte {
	// key is of the form GetREDsKey || valSrcAddrLen (1 byte) || valSrcAddr || valDstAddrLen (1 byte) || valDstAddr
	key := make([]byte, 1+3+len(delAddr)+len(valSrcAddr)+len(valDstAddr))

	copy(key[0:2+len(delAddr)], GetREDsKey(delAddr.Bytes()))
	key[2+len(delAddr)] = byte(len(valSrcAddr))
	copy(key[3+len(delAddr):3+len(delAddr)+len(valSrcAddr)], valSrcAddr.Bytes())
	key[3+len(delAddr)+len(valSrcAddr)] = byte(len(valDstAddr))
	copy(key[4+len(delAddr)+len(valSrcAddr):], valDstAddr.Bytes())

	return key
}

// GetREDByValSrcIndexKey creates the index-key for a redelegation, stored by source-validator-index
// VALUE: none (key rearrangement used)
func GetREDByValSrcIndexKey(delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) []byte {
	REDSFromValsSrcKey := GetREDsFromValSrcIndexKey(valSrcAddr)
	offset := len(REDSFromValsSrcKey)

	// key is of the form REDSFromValsSrcKey || delAddrLen (1 byte) || delAddr || valDstAddrLen (1 byte) || valDstAddr
	key := make([]byte, offset+2+len(delAddr)+len(valDstAddr))
	copy(key[0:offset], REDSFromValsSrcKey)
	key[offset] = byte(len(delAddr))
	copy(key[offset+1:offset+1+len(delAddr)], delAddr.Bytes())
	key[offset+1+len(delAddr)] = byte(len(valDstAddr))
	copy(key[offset+2+len(delAddr):], valDstAddr.Bytes())

	return key
}

// GetREDByValDstIndexKey creates the index-key for a redelegation, stored by destination-validator-index
// VALUE: none (key rearrangement used)
func GetREDByValDstIndexKey(delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) []byte {
	REDSToValsDstKey := GetREDsToValDstIndexKey(valDstAddr)
	offset := len(REDSToValsDstKey)

	// key is of the form REDSToValsDstKey || delAddrLen (1 byte) || delAddr || valSrcAddrLen (1 byte) || valSrcAddr
	key := make([]byte, offset+2+len(delAddr)+len(valSrcAddr))
	copy(key[0:offset], REDSToValsDstKey)
	key[offset] = byte(len(delAddr))
	copy(key[offset+1:offset+1+len(delAddr)], delAddr.Bytes())
	key[offset+1+len(delAddr)] = byte(len(valSrcAddr))
	copy(key[offset+2+len(delAddr):], valSrcAddr.Bytes())

	return key
}

// GetREDKeyFromValSrcIndexKey rearranges the ValSrcIndexKey to get the REDKey
func GetREDKeyFromValSrcIndexKey(indexKey []byte) []byte {
	// note that first byte is prefix byte, which we remove
	kv.AssertKeyAtLeastLength(indexKey, 2)
	addrs := indexKey[1:]

	valSrcAddrLen := addrs[0]
	kv.AssertKeyAtLeastLength(addrs, int(valSrcAddrLen)+2)
	valSrcAddr := addrs[1 : valSrcAddrLen+1]
	delAddrLen := addrs[valSrcAddrLen+1]
	kv.AssertKeyAtLeastLength(addrs, int(valSrcAddrLen)+int(delAddrLen)+2)
	delAddr := addrs[valSrcAddrLen+2 : valSrcAddrLen+2+delAddrLen]
	kv.AssertKeyAtLeastLength(addrs, int(valSrcAddrLen)+int(delAddrLen)+4)
	valDstAddr := addrs[valSrcAddrLen+delAddrLen+3:]

	return GetREDKey(delAddr, valSrcAddr, valDstAddr)
}

// GetREDKeyFromValDstIndexKey rearranges the ValDstIndexKey to get the REDKey
func GetREDKeyFromValDstIndexKey(indexKey []byte) []byte {
	// note that first byte is prefix byte, which we remove
	kv.AssertKeyAtLeastLength(indexKey, 2)
	addrs := indexKey[1:]

	valDstAddrLen := addrs[0]
	kv.AssertKeyAtLeastLength(addrs, int(valDstAddrLen)+2)
	valDstAddr := addrs[1 : valDstAddrLen+1]
	delAddrLen := addrs[valDstAddrLen+1]
	kv.AssertKeyAtLeastLength(addrs, int(valDstAddrLen)+int(delAddrLen)+3)
	delAddr := addrs[valDstAddrLen+2 : valDstAddrLen+2+delAddrLen]
	kv.AssertKeyAtLeastLength(addrs, int(valDstAddrLen)+int(delAddrLen)+4)
	valSrcAddr := addrs[valDstAddrLen+delAddrLen+3:]

	return GetREDKey(delAddr, valSrcAddr, valDstAddr)
}

// GetRedelegationTimeKey returns a key prefix for indexing an unbonding
// redelegation based on a completion time.
func GetRedelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(RedelegationQueueKey, bz...)
}

// GetREDsKey returns a key prefix for indexing a redelegation from a delegator
// address.
func GetREDsKey(delAddr sdk.AccAddress) []byte {
	return append(RedelegationKey, address.MustLengthPrefix(delAddr)...)
}
