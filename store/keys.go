package store

import (
	"bytes"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PrefixKeyString returns a byte slice consisting of the prefix string concatenated
// with all subkeys.
func PrefixKeyString(prefix string, subkeys ...[]byte) []byte {
	buf := [][]byte{[]byte(prefix)}
	return bytes.Join(append(buf, subkeys...), []byte("/"))
}

// Int64Subkey returns a byte slice from the provided subkey suitable for use
// as a key in the store. This method will panic on negative numbers.
func Int64Subkey(subkey int64) []byte {
	if subkey < 0 {
		panic("cannot use negative numbers in subkeys")
	}
	return Uint64Subkey(uint64(subkey))
}

// Uint64Subkey returns a byte slice from the provided subkey suitable for use
// as a key in the store.
func Uint64Subkey(subkey uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, subkey)
	return b
}

// SDKUintSubkey returns a byte slice from the provided subkey suitable for use
// an iterable key in the store. The bytes representing the number will be left-padded
// to padToBits bits. If padTo < len(subkey.Bytes()), or padToBits % 8 != 0,
// this method will panic.
func SDKUintSubkey(subkey sdk.Uint, padToBits int) []byte {
	if padToBits%8 != 0 {
		panic("padToBits must be divisible by 8")
	}
	padToBytes := padToBits / 8
	buf := make([]byte, padToBytes, padToBytes)
	b := subkey.BigInt().Bytes()
	copy(buf[padToBytes-len(b):], b)
	return buf
}
