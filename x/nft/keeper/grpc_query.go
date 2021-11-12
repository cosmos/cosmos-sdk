package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

var _ nft.QueryServer = Keeper{}

// Balance return the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721
func (k Keeper) Balance(goCtx context.Context, r *nft.QueryBalanceRequest) (*nft.QueryBalanceResponse, error) {
	if r == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty request")
	}

	if err := nft.ValidateClassID(r.ClassId); err != nil {
		return nil, err
	}

	owner, err := sdk.AccAddressFromBech32(r.Owner)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	balance := k.GetBalance(ctx, r.ClassId, owner)
	return &nft.QueryBalanceResponse{Amount: balance}, nil
}

// Owner return the owner of the NFT based on its class and id, same as ownerOf in ERC721
func (k Keeper) Owner(goCtx context.Context, r *nft.QueryOwnerRequest) (*nft.QueryOwnerResponse, error) {
	if r == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty request")
	}

	if err := nft.ValidateClassID(r.ClassId); err != nil {
		return nil, err
	}

	if err := nft.ValidateNFTID(r.Id); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	owner := k.GetOwner(ctx, r.ClassId, r.Id)
	return &nft.QueryOwnerResponse{Owner: owner.String()}, nil
}

// Supply return the number of NFTs from the given class, same as totalSupply of ERC721.
func (k Keeper) Supply(goCtx context.Context, r *nft.QuerySupplyRequest) (*nft.QuerySupplyResponse, error) {
	if r == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty request")
	}

	if err := nft.ValidateClassID(r.ClassId); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	supply := k.GetTotalSupply(ctx, r.ClassId)
	return &nft.QuerySupplyResponse{Amount: supply}, nil
}

// NFTsOfClass return all NFTs of a given class or optional owner, similar to tokenByIndex in ERC721Enumerable
func (k Keeper) NFTsOfClass(goCtx context.Context, r *nft.QueryNFTsOfClassRequest) (*nft.QueryNFTsOfClassResponse, error) {
	if r == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty request")
	}

	if err := nft.ValidateClassID(r.ClassId); err != nil {
		return nil, err
	}

	var nfts []*nft.NFT
	ctx := sdk.UnwrapSDKContext(goCtx)

	// if owner is not empty, filter nft by owner
	if len(r.Owner) > 0 {
		owner, err := sdk.AccAddressFromBech32(r.Owner)
		if err != nil {
			return nil, err
		}

		ownerStore := k.getClassStoreByOwner(ctx, owner, r.ClassId)
		pageRes, err := query.Paginate(ownerStore, r.Pagination, func(key []byte, _ []byte) error {
			nft, has := k.GetNFT(ctx, r.ClassId, string(key))
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

	nftStore := k.getNFTStore(ctx, r.ClassId)
	pageRes, err := query.Paginate(nftStore, r.Pagination, func(_ []byte, value []byte) error {
		var nft nft.NFT
		if err := k.cdc.Unmarshal(value, &nft); err != nil {
			return err
		}
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

// NFT return an NFT based on its class and id.
func (k Keeper) NFT(goCtx context.Context, r *nft.QueryNFTRequest) (*nft.QueryNFTResponse, error) {
	if r == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty request")
	}

	if err := nft.ValidateClassID(r.ClassId); err != nil {
		return nil, err
	}
	if err := nft.ValidateNFTID(r.Id); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	n, has := k.GetNFT(ctx, r.ClassId, r.Id)
	if !has {
		return nil, sdkerrors.Wrapf(nft.ErrNFTNotExists, "not found nft: class: %s, id: %s", r.ClassId, r.Id)
	}
	return &nft.QueryNFTResponse{Nft: &n}, nil

}

// Class return an NFT class based on its id
func (k Keeper) Class(goCtx context.Context, r *nft.QueryClassRequest) (*nft.QueryClassResponse, error) {
	if r == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty request")
	}

	if err := nft.ValidateClassID(r.ClassId); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	class, has := k.GetClass(ctx, r.ClassId)
	if !has {
		return nil, sdkerrors.Wrapf(nft.ErrClassNotExists, "not found class: %s", r.ClassId)
	}
	return &nft.QueryClassResponse{Class: &class}, nil
}

// Classes return all NFT classes
func (k Keeper) Classes(goCtx context.Context, r *nft.QueryClassesRequest) (*nft.QueryClassesResponse, error) {
	if r == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	classStore := prefix.NewStore(store, ClassKey)

	var classes []*nft.Class
	pageRes, err := query.Paginate(classStore, r.Pagination, func(_ []byte, value []byte) error {
		var class nft.Class
		if err := k.cdc.Unmarshal(value, &class); err != nil {
			return err
		}
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
