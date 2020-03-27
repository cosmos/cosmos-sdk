package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier path constants
const (
	QueryBalance     = "balance"
	QueryAllBalances = "all_balances"
)

// NewQueryBalanceParams creates a new instance of QueryBalanceParams.
func NewQueryBalanceParams(addr sdk.AccAddress, denom string) QueryBalanceParams {
	return QueryBalanceParams{Address: addr, Denom: denom}
}

// NewQueryAllBalancesParams creates a new instance of QueryAllBalancesParams.
func NewQueryAllBalancesParams(addr sdk.AccAddress) QueryAllBalancesParams {
	return QueryAllBalancesParams{Address: addr}
}
