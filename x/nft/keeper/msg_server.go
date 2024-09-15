package keeper

import (
	"bytes"
	"context"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/nft"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ nft.MsgServer = Keeper{}

// Send implements Send method of the types.MsgServer.
func (k Keeper) Send(ctx context.Context, msg *nft.MsgSend) (*nft.MsgSendResponse, error) {
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

	receiver, err := k.ac.StringToBytes(msg.Receiver)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid receiver address (%s)", msg.Receiver)
	}

	if err := k.Transfer(ctx, msg.ClassId, msg.Id, sender, receiver); err != nil {
		return nil, err
	}

	if err = k.EventService.EventManager(ctx).Emit(&nft.EventSend{
		ClassId:  msg.ClassId,
		Id:       msg.Id,
		Sender:   msg.Sender,
		Receiver: msg.Receiver,
	}); err != nil {
		return nil, err
	}

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
		Creator: msg.Creator,
		Owner:   msg.Owner,
	}

	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, err
	}

	// If owner is not specified, set it to the sender
	if token.Owner == "" {
		token.Owner = msg.Sender
	}

	// If creator is not specified, set it to the sender
	if token.Creator == "" {
		token.Creator = msg.Sender
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
	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", msg.Sender)
	}

	// Check if the sender is the owner of the NFT
	owner := k.GetOwner(ctx, msg.ClassId, msg.Id)
	if !bytes.Equal(owner, sender) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "Sender is not the owner of the NFT")
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

// StreamNFT handles the MsgStreamNFT message
func (k Keeper) StreamNFT(goCtx context.Context, msg *nft.MsgStreamNFT) (*nft.MsgStreamNFTResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	payment, err := sdk.ParseCoinNormalized(msg.Payment)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, err.Error())
	}

	sender, err := k.ac.StringToBytes(msg.Sender)
	if err != nil {
		return nil, err
	}

	err = k.bk.SendCoinsFromAccountToModule(ctx, sender, nft.ModuleName, sdk.NewCoins(payment))
	if err != nil {
		return nil, err
	}

	// Increment total plays
	err = k.IncrementTotalPlays(ctx, msg.ClassId, msg.Id, msg.PlayCount)
	if err != nil {
		return nil, err
	}

	err = k.StreamPayment(ctx, msg.ClassId, msg.Id, payment)
	if err != nil {
		return nil, err
	}

	return &nft.MsgStreamNFTResponse{}, nil
}

// WithdrawRoyalties handles the MsgWithdrawRoyalties message
func (k Keeper) WithdrawRoyalties(goCtx context.Context, msg *nft.MsgWithdrawRoyalties) (*nft.MsgWithdrawRoyaltiesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Extract necessary variables
	classID := msg.ClassId
	nftID := msg.Id
	role := msg.Role
	caller, err := sdk.AccAddressFromBech32(msg.Recipient) // Use msg.Recipient as the caller
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrap("Invalid recipient address")
	}

	// Check if the caller can withdraw royalties
	if !k.canWithdrawRoyalties(ctx, classID, nftID, role, caller) {
		return nil, sdkerrors.ErrUnauthorized.Wrap("caller cannot withdraw royalties for this role")
	}

	recipient, err := k.ac.StringToBytes(msg.Recipient)
	if err != nil {
		return nil, err
	}

	amount, err := k.WithdrawRoyaltiesInternal(ctx, msg.ClassId, msg.Id, msg.Role, recipient)
	if err != nil {
		return nil, err
	}

	return &nft.MsgWithdrawRoyaltiesResponse{
		Amount: amount.String(),
	}, nil
}

func (k Keeper) canWithdrawRoyalties(ctx context.Context, classID, nftID, role string, caller sdk.AccAddress) bool {
	nft, found := k.GetNFT(ctx, classID, nftID)
	if !found {
		return false
	}

	switch role {
	case "creator":
		return nft.Creator == caller.String()
	case "owner":
		return nft.Owner == caller.String()
	case "platform":
		platformAddress := sdk.MustAccAddressFromBech32("cosmos1d9ms9wf4yx3vky2kp6fc7t3qm9p8ps33g49c9s")
		return caller.Equals(platformAddress)
	default:
		return false
	}
}
