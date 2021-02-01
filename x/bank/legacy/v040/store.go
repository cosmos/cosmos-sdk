package v040

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
)

// KVStore keys
var (
	BalancesPrefix = []byte("balances")
)

// AddressFromBalancesStore returns an account address from a balances prefix
// store. The key must not contain the perfix BalancesPrefix as the prefix store
// iterator discards the actual prefix.
func AddressFromBalancesStore(key []byte) sdk.AccAddress {
	addr := key[:v040auth.AddrLen]
	if len(addr) != v040auth.AddrLen {
		panic(fmt.Sprintf("unexpected account address key length; got: %d, expected: %d", len(addr), v040auth.AddrLen))
	}

	return sdk.AccAddress(addr)
}
