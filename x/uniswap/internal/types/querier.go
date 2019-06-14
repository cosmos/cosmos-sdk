package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Query endpoints supported by the uniswap querier
	QueryBalance    = "balance"
	QueryLiquidity  = "liquidity"
	QueryParameters = "parameters"

	ParamFee         = "fee"
	ParamNativeAsset = "nativeAsset"
)

// QueryBalanceParams defines the params for the query:
// - 'custom/uniswap/balance'
type QueryBalanceParams struct {
	Address sdk.AccAddress
}

// NewQueryBalanceParams is a constructor function for QueryBalanceParams
func NewQueryBalanceParams(address sdk.AccAddress) QueryBalanceParams {
	return QueryBalanceParams{
		Address: address,
	}
}

// QueryLiquidity defines the params for the query:
// - 'custom/uniswap/liquidity'
type QueryLiquidityParams struct {
	Denom string
}

// NewQueryLiquidityParams is a constructor function for QueryLiquidityParams
func NewQueryLiquidityParams(denom string) QueryLiquidityParams {
	return QueryLiquidityParams{
		Denom: denom,
	}
}
