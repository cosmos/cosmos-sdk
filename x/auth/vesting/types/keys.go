package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const (
	// StoreKey is string representation of the store key for vesting
	StoreKey = ModuleName
)

var (
	VestingAccountStoreKeyPrefix = []byte{0x01}
)

// VestingAccountStoreKey turn an address to key used to record it in the vesting store
func VestingAccountStoreKey(addr sdk.AccAddress) []byte {
	return append(VestingAccountStoreKeyPrefix, addr.Bytes()...)
}

// AddressFromVestingAccountKey creates the address from VestingAccountKey
func AddressFromVestingAccountKey(key []byte) sdk.AccAddress {
	kv.AssertKeyAtLeastLength(key, 2)
	return key[1:] // remove prefix byte
}
