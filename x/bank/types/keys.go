package types

import (
	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName defines the module name
	ModuleName = "bank"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

// KVStore keys
var (
	SupplyKey           = collections.NewPrefix(0)
	DenomMetadataPrefix = collections.NewPrefix(1)
	// BalancesPrefix is the prefix for the account balances store. We use a byte
	// (instead of `[]byte("balances")` to save some disk space).
	BalancesPrefix     = collections.NewPrefix(2)
	DenomAddressPrefix = collections.NewPrefix(3)
	// SendEnabledPrefix is the prefix for the SendDisabled flags for a Denom.
	SendEnabledPrefix = collections.NewPrefix(4)

	// ParamsKey is the prefix for x/bank parameters
	ParamsKey = collections.NewPrefix(5)
)

// CreatePrefixedAccountStoreKey returns the key for the given account and denomination.
// This method can be used when performing an ABCI query for the balance of an account.
func CreatePrefixedAccountStoreKey(addr []byte, denom []byte) []byte {
	return append(CreateAccountBalancesPrefix(addr), denom...)
}

// CreateAccountBalancesPrefix creates the prefix for an account's balances.
func CreateAccountBalancesPrefix(addr []byte) []byte {
	return append(BalancesPrefix, address.MustLengthPrefix(addr)...)
}

// CreateDenomAddressPrefix creates a prefix for a reverse index of denomination
// to account balance for that denomination.
func CreateDenomAddressPrefix(denom string) []byte {
	// we add a "zero" byte at the end - null byte terminator, to allow prefix denom prefix
	// scan. Setting it is not needed (key[last] = 0) - because this is the default.
	key := make([]byte, len(DenomAddressPrefix)+len(denom)+1)
	copy(key, DenomAddressPrefix)
	copy(key[len(DenomAddressPrefix):], denom)
	return key
}
