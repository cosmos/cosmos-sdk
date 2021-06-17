package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/nft"
)

var _ nft.MsgServer = keeper{}

func (k keeper) IssueCollection(goCtx context.Context, msg *nft.MsgIssueCollection) (*nft.MsgIssueCollectionResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.issueCollection(ctx, msg.Id, msg.Name, msg.Schema, sender); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			nft.EventTypeIssueCollection,
			sdk.NewAttribute(nft.AttributeKeyCollectionID, msg.Id),
			sdk.NewAttribute(nft.AttributeKeyCollectionName, msg.Name),
			sdk.NewAttribute(nft.AttributeKeyCreator, msg.Sender),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, nft.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &nft.MsgIssueCollectionResponse{}, nil
}

func (k keeper) MintNFT(goCtx context.Context, msg *nft.MsgMintNFT) (*nft.MsgMintNFTResponse, error) {
	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.mintNFT(ctx, msg.CollectionId, msg.Id,
		msg.Name,
		msg.Uri,
		msg.Data,
		recipient,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			nft.EventTypeMintNFT,
			sdk.NewAttribute(nft.AttributeKeyTokenID, msg.Id),
			sdk.NewAttribute(nft.AttributeKeyCollectionID, msg.CollectionId),
			sdk.NewAttribute(nft.AttributeKeyTokenURI, msg.Uri),
			sdk.NewAttribute(nft.AttributeKeyRecipient, msg.Recipient),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, nft.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &nft.MsgMintNFTResponse{}, nil
}

func (k keeper) EditNFT(ctx context.Context, nft *nft.MsgEditNFT) (*nft.MsgEditNFTResponse, error) {
	panic("implement me")
}

func (k keeper) BurnNFT(goCtx context.Context, msg *nft.MsgBurnNFT) (*nft.MsgBurnNFTResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.burnNFT(ctx, msg.CollectionId, msg.Id, sender); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			nft.EventTypeBurnNFT,
			sdk.NewAttribute(nft.AttributeKeyCollectionID, msg.CollectionId),
			sdk.NewAttribute(nft.AttributeKeyTokenID, msg.Id),
			sdk.NewAttribute(nft.AttributeKeyOwner, msg.Sender),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, nft.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
		),
	})

	return &nft.MsgBurnNFTResponse{}, nil
}
