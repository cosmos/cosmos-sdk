package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "bank"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	BalancesPrefix = []byte("balances")
)

// AddressFromBalancesKey returns an account address from a balances key which
// is used as an index to store balances per account.
func AddressFromBalancesKey(key []byte) sdk.AccAddress {
	addr := key[len(BalancesPrefix) : len(BalancesPrefix)+sdk.AddrLen]
	if len(addr) != sdk.AddrLen {
		panic(fmt.Sprintf("unexpected account address key length; got: %d, expected: %d", len(addr), sdk.AddrLen))
	}

	return sdk.AccAddress(addr)
}
