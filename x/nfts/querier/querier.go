package querier

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nfts/keeper"
	"github.com/cosmos/cosmos-sdk/x/nfts/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the NFT Querier
const (
	QuerySupply     = "supply"
	QueryBalance    = "balance"
	QueryCollection = "collection"
	QueryNFT        = "nft"
)

// QueryColectionParams defines the params for queries:
// - 'custom/nfts/supply'
// - 'custom/nfts/collection'
type QueryColectionParams struct {
	Denom types.Denom
}

// NewQueryColectionParams creates a new instance of QuerySupplyParams
func NewQueryColectionParams(denom string) QueryColectionParams {
	return QueryColectionParams{Denom: types.Denom(denom)}
}

// QueryBalanceParams params for query 'custom/nfts/balance'
type QueryBalanceParams struct {
	Denom types.Denom
	Owner sdk.AccAddress
}

// NewQueryBalanceParams creates a new instance of QuerySupplyParams
func NewQueryBalanceParams(denom string, owner sdk.AccAddress) QueryBalanceParams {
	return QueryBalanceParams{
		Denom: types.Denom(denom),
		Owner: owner,
	}
}

// QueryNFTParams params for query 'custom/nfts/nft'
type QueryNFTParams struct {
	Denom   types.Denom
	TokenID types.TokenID
}

// NewQueryNFTParams creates a new instance of QueryNFTParams
func NewQueryNFTParams(denom string, id uint) QueryNFTParams {
	return QueryNFTParams{
		Denom:   types.Denom(denom),
		TokenID: types.TokenID(id),
	}
}

// NewQuerier is the module level router for state queries
func NewQuerier(k keeper.Keeper, cdc *codec.Codec) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QuerySupply:
			return querySupply(ctx, cdc, path[1:], req, k)
		case QueryBalance:
			return queryBalance(ctx, cdc, path[1:], req, k)
		case QueryCollection:
			return queryCollection(ctx, cdc, path[1:], req, k)
		case QueryNFT:
			return queryNFT(ctx, cdc, path[1:], req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown nft query endpoint")
		}
	}
}

func querySupply(ctx sdk.Context, cdc *codec.Codec, path []string, req abci.RequestQuery, k keeper.Keeper) ([]byte, sdk.Error) {

	var params QueryColectionParams
	err := cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	collection, found := k.GetCollection(ctx, params.Denom)
	if !found {
		return nil, types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("unknown denom %s", params.Denom))
	}

	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, int64(len(collection.Supply)))
	return bz, nil
}

func queryBalance(ctx sdk.Context, cdc *codec.Codec, path []string, req abci.RequestQuery, k keeper.Keeper) ([]byte, sdk.Error) {

	var params QueryBalanceParams
	err := cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	// TODO: get collections of NFTs held by address

	if params.Denom == "" {
		// TODO: return array of collections
		// collections := k.GetOwnerCollections(ctx, params.Owner)
		// bz, err := cdc.MarshalJSONIndent(cdc, collections)
		// if err != nil {
		// 	return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
		// }
		// return bz, nil
	}
	// TODO: return a single collection

	collection, found := k.GetCollection(ctx, params.Denom)
	if !found {
		return nil, types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("unknown denom %s", params.Denom))
	}

	bz, err := cdc.MarshalJSONIndent(cdc, collection)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

func queryCollection(ctx sdk.Context, cdc *codec.Codec, path []string, req abci.RequestQuery, k keeper.Keeper) ([]byte, sdk.Error) {

	var params QueryColectionParams
	err := cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	collection, found := k.GetCollection(ctx, params.Denom)
	if !found {
		return nil, types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("unknown denom %s", params.Denom))
	}

	bz, err := cdc.MarshalJSONIndent(cdc, collection)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

func queryNFT(ctx sdk.Context, cdc *codec.Codec, path []string, req abci.RequestQuery, k keeper.Keeper) ([]byte, sdk.Error) {

	var params QueryNFTParams
	err := cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	nft, err := k.GetNFT(ctx, params.Denom, params.TokenID)
	if err != nil {
		return nil, types.ErrUnknownNFT(types.DefaultCodespace, fmt.Sprintf("invalid NFT #%d from collection %s", params.TokenID, params.Denom))
	}

	bz, err := codec.MarshalJSONIndent(cdc, nft)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}
