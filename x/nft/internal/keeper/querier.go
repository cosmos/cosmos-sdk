package keeper

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the NFT Querier
const (
	QuerySupply       = "supply"
	QueryOwner        = "owner"
	QueryOwnerByDenom = "ownerByDenom"
	QueryCollection   = "collection"
	QueryDenoms       = "denoms"
	QueryNFT          = "nft"
)

// NewQuerier is the module level router for state queries
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QuerySupply:
			return querySupply(ctx, path[1:], req, k)
		case QueryOwner:
			return queryOwner(ctx, path[1:], req, k)
		case QueryOwnerByDenom:
			return queryOwnerByDenom(ctx, path[1:], req, k)
		case QueryCollection:
			return queryCollection(ctx, path[1:], req, k)
		case QueryDenoms:
			return queryDenoms(ctx, path[1:], req, k)
		case QueryNFT:
			return queryNFT(ctx, path[1:], req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown nft query endpoint")
		}
	}
}

func querySupply(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryCollectionParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	collection, found := k.GetCollection(ctx, params.Denom)
	if !found {
		return nil, types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("unknown denom %s", params.Denom))
	}

	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, uint64(collection.Supply()))
	return bz, nil
}

func queryOwner(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryBalanceParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	owner := k.GetOwner(ctx, params.Owner)
	bz, err := types.ModuleCdc.MarshalJSON(owner)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return bz, nil
}

func queryOwnerByDenom(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryBalanceParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	var owner types.Owner

	idCollection, _ := k.GetOwnerByDenom(ctx, params.Owner, params.Denom)
	owner.Address = params.Owner
	owner.IDCollections = append(owner.IDCollections, idCollection).Sort()

	bz, err := types.ModuleCdc.MarshalJSON(owner)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return bz, nil
}

func queryCollection(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryCollectionParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	collection, found := k.GetCollection(ctx, params.Denom)
	if !found {
		return nil, types.ErrUnknownCollection(types.DefaultCodespace, fmt.Sprintf("unknown denom %s", params.Denom))
	}

	// use Collections custom JSON to make the denom the key of the object
	collections := types.NewCollections(collection)
	bz, err := types.ModuleCdc.MarshalJSON(collections)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return bz, nil
}

func queryDenoms(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	denoms := k.GetDenoms(ctx)

	bz, err := types.ModuleCdc.MarshalJSON(denoms)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return bz, nil
}

func queryNFT(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryNFTParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrUnknownRequest(sdk.AppendMsgToErr("incorrectly formatted request data", err.Error()))
	}

	nft, err := k.GetNFT(ctx, params.Denom, params.TokenID)
	if err != nil {
		return nil, types.ErrUnknownNFT(types.DefaultCodespace, fmt.Sprintf("invalid NFT #%s from collection %s", params.TokenID, params.Denom))
	}

	bz, err := types.ModuleCdc.MarshalJSON(nft)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return bz, nil
}
