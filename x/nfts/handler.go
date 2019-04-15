package nfts

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nfts/tags"
)

// NewHandler routes the messages to the handlers
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgTransferNFT:
			return handleMsgTransferNFT(ctx, msg, k)
		case MsgEditNFTMetadata:
			return handleMsgEditNFTMetadata(ctx, msg, k)
		case MsgMintNFT:
			return handleMsgMintNFT(ctx, msg, k)
		case MsgBurnNFT:
			return handleMsgBurnNFT(ctx, msg, k)
		// case MsgBuyNFT:
		// 	return handleMsgBuyNFT(ctx, msg, k)
		default:
			errMsg := fmt.Sprintf("unrecognized nft message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgTransferNFT(ctx sdk.Context, msg MsgTransferNFT, k Keeper,
) sdk.Result {

	nft, err := k.GetNFT(ctx, msg.Denom, msg.ID)
	if err != nil {
		return err.Result()
	}

	if !nft.Owner.Equals(msg.Sender) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("%s is not the owner of NFT #%d", msg.Sender.String(), msg.ID)).Result()
	}

	// remove from previous owner
	owner, found := k.GetOwner(ctx, nft.Owner)
	if !found {
		return ErrInvalidNFT(DefaultCodespace).Result()
	}
	owner.RemoveNFT(msg.Denom, msg.ID)
	k.SetOwner(ctx, nft.Owner, owner)

	// update NFT
	nft.Owner = msg.Recipient

	// add to new owner
	k.AddToOwner(ctx, msg.Denom, msg.ID, nft)

	// save new NFT
	err = k.SetNFT(ctx, msg.Denom, msg.ID, nft)
	if err != nil {
		return err.Result()
	}

	resTags := sdk.NewTags(
		tags.Category, tags.TxCategory,
		tags.Sender, msg.Sender.String(),
		tags.Recipient, msg.Recipient.String(),
		tags.Denom, string(msg.Denom),
		tags.NFTID, uint64(msg.ID),
	)
	return sdk.Result{
		Tags: resTags,
	}
}

func handleMsgEditNFTMetadata(ctx sdk.Context, msg MsgEditNFTMetadata, k Keeper,
) sdk.Result {

	nft, err := k.GetNFT(ctx, msg.Denom, msg.ID)
	if err != nil {
		return err.Result()
	}

	// Make sure msg sender (Owner) is actually the Owner of the NFT
	if !nft.Owner.Equals(msg.Owner) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("%s is not the owner of NFT #%d", msg.Owner.String(), msg.ID)).Result()
	}

	nft = nft.EditMetadata(msg.Name, msg.Description, msg.Image, msg.TokenURI)
	err = k.SetNFT(ctx, msg.Denom, msg.ID, nft)
	if err != nil {
		return err.Result()
	}

	resTags := sdk.NewTags(
		tags.Category, tags.TxCategory,
		tags.Sender, msg.Owner.String(),
		tags.Denom, string(msg.Denom),
		tags.NFTID, msg.ID,
	)
	return sdk.Result{
		Tags: resTags,
	}
}

// TODO: move to separate Module?
func handleMsgMintNFT(ctx sdk.Context, msg MsgMintNFT, k Keeper,
) sdk.Result {

	// make sure NFT with that ID and denom doesn't exist
	exists := k.IsNFT(ctx, msg.Denom, msg.ID)
	if exists {
		return ErrNFTAlreadyExists(DefaultCodespace, fmt.Sprintf("%s NFT with id %d already exists", msg.Denom, msg.ID)).Result()
	}

	// make sure collection exists, if not create it
	_, found := k.GetCollection(ctx, msg.Denom)
	if !found {
		k.SetCollection(ctx, msg.Denom, NewCollection())
	}

	// make new NFT and set it
	nft := NewNFT(msg.Recipient, msg.TokenURI, msg.Description, msg.Image, msg.Name)
	err := k.SetNFT(ctx, msg.Denom, msg.ID, nft)
	if err != nil {
		return err.Result()
	}

	// add ne NFT to Owners
	k.AddToOwner(ctx, msg.Denom, msg.ID, nft)

	resTags := sdk.NewTags(
		tags.Category, tags.TxCategory,
		tags.Sender, msg.Sender.String(),
		tags.Recipient, string(msg.Recipient),
		tags.Denom, string(msg.Denom),
		tags.NFTID, string(msg.ID),
	)
	return sdk.Result{
		Tags: resTags,
	}
}

func handleMsgBurnNFT(ctx sdk.Context, msg MsgBurnNFT, k Keeper,
) sdk.Result {

	nft, err := k.GetNFT(ctx, msg.Denom, msg.ID)
	if err != nil {
		return err.Result()
	}

	if !nft.Owner.Equals(msg.Sender) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("%s is not the owner of %s NFT #%d", msg.Sender.String(), msg.Denom, msg.ID)).Result()
	}

	// remove from owner
	owner, found := k.GetOwner(ctx, nft.Owner)
	if !found {
		return ErrInvalidNFT(DefaultCodespace).Result()
	}
	owner.RemoveNFT(msg.Denom, msg.ID)
	k.SetOwner(ctx, nft.Owner, owner)

	// remove actual NFT
	err = k.BurnNFT(ctx, msg.Denom, msg.ID)
	if err != nil {
		return err.Result()
	}

	resTags := sdk.NewTags(
		tags.Category, tags.TxCategory,
		tags.Sender, msg.Sender.String(),
		tags.Denom, string(msg.Denom),
		tags.NFTID, string(msg.ID),
	)
	return sdk.Result{
		Tags: resTags,
	}
}

// func handleMsgBuyNFT(ctx sdk.Context, msg MsgBuyNFT, k Keeper,
// ) sdk.Result {

// 	nft, err := k.GetNFT(ctx, msg.Denom, msg.ID)
// 	if err != nil {
// 		return err.Result()
// 	}

// 	owner, found := k.GetOwner(ctx, nft.Owner)
// 	if !found {
// 		panic(fmt.Sprintf("%s should have an ownership relation with NFT %d", nft.Owner, msg.ID))
// 	}
// 	// owner[msg.Denom]

// 	_, err = k.bk.SubtractCoins(msg.Sender, msg.Amount)
// 	if err != nil {
// 		return err.Result()
// 	}
// 	_, err = k.bk.AddCoins(nft.Owner, msg.Amount)
// 	if err != nil {
// 		return err.Result()
// 	}

// 	nft.Owner = msg.Sender

// 	// TODO: add to new owners ownership

// 	err = keepr.SetNFT(ctx, nft)
// 	if err != nil {
// 		return err.Result()
// 	}

// 	resTags := sdk.NewTags(
// 		tags.Category, tags.TxCategory,
// 		tags.Sender, msg.Sender.String(),
// 		tags.Owner, msg.Owner.String(),
// 		tags.Denom, msg.Denom.String(),
// 		tags.NFTID, msg.ID,
// 	)
// 	return sdk.Result{
// 		Tags: resTags,
// 	}
// }
