package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// keys
var (
	ReservePoolPrefix = []byte{0x00} // prefix for reserve pool keys
	UNIBalancesPrefix = []byte{0x01} // prefix for UNI balances keys
	TotalUNIKey       = []byte{0x02} // key for total
)

// GetReservePoolKey gets the key for a reserve pool's total liquidity
func GetReservePoolKey(denom string) []byte {
	return append(ReservePoolPrefix, []byte(denom)...)
}

// GetUNIBalancesKey gets the key for an addresses UNI balance
func GetUNIBalancesKey(addr sdk.AccAddress) []byte {
	return append(UNIBalancesPrefix, addr.Bytes()...)
}
