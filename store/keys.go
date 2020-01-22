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
	return PrefixKeyBytes(append(buf, subkeys...)...)
}

// PrefixKeyBytes returns a byte slice consisting of all subkeys concatenated together.
func PrefixKeyBytes(subkeys ...[]byte) []byte {
	if len(subkeys) == 0 {
		return []byte{}
	}

	var buf bytes.Buffer
	buf.Write(subkeys[0])

	if len(subkeys) > 1 {
		for _, sk := range subkeys[1:] {
			if len(sk) == 0 {
				continue
			}

			buf.WriteRune('/')
			buf.Write(sk)
		}
	}

	return buf.Bytes()
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
// an iterable key in the store.
func SDKUintSubkey(subkey sdk.Uint) []byte {
	var buf [32]byte
	b := subkey.BigInt().Bytes()
	copy(buf[32-len(b):], b)
	return buf[:]
}
