package types

import (
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
	return sdk.AccAddress(key[len(BalancesPrefix):])
}
