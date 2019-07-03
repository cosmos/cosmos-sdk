package types

// query endpoints supported by the supply Querier
const (
	QueryTotalSupply = "total_supply"
	QuerySupplyOf    = "supply_of"
)

// QueryTotalSupply defines the params for the following queries:
//
// - 'custom/supply/totalSupply'
type QueryTotalSupplyParams struct {
	Page, Limit int
}

// NewQueryTotalSupplyParams creates a new instance to query the total supply
func NewQueryTotalSupplyParams(page, limit int) QueryTotalSupplyParams {
	return QueryTotalSupplyParams{page, limit}
}

// QuerySupplyOfParams defines the params for the following queries:
//
// - 'custom/supply/totalSupplyOf'
type QuerySupplyOfParams struct {
	Denom string
}

// NewQuerySupplyOfParams creates a new instance to query the total supply
// of a given denomination
func NewQuerySupplyOfParams(denom string) QuerySupplyOfParams {
	return QuerySupplyOfParams{denom}
}
