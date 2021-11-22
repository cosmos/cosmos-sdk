package group

import (
	"crypto/sha256"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// AddressLength is the length of all addresses
	// You can modify it in init() before any addresses are calculated,
	// but it must not change during the lifetime of the kvstore
	AddressLength = 20
)

// Condition is a specially formatted array, containing
// information on who can authorize an action.
// It is of the format:
//
//   sprintf("%s/%s/%s", extension, type, data)
type Condition []byte

func NewCondition(ext, typ string, data []byte) Condition {
	pre := fmt.Sprintf("%s/%s/", ext, typ)
	return append([]byte(pre), data...)
}

// Address will convert a Condition into an Address
func (c Condition) Address() sdk.AccAddress {
	return newAddress(c)
}

// newAddress hashes and truncates into the proper size
func newAddress(data []byte) sdk.AccAddress {
	if data == nil {
		return nil
	}
	// h := blake2b.Sum256(data)
	h := sha256.Sum256(data)
	return h[:AddressLength]
}
