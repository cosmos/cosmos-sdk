package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"

	nft2 "github.com/cosmos/cosmos-sdk/contrib/x/nft"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
)

var _ nft2.QueryServer = Keeper{}

// Balance return the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721
func (k Keeper) Balance(goCtx context.Context, r *nft2.QueryBalanceRequest) (*nft2.QueryBalanceResponse, error) {
	if r == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}

	if len(r.ClassId) == 0 {
		return nil, nft2.ErrEmptyClassID
	}

	owner, err := k.ac.StringToBytes(r.Owner)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	balance := k.GetBalance(ctx, r.ClassId, owner)
	return &nft2.QueryBalanceResponse{Amount: balance}, nil
}

// Owner return the owner of the NFT based on its class and id, same as ownerOf in ERC721
func (k Keeper) Owner(goCtx context.Context, r *nft2.QueryOwnerRequest) (*nft2.QueryOwnerResponse, error) {
	if r == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}

	if len(r.ClassId) == 0 {
		return nil, nft2.ErrEmptyClassID
	}

	if len(r.Id) == 0 {
		return nil, nft2.ErrEmptyNFTID
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	owner := k.GetOwner(ctx, r.ClassId, r.Id)
	if owner.Empty() {
		return &nft2.QueryOwnerResponse{Owner: ""}, nil
	}
	ownerstr, err := k.ac.BytesToString(owner.Bytes())
	if err != nil {
		return nil, err
	}
	return &nft2.QueryOwnerResponse{Owner: ownerstr}, nil
}

// Supply return the number of NFTs from the given class, same as totalSupply of ERC721.
func (k Keeper) Supply(goCtx context.Context, r *nft2.QuerySupplyRequest) (*nft2.QuerySupplyResponse, error) {
	if r == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}

	if len(r.ClassId) == 0 {
		return nil, nft2.ErrEmptyClassID
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	supply := k.GetTotalSupply(ctx, r.ClassId)
	return &nft2.QuerySupplyResponse{Amount: supply}, nil
}

// NFTs queries all NFTs of a given class or owner (at least one must be provided), similar to tokenByIndex in ERC721Enumerable
func (k Keeper) NFTs(goCtx context.Context, r *nft2.QueryNFTsRequest) (*nft2.QueryNFTsResponse, error) {
	if r == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}

	var err error
	var owner sdk.AccAddress

	if len(r.Owner) > 0 {
		owner, err = k.ac.StringToBytes(r.Owner)
		if err != nil {
			return nil, err
		}
	}

	var nfts []*nft2.NFT
	var pageRes *query.PageResponse
	ctx := sdk.UnwrapSDKContext(goCtx)

	switch {
	case len(r.ClassId) > 0 && len(r.Owner) > 0:
		if pageRes, err = query.Paginate(k.getClassStoreByOwner(ctx, owner, r.ClassId), r.Pagination, func(key, _ []byte) error {
			nft, has := k.GetNFT(ctx, r.ClassId, string(key))
			if has {
				nfts = append(nfts, &nft)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	case len(r.ClassId) > 0 && len(r.Owner) == 0:
		nftStore := k.getNFTStore(ctx, r.ClassId)
		if pageRes, err = query.Paginate(nftStore, r.Pagination, func(_, value []byte) error {
			var nft nft2.NFT
			if err := k.cdc.Unmarshal(value, &nft); err != nil {
				return err
			}
			nfts = append(nfts, &nft)
			return nil
		}); err != nil {
			return nil, err
		}
	case len(r.ClassId) == 0 && len(r.Owner) > 0:
		if pageRes, err = query.Paginate(k.prefixStoreNftOfClassByOwner(ctx, owner), r.Pagination, func(key, value []byte) error {
			classID, nftID := parseNftOfClassByOwnerStoreKey(key)
			if n, has := k.GetNFT(ctx, classID, nftID); has {
				nfts = append(nfts, &n)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	default:
		return nil, sdkerrors.ErrInvalidRequest.Wrap("must provide at least one of classID or owner")
	}
	return &nft2.QueryNFTsResponse{
		Nfts:       nfts,
		Pagination: pageRes,
	}, nil
}

// NFT return an NFT based on its class and id.
func (k Keeper) NFT(goCtx context.Context, r *nft2.QueryNFTRequest) (*nft2.QueryNFTResponse, error) {
	if r == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}

	if len(r.ClassId) == 0 {
		return nil, nft2.ErrEmptyClassID
	}
	if len(r.Id) == 0 {
		return nil, nft2.ErrEmptyNFTID
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	n, has := k.GetNFT(ctx, r.ClassId, r.Id)
	if !has {
		return nil, nft2.ErrNFTNotExists.Wrapf("not found nft: class: %s, id: %s", r.ClassId, r.Id)
	}
	return &nft2.QueryNFTResponse{Nft: &n}, nil
}

// Class return an NFT class based on its id
func (k Keeper) Class(goCtx context.Context, r *nft2.QueryClassRequest) (*nft2.QueryClassResponse, error) {
	if r == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}

	if len(r.ClassId) == 0 {
		return nil, nft2.ErrEmptyClassID
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	class, has := k.GetClass(ctx, r.ClassId)
	if !has {
		return nil, nft2.ErrClassNotExists.Wrapf("not found class: %s", r.ClassId)
	}
	return &nft2.QueryClassResponse{Class: &class}, nil
}

// Classes return all NFT classes
func (k Keeper) Classes(goCtx context.Context, r *nft2.QueryClassesRequest) (*nft2.QueryClassesResponse, error) {
	if r == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := k.storeService.OpenKVStore(ctx)
	classStore := prefix.NewStore(runtime.KVStoreAdapter(store), ClassKey)

	var classes []*nft2.Class
	pageRes, err := query.Paginate(classStore, r.Pagination, func(_, value []byte) error {
		var class nft2.Class
		if err := k.cdc.Unmarshal(value, &class); err != nil {
			return err
		}
		classes = append(classes, &class)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &nft2.QueryClassesResponse{
		Classes:    classes,
		Pagination: pageRes,
	}, nil
}
