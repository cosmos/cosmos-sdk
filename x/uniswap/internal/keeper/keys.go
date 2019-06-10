package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// keys
var (
	ExchangePrefix    = []byte{0x00} // prefix for exchange liquidity keys
	UNIBalancesPrefix = []byte{0x01} // prefix for UNI balances keys
	TotalUNIKey       = []byte{0x02} // key for total
)

// GetExchangeKey gets the key for an exchanges total liquidity
func GetExchangeKey(denom string) []byte {
	return append(ExchangePrefix, []byte(denom)...)
}

// GetUNIBalancesKey gets the key for an addresses UNI balance
func GetUNIBalancesKey(addr sdk.AccAddress) []byte {
	return append(UNIBalancesPrefix, addr.Bytes()...)
}
