package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier path constants
const (
	QueryBalance     = "balance"
	QueryAllBalances = "all_balances"
)

// QueryBalanceParams defines the params for querying an account balance.
type QueryBalanceParams struct {
	Address sdk.AccAddress
	Denom   string
}

// NewQueryBalanceParams creates a new instance of QueryBalanceParams.
func NewQueryBalanceParams(addr sdk.AccAddress, denom string) QueryBalanceParams {
	return QueryBalanceParams{Address: addr, Denom: denom}
}

// QueryAllBalancesParams defines the params for querying all account balances
type QueryAllBalancesParams struct {
	Address sdk.AccAddress
}

// NewQueryAllBalancesParams creates a new instance of QueryAllBalancesParams.
func NewQueryAllBalancesParams(addr sdk.AccAddress) QueryAllBalancesParams {
	return QueryAllBalancesParams{Address: addr}
}
