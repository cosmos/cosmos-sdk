package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

// BatchMint defines a method for minting a batch of nfts
func (k Keeper) BatchMint(ctx sdk.Context,
	tokens []nft.NFT,
	receiver sdk.AccAddress,
) error {
	classIDs := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		if !classIDs[token.ClassId] && !k.HasClass(ctx, token.ClassId) {
			return sdkerrors.Wrap(nft.ErrClassNotExists, token.ClassId)
		}

		if k.HasNFT(ctx, token.ClassId, token.Id) {
			return sdkerrors.Wrap(nft.ErrNFTExists, token.Id)
		}

		k.setNFT(ctx, token)
		k.setOwner(ctx, token.ClassId, token.Id, receiver)
		k.incrTotalSupply(ctx, token.ClassId)

		ctx.EventManager().EmitTypedEvent(&nft.EventMint{
			ClassId: token.ClassId,
			Id:      token.Id,
			Owner:   receiver.String(),
		})
		classIDs[token.ClassId] = true
	}
	return nil
}

// BatchBurn defines a method for burning a batch of nfts from a specific classID.
// Note: When the upper module uses this method, it needs to authenticate nft
func (k Keeper) BatchBurn(ctx sdk.Context, classID string, nftIDs []string) error {
	if !k.HasClass(ctx, classID) {
		return sdkerrors.Wrap(nft.ErrClassNotExists, classID)
	}

	for _, nftID := range nftIDs {
		if !k.HasNFT(ctx, classID, nftID) {
			return sdkerrors.Wrap(nft.ErrNFTNotExists, nftID)
		}

		owner := k.GetOwner(ctx, classID, nftID)
		nftStore := k.getNFTStore(ctx, classID)
		nftStore.Delete([]byte(nftID))

		k.deleteOwner(ctx, classID, nftID, owner)
		ctx.EventManager().EmitTypedEvent(&nft.EventBurn{
			ClassId: classID,
			Id:      nftID,
			Owner:   owner.String(),
		})
	}
	k.updateTotalSupply(
		ctx,
		classID,
		k.GetTotalSupply(ctx, classID)-uint64(len(nftIDs)),
	)
	return nil
}

// BatchUpdate defines a method for updating a batch of exist nfts
// Note: When the upper module uses this method, it needs to authenticate nft
func (k Keeper) BatchUpdate(ctx sdk.Context, tokens []nft.NFT) error {
	for _, token := range tokens {
		if !k.HasClass(ctx, token.ClassId) {
			return sdkerrors.Wrap(nft.ErrClassNotExists, token.ClassId)
		}

		if !k.HasNFT(ctx, token.ClassId, token.Id) {
			return sdkerrors.Wrap(nft.ErrNFTNotExists, token.Id)
		}
		k.setNFT(ctx, token)
	}
	return nil
}

// BatchTransfer defines a method for sending a batch of nfts from one account to another account from a specific classID.
// Note: When the upper module uses this method, it needs to authenticate nft
func (k Keeper) BatchTransfer(ctx sdk.Context,
	classID string,
	nftIDs []string,
	receiver sdk.AccAddress,
) error {
	for _, nftID := range nftIDs {
		if !k.HasClass(ctx, classID) {
			return sdkerrors.Wrap(nft.ErrClassNotExists, classID)
		}

		if !k.HasNFT(ctx, classID, nftID) {
			return sdkerrors.Wrap(nft.ErrNFTNotExists, nftID)
		}

		owner := k.GetOwner(ctx, classID, nftID)
		k.deleteOwner(ctx, classID, nftID, owner)
		k.setOwner(ctx, classID, nftID, receiver)
	}
	return nil
}
