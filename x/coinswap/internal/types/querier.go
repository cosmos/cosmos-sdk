package types

const (
	// Query endpoints supported by the coinswap querier
	QueryLiquidity  = "liquidity"
	QueryParameters = "parameters"

	ParamFee         = "fee"
	ParamNativeDenom = "nativeDenom"
)

// defines the params for the following queries:
// - 'custom/coinswap/liquidity'
type QueryLiquidityParams struct {
	NonNativeDenom string
}

func NewQueryLiquidityParams(nonNativeDenom string) QueryLiquidityParams {
	return QueryLiquidityParams{
		NonNativeDenom: strings.TrimSpace(nonNativeDenom),
	}
}
