package types

import (
	"encoding/binary"
	"fmt"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "distribution"

	// StoreKey is the store key string for distribution
	StoreKey = ModuleName

	// RouterKey is the message route for distribution
	RouterKey = ModuleName

	// GovModuleName is the name of the gov module
	GovModuleName = "gov"
)

// Keys for distribution store
// Items are stored with the following key: values
//
// - 0x00<proposalID_Bytes>: FeePol
//
// - 0x01: sdk.ConsAddress
//
// - 0x02<valAddrLen (1 Byte)><valAddr_Bytes>: ValidatorOutstandingRewards
//
// - 0x03<accAddrLen (1 Byte)><accAddr_Bytes>: sdk.AccAddress
//
// - 0x04<valAddrLen (1 Byte)><valAddr_Bytes><accAddrLen (1 Byte)><accAddr_Bytes>: DelegatorStartingInfo
//
// - 0x05<valAddrLen (1 Byte)><valAddr_Bytes><period_Bytes>: ValidatorHistoricalRewards
//
// - 0x06<valAddrLen (1 Byte)><valAddr_Bytes>: ValidatorCurrentRewards
//
// - 0x07<valAddrLen (1 Byte)><valAddr_Bytes>: ValidatorCurrentCommission
//
// - 0x08<valAddrLen (1 Byte)><valAddr_Bytes><height>: ValidatorSlashEvent
//
// - 0x09: Params
var (
	FeePoolKey                           = collections.NewPrefix(0) // key for global distribution state
	ProposerKey                          = []byte{0x01}             // key for the proposer operator address
	ValidatorOutstandingRewardsPrefix    = collections.NewPrefix(2) // key for outstanding rewards
	DelegatorWithdrawAddrPrefix          = collections.NewPrefix(3) // key for delegator withdraw address
	DelegatorStartingInfoPrefix          = collections.NewPrefix(4) // key for delegator starting info
	ValidatorHistoricalRewardsPrefix     = collections.NewPrefix(5) // key for historical validators rewards / stake
	ValidatorCurrentRewardsPrefix        = collections.NewPrefix(6) // key for current validator rewards
	ValidatorAccumulatedCommissionPrefix = collections.NewPrefix(7) // key for accumulated validator commission
	ValidatorSlashEventPrefix            = []byte{0x08}             // key for validator slash fraction
	ParamsKey                            = collections.NewPrefix(9) // key for distribution module params
)

// LEUint64Key is a collections KeyCodec that encodes uint64 using little endian.
// NOTE: it MUST NOT be used by other modules, distribution relies on this only for
// state backwards compatibility.
var LEUint64Key collcodec.KeyCodec[uint64] = leUint64Key{}

type leUint64Key struct{}

func (l leUint64Key) Encode(buffer []byte, key uint64) (int, error) {
	binary.LittleEndian.PutUint64(buffer, key)
	return 8, nil
}

func (l leUint64Key) Decode(buffer []byte) (int, uint64, error) {
	if size := len(buffer); size < 8 {
		return 0, 0, fmt.Errorf("invalid buffer size, wanted 8 at least got %d", size)
	}
	return 8, binary.LittleEndian.Uint64(buffer), nil
}

func (l leUint64Key) Size(_ uint64) int { return 8 }

func (l leUint64Key) EncodeJSON(value uint64) ([]byte, error) {
	return collections.Uint64Key.EncodeJSON(value)
}

func (l leUint64Key) DecodeJSON(b []byte) (uint64, error) { return collections.Uint64Key.DecodeJSON(b) }

func (l leUint64Key) Stringify(key uint64) string { return collections.Uint64Key.Stringify(key) }

func (l leUint64Key) KeyType() string { return "little-endian-uint64" }

func (l leUint64Key) EncodeNonTerminal(buffer []byte, key uint64) (int, error) {
	return l.Encode(buffer, key)
}

func (l leUint64Key) DecodeNonTerminal(buffer []byte) (int, uint64, error) { return l.Decode(buffer) }

func (l leUint64Key) SizeNonTerminal(_ uint64) int { return 8 }

// GetValidatorSlashEventAddressHeight creates the height from a validator's slash event key.
func GetValidatorSlashEventAddressHeight(key []byte) (valAddr sdk.ValAddress, height uint64) {
	// key is in the format:
	// 0x08<valAddrLen (1 Byte)><valAddr_Bytes><height>: ValidatorSlashEvent
	kv.AssertKeyAtLeastLength(key, 2)
	valAddrLen := int(key[1])
	kv.AssertKeyAtLeastLength(key, 3+valAddrLen)
	valAddr = key[2 : 2+valAddrLen]
	startB := 2 + valAddrLen
	kv.AssertKeyAtLeastLength(key, startB+9)
	b := key[startB : startB+8] // the next 8 bytes represent the height
	height = binary.BigEndian.Uint64(b)
	return
}

// GetValidatorSlashEventPrefix creates the prefix key for a validator's slash fractions.
func GetValidatorSlashEventPrefix(v sdk.ValAddress) []byte {
	return append(ValidatorSlashEventPrefix, address.MustLengthPrefix(v.Bytes())...)
}

// GetValidatorSlashEventKeyPrefix creates the prefix key for a validator's slash fraction (ValidatorSlashEventPrefix + height).
func GetValidatorSlashEventKeyPrefix(v sdk.ValAddress, height uint64) []byte {
	heightBz := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBz, height)

	return append(
		ValidatorSlashEventPrefix,
		append(address.MustLengthPrefix(v.Bytes()), heightBz...)...,
	)
}

// GetValidatorSlashEventKey creates the key for a validator's slash fraction.
func GetValidatorSlashEventKey(v sdk.ValAddress, height, period uint64) []byte {
	periodBz := make([]byte, 8)
	binary.BigEndian.PutUint64(periodBz, period)
	prefix := GetValidatorSlashEventKeyPrefix(v, height)

	return append(prefix, periodBz...)
}
