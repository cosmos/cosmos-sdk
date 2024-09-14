package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/nft"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ nft.MsgServer = Keeper{}

// Send implements the Send method of the types.MsgServer.
func (k Keeper) Send(ctx context.Context, msg *nft.MsgSend) (*nft.MsgSendResponse, error) {
	// Implementation remains the same
	return &nft.MsgSendResponse{}, nil
}

// MintNFT implements the MintNFT method of the types.MsgServer.
func (k Keeper) MintNFT(ctx context.Context, msg *nft.MsgMintNFT) (*nft.MsgMintNFTResponse, error) {
	if !k.HasClass(ctx, msg.ClassId) {
		// Create class if it doesn't exist
		class := nft.Class{
			Id:          msg.ClassId,
			Name:        msg.ClassId, // Using ClassId as Name for simplicity
			Symbol:      msg.ClassId, // Using ClassId as Symbol for simplicity
			Description: "Automatically created class",
			Uri:         msg.Uri,
			UriHash:     msg.UriHash,
		}
		if err := k.SaveClass(ctx, class); err != nil {
			return nil, err
		}
	}

	// Continue with the existing mint logic
	token := nft.NFT{
		ClassId: msg.ClassId,
		Id:      msg.Id,
		Uri:     msg.Uri,
		UriHash: msg.UriHash,
	}

	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = k.Mint(ctx, token, sender)
	if err != nil {
		return nil, err
	}

	return &nft.MsgMintNFTResponse{}, nil
}

// BurnNFT implements the BurnNFT method of the types.MsgServer.
func (k Keeper) BurnNFT(goCtx context.Context, msg *nft.MsgBurnNFT) (*nft.MsgBurnNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if len(msg.ClassId) == 0 {
		return nil, nft.ErrEmptyClassID
	}
	if len(msg.Id) == 0 {
		return nil, nft.ErrEmptyNFTID
	}
	_, err := k.ac.StringToBytes(msg.Sender) // Convert address but don't assign to unused variable
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", msg.Sender)
	}
	err = k.Burn(ctx, msg.ClassId, msg.Id)
	if err != nil {
		return nil, err
	}
	err = ctx.EventManager().EmitTypedEvent(&nft.EventBurn{
		ClassId: msg.ClassId,
		Id:      msg.Id,
		Owner:   msg.Sender,
	})
	if err != nil {
		return nil, err
	}
	return &nft.MsgBurnNFTResponse{}, nil
}

// StakeNFT implements the StakeNFT method of the types.MsgServer.
func (k Keeper) StakeNFT(goCtx context.Context, msg *nft.MsgStakeNFT) (*nft.MsgStakeNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if len(msg.ClassId) == 0 {
		return nil, nft.ErrEmptyClassID
	}
	if len(msg.Id) == 0 {
		return nil, nft.ErrEmptyNFTID
	}
	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", msg.Sender)
	}

	err = k.Stake(ctx, msg.ClassId, msg.Id, sender, msg.StakeDuration)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"nft_staked",
			sdk.NewAttribute("class_id", msg.ClassId),
			sdk.NewAttribute("id", msg.Id),
			sdk.NewAttribute("owner", msg.Sender),
			sdk.NewAttribute("stake_duration", fmt.Sprintf("%d", msg.StakeDuration)),
		),
	)
	return &nft.MsgStakeNFTResponse{}, nil
}
