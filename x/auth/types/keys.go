package types

import (
	"crypto/sha256"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is "auth"
	ModuleName = "auth"

	// StoreKey is string representation of the store key for auth
	StoreKey = "acc"

	// FeeCollectorName the root string for the fee collector account address
	FeeCollectorName = "fee_collector"

	// QuerierRoute is the querier route for auth
	QuerierRoute = ModuleName
)

var (
	// AddressStoreKeyPrefix prefix for account-by-address store
	AddressStoreKeyPrefix = []byte{0x01}

	// GlobalAccountNumberKey param key for global account number
	GlobalAccountNumberKey = []byte("globalAccountNumber")
)

// AddressStoreKey turn an address to key used to get it from the account store
func AddressStoreKey(accountType string, addr sdk.AccAddress) []byte {
	typeHash := sha256.Sum256([]byte(accountType))
	key := append(AddressStoreKeyPrefix, typeHash[:]...)
	return append(key, addr.Bytes()...)
}
