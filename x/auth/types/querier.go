package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// query endpoints supported by the auth Querier
const (
	QueryAccount = "account"
	QueryParams  = "params"
)

// NewQueryAccountRequest creates a new instance of QueryAccountRequest.
func NewQueryAccountRequest(addr sdk.AccAddress) *QueryAccountRequest {
	return &QueryAccountRequest{Address: addr}
}

// NewQueryParametersRequest creates a new instance of QueryParametersRequest.
func NewQueryParametersRequest() *QueryParametersRequest {
	return &QueryParametersRequest{}
}
