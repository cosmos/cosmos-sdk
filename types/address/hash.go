package address

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"
	"sort"
	"unsafe"

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
	hasher.Write(unsafeStrToByteArray(typ))
	th := hasher.Sum(nil)

	hasher.Reset()
	_, err := hasher.Write(th)
	// the error always nil, it's here only to satisfy the io.Writer interface
	errors.AssertNil(err)
	_, err = hasher.Write(key)
	errors.AssertNil(err)
	return hasher.Sum(nil)
}

// NewComposed creates a new address based on sub addresses.
func NewComposed(typ string, subAddresses []Addressable) ([]byte, error) {
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

// unsafeStrToByteArray uses unsafe to convert string into byte array. Returned array
// cannot be altered after this functions is called
func unsafeStrToByteArray(s string) []byte {
	sh := *(*reflect.SliceHeader)(unsafe.Pointer(&s))
	sh.Cap = sh.Len
	bs := *(*[]byte)(unsafe.Pointer(&sh))
	return bs
}
