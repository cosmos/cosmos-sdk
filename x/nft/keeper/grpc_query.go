package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

var _ nft.QueryServer = Keeper{}

func (k Keeper) Balance(goCtx context.Context, request *nft.QueryBalanceRequest) (*nft.QueryBalanceResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(request.ClassId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "class id can not be empty")
	}

	owner, err := sdk.AccAddressFromBech32(request.Owner)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	balance := k.GetBalance(ctx, request.ClassId, owner)
	return &nft.QueryBalanceResponse{Amount: balance}, nil
}

func (k Keeper) Owner(goCtx context.Context, request *nft.QueryOwnerRequest) (*nft.QueryOwnerResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(request.ClassId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "class id can not be empty")
	}
	if len(request.Id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "nft id can not be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	owner := k.GetOwner(ctx, request.ClassId, request.Id)
	return &nft.QueryOwnerResponse{Owner: owner.String()}, nil
}

func (k Keeper) Supply(goCtx context.Context, request *nft.QuerySupplyRequest) (*nft.QuerySupplyResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(request.ClassId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "class id can not be empty")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	supply := k.GetTotalSupply(ctx, request.ClassId)
	return &nft.QuerySupplyResponse{Amount: supply}, nil
}

func (k Keeper) NFTsOfClass(goCtx context.Context, request *nft.QueryNFTsOfClassRequest) (*nft.QueryNFTsOfClassResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(request.ClassId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "class id can not be empty")
	}

	var nfts []*nft.NFT
	ctx := sdk.UnwrapSDKContext(goCtx)

	// if owner is not empty, filter nft by owner
	if len(request.Owner) > 0 {
		owner, err := sdk.AccAddressFromBech32(request.Owner)
		if err != nil {
			return nil, err
		}

		ownerStore := k.getClassStoreByOwner(ctx, owner, request.ClassId)
		pageRes, err := query.Paginate(ownerStore, request.Pagination, func(key []byte, _ []byte) error {
			nft, has := k.GetNFT(ctx, request.ClassId, string(key))
			if has {
				nfts = append(nfts, &nft)
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
		return &nft.QueryNFTsOfClassResponse{
			Nfts:       nfts,
			Pagination: pageRes,
		}, nil
	}

	nftStore := k.getNFTStore(ctx, request.ClassId)
	pageRes, err := query.Paginate(nftStore, request.Pagination, func(_ []byte, value []byte) error {
		var nft nft.NFT
		k.cdc.MustUnmarshal(value, &nft)
		nfts = append(nfts, &nft)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &nft.QueryNFTsOfClassResponse{
		Nfts:       nfts,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) NFT(goCtx context.Context, request *nft.QueryNFTRequest) (*nft.QueryNFTResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(request.ClassId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "class id can not be empty")
	}
	if len(request.Id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "nft id can not be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	n, has := k.GetNFT(ctx, request.ClassId, request.Id)
	if !has {
		return nil, status.Errorf(codes.NotFound,
			"not found nft: class: %s, id: %s", request.ClassId, request.Id)
	}
	return &nft.QueryNFTResponse{Nft: &n}, nil

}

func (k Keeper) Class(goCtx context.Context, request *nft.QueryClassRequest) (*nft.QueryClassResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if len(request.ClassId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "class id can not be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	class, has := k.GetClass(ctx, request.ClassId)
	if !has {
		return nil, status.Errorf(codes.NotFound,
			"not found class: %s", request.ClassId)
	}
	return &nft.QueryClassResponse{Class: &class}, nil
}

func (k Keeper) Classes(goCtx context.Context, request *nft.QueryClassesRequest) (*nft.QueryClassesResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	classStore := prefix.NewStore(store, ClassKey)

	var classes []*nft.Class
	pageRes, err := query.Paginate(classStore, request.Pagination, func(_ []byte, value []byte) error {
		var class nft.Class
		k.cdc.MustUnmarshal(value, &class)
		classes = append(classes, &class)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &nft.QueryClassesResponse{
		Classes:    classes,
		Pagination: pageRes,
	}, nil
}
