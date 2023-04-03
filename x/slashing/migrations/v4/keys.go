package v4

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const (
	addrLen = 20

	missedBlockBitmapChunkSize = 1024 // 2^10 bits
)

var (
	validatorSigningInfoKeyPrefix         = []byte{0x01}
	validatorMissedBlockBitArrayKeyPrefix = []byte{0x02}
)

func validatorSigningInfoAddress(key []byte) (v sdk.ConsAddress) {
	kv.AssertKeyAtLeastLength(key, 2)
	addr := key[1:]
	kv.AssertKeyLength(addr, addrLen)
	return sdk.ConsAddress(addr)
}

func validatorMissedBlockBitArrayPrefixKey(v sdk.ConsAddress) []byte {
	return append(validatorMissedBlockBitArrayKeyPrefix, v.Bytes()...)
}

func validatorMissedBlockBitArrayKey(v sdk.ConsAddress, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append(validatorMissedBlockBitArrayPrefixKey(v), b...)
}

func validatorMissedBlockBitmapPrefixKey(v sdk.ConsAddress) []byte {
	return append(validatorMissedBlockBitArrayKeyPrefix, address.MustLengthPrefix(v.Bytes())...)
}

func validatorMissedBlockBitmapKey(v sdk.ConsAddress, chunkIndex int64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(chunkIndex))

	return append(validatorMissedBlockBitmapPrefixKey(v), bz...)
}
