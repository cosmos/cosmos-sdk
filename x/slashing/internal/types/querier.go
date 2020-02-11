package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DONTCOVER

// Query endpoints supported by the slashing querier
const (
	QueryParameters   = "parameters"
	QuerySigningInfo  = "signingInfo"
	QuerySigningInfos = "signingInfos"
)

// NewQuerySigningInfoParams creates a new QuerySigningInfoParams instance
func NewQuerySigningInfoParams(consAddr sdk.ConsAddress) QuerySigningInfoParams {
	return QuerySigningInfoParams{consAddr}
}

// NewQuerySigningInfosParams creates a new QuerySigningInfosParams instance
func NewQuerySigningInfosParams(page, limit int32) QuerySigningInfosParams {
	return QuerySigningInfosParams{page, limit}
}
