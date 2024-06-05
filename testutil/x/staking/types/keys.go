package types

import (
	"encoding/binary"

	"cosmossdk.io/collections"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	ValidatorsKey             = collections.NewPrefix(33) // prefix for each key to a validator
	ValidatorsByConsAddrKey   = collections.NewPrefix(34) // prefix for each key to a validator index, by pubkey
	ValidatorsByPowerIndexKey = []byte{0x23}              // prefix for each key to a validator index, sorted by power

	DelegationKey = collections.NewPrefix(49) // key for a delegation

	ValidatorQueueKey = collections.NewPrefix(67) // prefix for the timestamps in validator queue

	ParamsKey = collections.NewPrefix(81) // prefix for parameters for module x/staking
)

// Reserved kvstore keys
var (
	_ = collections.NewPrefix(97)
)

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
