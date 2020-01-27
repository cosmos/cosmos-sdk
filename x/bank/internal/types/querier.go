package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// QueryBalancesParams defines the params for querying all account balances
type QueryBalancesParams struct {
	Address sdk.AccAddress
}

// NewQueryBalancesParams creates a new instance of QueryBalanceParams.
func NewQueryBalancesParams(addr sdk.AccAddress) QueryBalancesParams {
	return QueryBalancesParams{Address: addr}
}
