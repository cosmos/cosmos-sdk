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
	for _, token := range tokens {
		if err := k.Mint(ctx, token, receiver); err != nil {
			return err
		}
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
		if err := k.Burn(ctx, classID, nftID); err != nil {
			return err
		}
	}
	return nil
}

// BatchUpdate defines a method for updating a batch of exist nfts
// Note: When the upper module uses this method, it needs to authenticate nft
func (k Keeper) BatchUpdate(ctx sdk.Context, tokens []nft.NFT) error {
	for _, token := range tokens {
		if err := k.Update(ctx, token); err != nil {
			return err
		}
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
		if err := k.Transfer(ctx, classID, nftID, receiver); err != nil {
			return sdkerrors.Wrap(nft.ErrNFTNotExists, nftID)
		}
	}
	return nil
}
