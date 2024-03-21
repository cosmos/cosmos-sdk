package v4

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const MissedBlockBitmapChunkSize = 1024 // 2^10 bits

var (
	ValidatorSigningInfoKeyPrefix                   = []byte{0x01}
	deprecatedValidatorMissedBlockBitArrayKeyPrefix = []byte{0x02}

	// NOTE: sdk v0.50 uses the same key prefix for both deprecated and new missed block bitmaps.
	// We needed to use a new key, because we are skipping deletion of all old keys at upgrade time
	// due to how long this would bring the chain down. We use 0x10 here to prevent overlap with any future keys.
	validatorMissedBlockBitMapKeyPrefix = []byte{0x10}
)

func ValidatorSigningInfoKey(v sdk.ConsAddress) []byte {
	return append(ValidatorSigningInfoKeyPrefix, address.MustLengthPrefix(v.Bytes())...)
}

func ValidatorSigningInfoAddress(key []byte) (v sdk.ConsAddress) {
	// Remove prefix and address length.
	kv.AssertKeyAtLeastLength(key, 3)
	addr := key[2:]

	return sdk.ConsAddress(addr)
}

func DeprecatedValidatorMissedBlockBitArrayPrefixKey(v sdk.ConsAddress) []byte {
	return append(deprecatedValidatorMissedBlockBitArrayKeyPrefix, address.MustLengthPrefix(v.Bytes())...)
}

func DeprecatedValidatorMissedBlockBitArrayKey(v sdk.ConsAddress, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append(DeprecatedValidatorMissedBlockBitArrayPrefixKey(v), b...)
}

func validatorMissedBlockBitmapPrefixKey(v sdk.ConsAddress) []byte {
	return append(validatorMissedBlockBitMapKeyPrefix, address.MustLengthPrefix(v.Bytes())...)
}

func ValidatorMissedBlockBitmapKey(v sdk.ConsAddress, chunkIndex int64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(chunkIndex))

	return append(validatorMissedBlockBitmapPrefixKey(v), bz...)
}
