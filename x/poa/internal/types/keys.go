package types

import (
	"encoding/binary"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleNa,e is the name of the POA module
	ModuleName = "poa"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + ModuleName

	// QuerierRoute is the querier route for the POA module
	QuerierRoute = ModuleName

	// RouterKey is the msg router key for POA module
	RouterKey = ModuleName
)

var (
	// Keys for store prefixes
	// Last* values are constant during a block.
	LastValidatorPowerKey = []byte{0x11} // prefix for each key to a validator index, for bonded validators
	LastTotalPowerKey     = []byte{0x12} // prefix for the total power

	ValidatorsKey             = []byte{0x21} // prefix for each key to a validator
	ValidatorsByConsAddrKey   = []byte{0x22} // prefix for each key to a validator index, by pubkey
	ValidatorsByPowerIndexKey = []byte{0x23} // prefix for each key to a validator index, sorted by power

	UnbondingQueueKey = []byte{0x41} // prefix for the timestamps in unbonding queue
	ValidatorQueueKey = []byte{0x43} // prefix for the timestamps in validator queue
)

// gets the key for the validator with address
// VALUE: poa/Validator
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

	consensusPower := sdk.TokensToConsensusPower(validator.Weight)
	consensusPowerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(consensusPowerBytes, uint64(consensusPower))

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
