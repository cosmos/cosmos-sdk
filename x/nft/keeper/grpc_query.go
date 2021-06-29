package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) NFT(ctx context.Context,
	request *types.QueryNFTRequest) (*types.QueryNFTResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	nft, has := k.GetNFT(sdkCtx, request.Type, request.Id)
	if !has {
		return nil, status.Errorf(codes.InvalidArgument, "not found nft")
	}
	return &types.QueryNFTResponse{NFT: &nft}, nil
}

func (k Keeper) NFTs(ctx context.Context,
	request *types.QueryNFTsRequest) (*types.QueryNFTsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var nfts []*types.NFT
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if len(request.Owner) == 0 {
		store := sdkCtx.KVStore(k.storeKey)
		typeStore := prefix.NewStore(store, types.NFTKey)
		pageRes, err := query.Paginate(typeStore, request.Pagination, func(_, value []byte) error {
			var nft types.NFT
			err := k.cdc.Unmarshal(value, &nft)
			if err != nil {
				return err
			}
			nfts = append(nfts, &nft)
			return nil
		})
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
		}
		return &types.QueryNFTsResponse{
			NFTs:       nfts,
			Pagination: pageRes,
		}, nil
	}

	owner, err := sdk.AccAddressFromBech32(request.Owner)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	balances := k.bk.GetAllBalances(sdkCtx, owner)
	for _, b := range balances {
		typ, id, err := types.GetTypeAndIDFrom(b.GetDenom())
		if err != nil {
			continue
		}
		if nft, has := k.GetNFT(sdkCtx, typ, id); has {
			nfts = append(nfts, &nft)
		}
	}
	return &types.QueryNFTsResponse{
		NFTs: nfts,
	}, nil
}

func (k Keeper) NFTsOf(ctx context.Context,
	request *types.QueryNFTsOfRequest) (*types.QueryNFTsOfResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	nftStore := k.getNFTStore(sdkCtx, request.Type)

	var nfts []*types.NFT
	pageRes, err := query.Paginate(nftStore, request.Pagination, func(_, value []byte) error {
		var nft types.NFT
		err := k.cdc.Unmarshal(value, &nft)
		if err != nil {
			return err
		}
		nfts = append(nfts, &nft)
		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}
	return &types.QueryNFTsOfResponse{
		NFTs:       nfts,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) Supply(ctx context.Context, request *types.QuerySupplyRequest) (*types.QuerySupplyResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	nftStore := k.getNFTStore(sdkCtx, request.Type)

	iterator := nftStore.Iterator(nil, nil)
	defer iterator.Close()

	var supply uint64
	for ; iterator.Valid(); iterator.Next() {
		supply++
	}
	return &types.QuerySupplyResponse{Amount: supply}, nil
}

func (k Keeper) Balance(ctx context.Context, request *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	owner, err := sdk.AccAddressFromBech32(request.Owner)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	var balance uint64
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	balances := k.bk.GetAllBalances(sdkCtx, owner)
	for _, b := range balances {
		typ, id, err := types.GetTypeAndIDFrom(b.GetDenom())
		if err != nil {
			continue
		}
		if k.HasNFT(sdkCtx, typ, id) {
			balance++
		}
	}
	return &types.QueryBalanceResponse{Amount: balance}, nil
}

func (k Keeper) Type(ctx context.Context, request *types.QueryTypeRequest) (*types.QueryTypeResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	metadata, has := k.GetMetadata(sdkCtx, request.Type)
	if !has {
		return nil, status.Errorf(codes.InvalidArgument, "type not found: %s", request.Type)
	}
	return &types.QueryTypeResponse{Metadata: &metadata}, nil
}

func (k Keeper) Types(ctx context.Context, request *types.QueryTypesRequest) (*types.QueryTypesResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var metadatas []*types.Metadata

	store := sdkCtx.KVStore(k.storeKey)
	typeStore := prefix.NewStore(store, types.TypeKey)
	pageRes, err := query.Paginate(typeStore, request.Pagination, func(_, value []byte) error {
		var metadata types.Metadata
		err := k.cdc.Unmarshal(value, &metadata)
		if err != nil {
			return err
		}
		metadatas = append(metadatas, &metadata)
		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}
	return &types.QueryTypesResponse{
		Metadatas:  metadatas,
		Pagination: pageRes,
	}, nil
}
