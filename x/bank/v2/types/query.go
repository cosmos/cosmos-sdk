package types

// NewQueryBalanceRequest creates a new instance of QueryBalanceRequest.
func NewQueryBalanceRequest(addr, denom string) *QueryBalanceRequest {
	return &QueryBalanceRequest{Address: addr, Denom: denom}
}
