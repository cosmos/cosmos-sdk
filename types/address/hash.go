package address

import (
	"crypto/sha256"
	"reflect"
	"sort"
	"unsafe"

	"github.com/cosmos/cosmos-sdk/types/errors"
)

/*
   TODO

   I still need to think how to organize it.
   Ideally, I wanted to Addressable Account abstraction
   - so all kinds of accounts which could be addressable (base, multisig, module...)

   Other idea is to leave away this abstraction, and only implement the related functions,
   which would take more
*/

// BaseLen is the length of generated addresses constructed by BaseAddress.
const BaseLen = sha256.Size

type Addressable interface {
	Address() []byte
}

func MkBase(typ string, key []byte) []byte {
	hasher := sha256.New()
	hasher.Write(unsafeStrToByteArray(typ))
	th := hasher.Sum(nil)

	hasher.Reset()
	_, err := hasher.Write(th)
	errors.Panic(err)
	_, err = hasher.Write(key)
	errors.Panic(err)
	return hasher.Sum(nil)
}

func MkComposed(typ string, subAddresses []Addressable) []byte {
	as := make([][]byte, len(subAddresses))
	totalLen := 0
	for i := range subAddresses {
		as[i] = subAddresses[i].Address()
		totalLen += len(as[i])
	}

	sort.Slice(as, func(i, j int) { byte.Compare(as[i], as[j]) <= 0 })
	key := make(key, totalLen)
	offset := 0
	for i := range as {
		key = copy(key[offset:], as[i])
		offset += len(as[i])
	}
	return MkBase(typ, key)
}

// unsafeStrToByteArray uses unsafe to convert string into byte array. Returned array
// cannot be altered after this functions is called
func unsafeStrToByteArray(s string) []byte {
	sh := *(*reflect.SliceHeader)(unsafe.Pointer(&s))
	sh.Cap = sh.Len
	bs := *(*[]byte)(unsafe.Pointer(&sh))
	return bs
}
