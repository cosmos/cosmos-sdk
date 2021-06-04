package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

var _ types.QueryServer = Keeper{}

// NFT queries NFT details based on id.
func (k Keeper) NFT(c context.Context, req *types.QueryNFTRequest) (*types.QueryNFTResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.Id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "nft id can not be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	nft, has := k.GetNFT(ctx, req.Id)
	if !has {
		return nil, status.Error(codes.InvalidArgument, "nft does not exist")
	}
	return &types.QueryNFTResponse{NFT: &nft}, nil
}

// NFTs queries all proposals based on the optional onwer
func (k Keeper) NFTs(c context.Context, req *types.QueryNFTsRequest) (*types.QueryNFTsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(store, types.NFTKey)
	onResult := onQueryAllNFTs
	if len(req.Owner) > 0 {
		owner, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		prefixStore = k.getOwnerStore(ctx, owner)
		onResult = onQueryNFTsByOwner
	}

	var nfts []*types.NFT
	pageRes, err := query.FilteredPaginate(prefixStore, req.Pagination,
		func(key []byte, value []byte, accumulate bool) (bool, error) {
			return onResult(ctx, k, nfts)(key, value, accumulate)
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryNFTsResponse{
		NFTs:       nfts,
		Pagination: pageRes,
	}, nil
}

func onQueryAllNFTs(_ sdk.Context, k Keeper, nfts []*types.NFT) func(key, value []byte, accumulate bool) (bool, error) {
	return func(key, value []byte, accumulate bool) (bool, error) {
		var nft types.NFT
		if err := k.cdc.UnmarshalBinaryBare(value, &nft); err != nil {
			return false, err
		}

		if accumulate {
			nfts = append(nfts, &nft)
		}
		return true, nil
	}
}

func onQueryNFTsByOwner(ctx sdk.Context, k Keeper, nfts []*types.NFT) func(key, value []byte, accumulate bool) (bool, error) {
	return func(key, value []byte, accumulate bool) (bool, error) {
		if nft, has := k.GetNFT(ctx, types.UnmarshalNFTID(value)); has && accumulate {
			nfts = append(nfts, &nft)
		}
		return true, nil
	}
}
