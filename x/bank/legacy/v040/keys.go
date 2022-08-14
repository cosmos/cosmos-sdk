// Package v040 is copy-pasted from:
// https://github.com/cosmos/cosmos-sdk/blob/v0.41.0/x/bank/types/key.go
package v040

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
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
	BalancesPrefix      = []byte("balances")
	SupplyKey           = []byte{0x00}
	DenomMetadataPrefix = []byte{0x1}
)

// DenomMetadataKey returns the denomination metadata key.
func DenomMetadataKey(denom string) []byte {
	d := []byte(denom)
	return append(DenomMetadataPrefix, d...)
}

// AddressFromBalancesStore returns an account address from a balances prefix
// store. The key must not contain the perfix BalancesPrefix as the prefix store
// iterator discards the actual prefix.
func AddressFromBalancesStore(key []byte) sdk.AccAddress {
	kv.AssertKeyAtLeastLength(key, 1+v040auth.AddrLen)
	addr := key[:v040auth.AddrLen]
	kv.AssertKeyLength(addr, v040auth.AddrLen)
	return sdk.AccAddress(addr)
}
