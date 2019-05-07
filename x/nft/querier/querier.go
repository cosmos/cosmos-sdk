package querier

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/types"

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
// - 'custom/nft/supply'
// - 'custom/nft/collection'
type QueryColectionParams struct {
	Denom string
}

// NewQueryColectionParams creates a new instance of QuerySupplyParams
func NewQueryColectionParams(denom string) QueryColectionParams {
	return QueryColectionParams{Denom: types.Denom(denom)}
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
	TokenID uint64
}

// NewQueryNFTParams creates a new instance of QueryNFTParams
func NewQueryNFTParams(denom string, ID uint64) QueryNFTParams {
	return QueryNFTParams{
		Denom:   denom,
		TokenID: ID,
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
	binary.LittleEndian.PutUint64(bz, int64(collection.Supply))
	return bz, nil
}

func queryBalance(ctx sdk.Context, cdc *codec.Codec, path []string, req abci.RequestQuery, k keeper.Keeper) ([]byte, sdk.Error) {

	var params QueryBalanceParams
	err := cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	var collections types.Collections
	if params.Denom == "" {
		collections, _ = k.GetBalance(ctx, params.Owner)
	} else {
		collection, _ := k.GetBalanceCollection(ctx, params.Owner, params.Denom)
		collections = types.NewCollections(collection)
	}

	bz, err := collections.MarshalJSON()
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

	collections = types.NewCollections(collection)

	bz, err := collections.MarshalJSON()
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

	nfts := types.NewNFTs(nft)

	bz, err := nfts.MarshalJSON()
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
