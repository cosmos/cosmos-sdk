package types

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// module name
	ModuleName = "auth"

	// StoreKey is string representation of the store key for auth
	StoreKey = "acc"

	// FeeCollectorName the root string for the fee collector account address
	FeeCollectorName = "fee_collector"
)

var (
	// ParamsKey is the prefix for params key
	ParamsKey = collections.NewPrefix(0)

	// AddressStoreKeyPrefix prefix for account-by-address store
	AddressStoreKeyPrefix = []byte{0x01}

	// GlobalAccountNumberKey identifies the prefix where the monotonically increasing
	// account number is stored.
	GlobalAccountNumberKey = collections.NewPrefix(2)

	// AccountNumberStoreKeyPrefix prefix for account-by-id store
	AccountNumberStoreKeyPrefix = []byte("accountNumber")
)

// AddressStoreKey turn an address to key used to get it from the account store
func AddressStoreKey(addr sdk.AccAddress) []byte {
	return append(AddressStoreKeyPrefix, addr.Bytes()...)
}

// AccountNumberStoreKey turn an account number to key used to get the account address from account store
func AccountNumberStoreKey(accountNumber uint64) []byte {
	return append(AccountNumberStoreKeyPrefix, sdk.Uint64ToBigEndian(accountNumber)...)
}
