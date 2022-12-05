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

// Addressable represents any type from which we can derive an address.
type Addressable interface {
	Address() []byte
}

// Hash creates a new address from address type and key.
// The functions should only be used by new types defining their own address function
// (eg public keys).
func Hash(typ string, key []byte) []byte {
	hasher := sha256.New()
	_, err := hasher.Write(conv.UnsafeStrToBytes(typ))
	// the error always nil, it's here only to satisfy the io.Writer interface
	errors.AssertNil(err)
	th := hasher.Sum(nil)

	hasher.Reset()
	_, err = hasher.Write(th)
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
// is constructed from a module name and a sequence of derivation keys (at least one
// derivation key must be provided).
func Module(moduleName string, derivationKey []byte, derivationKeys ...[]byte) []byte {
	mKey := append([]byte(moduleName), 0)
	addr := Hash("module", append(mKey, derivationKey...))
	for _, k := range derivationKeys {
		addr = Derive(addr, k)
	}
	return addr
}

// Derive derives a new address from the main `address` and a derivation `key`.
func Derive(address []byte, key []byte) []byte {
	return Hash(conv.UnsafeBytesToStr(address), key)
}
