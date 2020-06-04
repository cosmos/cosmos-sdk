package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier path constants
const (
	QueryBalance     = "balance"
	QueryAllBalances = "all_balances"
	QueryTotalSupply = "total_supply"
	QuerySupplyOf    = "supply_of"
)

// NewQueryBalanceRequest creates a new instance of QueryBalanceRequest.
func NewQueryBalanceRequest(addr sdk.AccAddress, denom string) *QueryBalanceRequest {
	return &QueryBalanceRequest{Address: addr, Denom: denom}
}

// NewQueryAllBalancesRequest creates a new instance of QueryAllBalancesRequest.
func NewQueryAllBalancesRequest(addr sdk.AccAddress) *QueryAllBalancesRequest {
	return &QueryAllBalancesRequest{Address: addr}
}

// QueryTotalSupply defines the params for the following queries:
//
// - 'custom/bank/totalSupply'
type QueryTotalSupplyParams struct {
	Page, Limit int
}

// NewQueryTotalSupplyParams creates a new instance to query the total supply
func NewQueryTotalSupplyParams(page, limit int) QueryTotalSupplyParams {
	return QueryTotalSupplyParams{page, limit}
}

// QuerySupplyOfParams defines the params for the following queries:
//
// - 'custom/bank/totalSupplyOf'
type QuerySupplyOfParams struct {
	Denom string
}

// NewQuerySupplyOfParams creates a new instance to query the total supply
// of a given denomination
func NewQuerySupplyOfParams(denom string) QuerySupplyOfParams {
	return QuerySupplyOfParams{denom}
}
