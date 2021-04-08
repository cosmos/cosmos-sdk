package address

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/internal/conv"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

// Len is the length of base addresses
const Len = sha256.Size

type Addressable interface {
	Address() []byte
}

// Hash creates a new address from address type and key
func Hash(typ string, key []byte) []byte {
	hasher := sha256.New()
	hasher.Write(conv.UnsafeStrToBytes(typ))
	th := hasher.Sum(nil)

	hasher.Reset()
	_, err := hasher.Write(th)
	// the error always nil, it's here only to satisfy the io.Writer interface
	errors.AssertNil(err)
	_, err = hasher.Write(key)
	errors.AssertNil(err)
	return hasher.Sum(nil)
}

// Compose creates a new address based on sub addresses.
func Compose(typ string, subAddresses []Addressable) ([]byte, error) {
	as := make([][]byte, len(subAddresses))
	totalLen := 0
	var err error
	for i := range subAddresses {
		a := subAddresses[i].Address()
		as[i], err = LengthPrefix(a)
		if err != nil {
			return nil, fmt.Errorf("not compatible sub-adddress=%v at index=%d [%w]", a, i, err)
		}
		totalLen += len(as[i])
	}

	sort.Slice(as, func(i, j int) bool { return bytes.Compare(as[i], as[j]) <= 0 })
	key := make([]byte, totalLen)
	offset := 0
	for i := range as {
		copy(key[offset:], as[i])
		offset += len(as[i])
	}
	return Hash(typ, key), nil
}

// Module is a specialized version of a composed address for modules. Each module account
// is constructed from a module name and module account key.
func Module(moduleName string, key []byte) []byte {
	mKey := append([]byte(moduleName), 0)
	return Hash("module", append(mKey, key...))
}

// Derive derives a new address from the main `address` and a derivation `key`.
func Derive(address []byte, key []byte) []byte {
	return Hash(conv.UnsafeBytesToStr(address), key)
}

// DeriveMulti generalizes `Derive` function for multiple keys - `path`. The keys are order
// sensitive. Changing an order of the elements in the path will create a different key.
// NOTE: DeriveMulti(addr, [k1, k2]) != Derive(Derive(addr, k1), k2)
//                                   != Derive(Derive(addr, k2), k1)
func DeriveMulti(address []byte, path [][]byte) ([]byte, error) {
	key := []byte{}
	for i, p := range path {
		a, err := LengthPrefix(p)
		if err != nil {
			return nil, fmt.Errorf("a path key=%v at index=%d is not compatible [%w]", p, i, err)
		}
		key = append(key, a...)
	}
	return Hash(conv.UnsafeBytesToStr(address), key), nil
}
