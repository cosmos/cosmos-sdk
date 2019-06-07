package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// keys
var (
	ExchangePrefix    = []byte{0x00} // key for exchange liquidity
	UNIBalancesPrefix = []byte{0x01} // key for UNI balances
)

const (
	// ModuleName is the name of the module
	ModuleName = "uniswap"
)

// GetExchangeKey gets the key for an exchanges total liquidity
func GetExchangeKey(denom string) []byte {
	return append(ExchangePrefix, []byte(denom)...)
}

// GetUNIBalancesKey gets the key for an addresses UNI balance
func GetUNIBalancesKey(addr sdk.AccAddress) []byte {
	return append(UNIBalancesPrefix, addr.Bytes()...)
}
