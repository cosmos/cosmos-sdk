package types

// DONTCOVER

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryCollectionParams defines the params for queries:
// - 'custom/nft/supply'
// - 'custom/nft/collection'
type QueryCollectionParams struct {
	Denom string
}

// NewQueryCollectionParams creates a new instance of QuerySupplyParams
func NewQueryCollectionParams(denom string) QueryCollectionParams {
	return QueryCollectionParams{Denom: denom}
}

// Bytes exports the Denom as bytes
func (q QueryCollectionParams) Bytes() []byte {
	return []byte(q.Denom)
}

// QueryBalanceParams params for query 'custom/nfts/balance'
type QueryBalanceParams struct {
	Owner sdk.AccAddress
	Denom string // optional
}

// NewQueryBalanceParams creates a new instance of QuerySupplyParams
func NewQueryBalanceParams(owner sdk.AccAddress, denom ...string) QueryBalanceParams {
	if len(denom) > 0 {
		return QueryBalanceParams{
			Owner: owner,
			Denom: denom[0],
		}
	}
	return QueryBalanceParams{Owner: owner}
}

// QueryNFTParams params for query 'custom/nfts/nft'
type QueryNFTParams struct {
	Denom   string
	TokenID string
}

// NewQueryNFTParams creates a new instance of QueryNFTParams
func NewQueryNFTParams(denom, id string) QueryNFTParams {
	return QueryNFTParams{
		Denom:   denom,
		TokenID: id,
	}
}
