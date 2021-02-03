package v042

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// KVStore keys
var (
	// BalancesPrefix is the for the account balances store. We use a byte
	// (instead of say `[]]byte("balances")` to save some disk space).
	BalancesPrefix = []byte{0x02}
)

// AddressFromBalancesStore returns an account address from a balances prefix
// store. The key must not contain the perfix BalancesPrefix as the prefix store
// iterator discards the actual prefix.
func AddressFromBalancesStore(key []byte) sdk.AccAddress {
	addrLen := key[0]
	addr := key[1 : addrLen+1]

	return sdk.AccAddress(addr)
}

// CreateAccountBalancesPrefix creates the prefix for an account's balances.
func CreateAccountBalancesPrefix(addr []byte) []byte {
	return append(BalancesPrefix, address.MustLengthPrefix(addr)...)
}
