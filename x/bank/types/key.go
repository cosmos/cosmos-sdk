package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
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

// KVStore keys
var (
	SupplyKey           = []byte{0x00}
	DenomMetadataPrefix = []byte{0x1}
	DenomAddressPrefix  = []byte{0x03}

	// BalancesPrefix is the prefix for the account balances store. We use a byte
	// (instead of `[]byte("balances")` to save some disk space).
	BalancesPrefix = []byte{0x02}
)

// AddressAndDenomFromBalancesStore returns an account address and denom from a balances prefix
// store. The key must not contain the prefix BalancesPrefix as the prefix store
// iterator discards the actual prefix.
//
// If invalid key is passed, AddressAndDenomFromBalancesStore returns ErrInvalidKey.
func AddressAndDenomFromBalancesStore(key []byte) (sdk.AccAddress, string, error) {
	if len(key) == 0 {
		return nil, "", ErrInvalidKey
	}

	kv.AssertKeyAtLeastLength(key, 1)

	addrBound := int(key[0])

	if len(key)-1 < addrBound {
		return nil, "", ErrInvalidKey
	}

	return key[1 : addrBound+1], string(key[addrBound+1:]), nil
}

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
