package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	addresscodec "cosmossdk.io/core/address"
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
	RedelegationByValSrcIndexKey = collections.NewPrefix(53) // prefix for each key for an redelegation, by source validator operator
	RedelegationByValDstIndexKey = collections.NewPrefix(54) // prefix for each key for an redelegation, by destination validator operator

	UnbondingIDKey    = collections.NewPrefix(55) // key for the counter for the incrementing id for UnbondingOperations
	UnbondingIndexKey = collections.NewPrefix(56) // prefix for an index for looking up unbonding operations by their IDs
	UnbondingTypeKey  = collections.NewPrefix(57) // prefix for an index containing the type of unbonding operations

	UnbondingQueueKey    = collections.NewPrefix(65) // prefix for the timestamps in unbonding queue
	RedelegationQueueKey = []byte{0x42}              // prefix for the timestamps in redelegations queue
	ValidatorQueueKey    = []byte{0x43}              // prefix for the timestamps in validator queue

	HistoricalInfoKey   = collections.NewPrefix(80) // prefix for the historical info
	ValidatorUpdatesKey = collections.NewPrefix(97) // prefix for the end block validator updates key

	ParamsKey = []byte{0x51} // prefix for parameters for module x/staking

	DelegationByValIndexKey = []byte{0x71} // key for delegations by a validator
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
func GetValidatorsByPowerIndexKey(validator Validator, powerReduction math.Int, valAc addresscodec.Codec) []byte {
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

// GetValidatorQueueKey returns the prefix key used for getting a set of unbonding
// validators whose unbonding completion occurs at the given time and height.
func GetValidatorQueueKey(timestamp time.Time, height int64) []byte {
	heightBz := sdk.Uint64ToBigEndian(uint64(height))
	timeBz := sdk.FormatTimeBytes(timestamp)
	timeBzL := len(timeBz)
	prefixL := len(ValidatorQueueKey)

	bz := make([]byte, prefixL+8+timeBzL+8)

	// copy the prefix
	copy(bz[:prefixL], ValidatorQueueKey)

	// copy the encoded time bytes length
	copy(bz[prefixL:prefixL+8], sdk.Uint64ToBigEndian(uint64(timeBzL)))

	// copy the encoded time bytes
	copy(bz[prefixL+8:prefixL+8+timeBzL], timeBz)

	// copy the encoded height
	copy(bz[prefixL+8+timeBzL:], heightBz)

	return bz
}

// ParseValidatorQueueKey returns the encoded time and height from a key created
// from GetValidatorQueueKey.
func ParseValidatorQueueKey(bz []byte) (time.Time, int64, error) {
	prefixL := len(ValidatorQueueKey)
	if prefix := bz[:prefixL]; !bytes.Equal(prefix, ValidatorQueueKey) {
		return time.Time{}, 0, fmt.Errorf("invalid prefix; expected: %X, got: %X", ValidatorQueueKey, prefix)
	}

	timeBzL := sdk.BigEndianToUint64(bz[prefixL : prefixL+8])
	ts, err := sdk.ParseTimeBytes(bz[prefixL+8 : prefixL+8+int(timeBzL)])
	if err != nil {
		return time.Time{}, 0, err
	}

	height := sdk.BigEndianToUint64(bz[prefixL+8+int(timeBzL):])

	return ts, int64(height), nil
}

// GetUBDKey creates the key for an unbonding delegation by delegator and validator addr
// VALUE: staking/UnbondingDelegation
func GetUBDKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(append(UnbondingDelegationKey, address.MustLengthPrefix(delAddr)...), address.MustLengthPrefix(valAddr)...)
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
