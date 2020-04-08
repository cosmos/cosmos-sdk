package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier path constants
const (
	QueryBalance     = "balance"
	QueryAllBalances = "all_balances"
)

// NewQueryBalanceRequest creates a new instance of QueryBalanceRequest.
func NewQueryBalanceRequest(addr sdk.AccAddress, denom string) *QueryBalanceRequest {
	return &QueryBalanceRequest{Address: addr, Denom: denom}
}

// NewQueryAllBalancesRequest creates a new instance of QueryAllBalancesRequest.
func NewQueryAllBalancesRequest(addr sdk.AccAddress) *QueryAllBalancesRequest {
	return &QueryAllBalancesRequest{Address: addr}
}
