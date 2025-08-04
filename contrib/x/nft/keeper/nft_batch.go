package keeper

import (
	"context"

	"cosmossdk.io/errors"

	nft2 "github.com/cosmos/cosmos-sdk/contrib/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BatchMint defines a method for minting a batch of nfts
func (k Keeper) BatchMint(ctx context.Context,
	tokens []nft2.NFT,
	receiver sdk.AccAddress,
) error {
	checked := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		if !checked[token.ClassId] && !k.HasClass(ctx, token.ClassId) {
			return errors.Wrap(nft2.ErrClassNotExists, token.ClassId)
		}

		if k.HasNFT(ctx, token.ClassId, token.Id) {
			return errors.Wrap(nft2.ErrNFTExists, token.Id)
		}

		checked[token.ClassId] = true
		if err := k.mintWithNoCheck(ctx, token, receiver); err != nil {
			return err
		}
	}
	return nil
}

// BatchBurn defines a method for burning a batch of nfts from a specific classID.
// Note: When the upper module uses this method, it needs to authenticate nft
func (k Keeper) BatchBurn(ctx context.Context, classID string, nftIDs []string) error {
	if !k.HasClass(ctx, classID) {
		return errors.Wrap(nft2.ErrClassNotExists, classID)
	}
	for _, nftID := range nftIDs {
		if !k.HasNFT(ctx, classID, nftID) {
			return errors.Wrap(nft2.ErrNFTNotExists, nftID)
		}
		if err := k.burnWithNoCheck(ctx, classID, nftID); err != nil {
			return err
		}
	}
	return nil
}

// BatchUpdate defines a method for updating a batch of exist nfts
// Note: When the upper module uses this method, it needs to authenticate nft
func (k Keeper) BatchUpdate(ctx context.Context, tokens []nft2.NFT) error {
	checked := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		if !checked[token.ClassId] && !k.HasClass(ctx, token.ClassId) {
			return errors.Wrap(nft2.ErrClassNotExists, token.ClassId)
		}

		if !k.HasNFT(ctx, token.ClassId, token.Id) {
			return errors.Wrap(nft2.ErrNFTNotExists, token.Id)
		}
		checked[token.ClassId] = true
		k.updateWithNoCheck(ctx, token)
	}
	return nil
}

// BatchTransfer defines a method for sending a batch of nfts from one account to another account from a specific classID.
// Note: When the upper module uses this method, it needs to authenticate nft
func (k Keeper) BatchTransfer(ctx context.Context,
	classID string,
	nftIDs []string,
	receiver sdk.AccAddress,
) error {
	if !k.HasClass(ctx, classID) {
		return errors.Wrap(nft2.ErrClassNotExists, classID)
	}
	for _, nftID := range nftIDs {
		if !k.HasNFT(ctx, classID, nftID) {
			return errors.Wrap(nft2.ErrNFTNotExists, nftID)
		}
		if err := k.transferWithNoCheck(ctx, classID, nftID, receiver); err != nil {
			return errors.Wrap(nft2.ErrNFTNotExists, nftID)
		}
	}
	return nil
}
