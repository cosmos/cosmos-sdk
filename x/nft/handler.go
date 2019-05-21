package nft

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/tags"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

// NewHandler routes the messages to the handlers
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgTransferNFT:
			return HandleMsgTransferNFT(ctx, msg, k)
		case types.MsgEditNFTMetadata:
			return HandleMsgEditNFTMetadata(ctx, msg, k)
		default:
			errMsg := fmt.Sprintf("unrecognized nft message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// HandleMsgTransferNFT handler for MsgTransferNFT
func HandleMsgTransferNFT(ctx sdk.Context, msg types.MsgTransferNFT, k keeper.Keeper,
) sdk.Result {

	nft, err := k.GetNFT(ctx, msg.Denom, msg.ID)
	if err != nil {
		return err.Result()
	}

	if !nft.GetOwner().Equals(msg.Sender) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("%s is not the owner of NFT #%d", msg.Sender.String(), msg.ID)).Result()
	}

	// delete NFT from original owner balance
	ownerBalance, found := k.GetBalance(ctx, nft.GetOwner(), msg.Denom)
	if !found {
		// safety check
		panic(fmt.Sprintf("NFT #%d is not registered in it's original owner's balance (%s)", nft.GetID(), nft.GetOwner()))
	}

	err = ownerBalance.DeleteNFT(nft)
	if err != nil {
		return err.Result()
	}

	// update NFT owner and new owner balance
	nftTranfered := nft.SetOwner(msg.Recipient)

	recipientBalance, found := k.GetBalance(ctx, msg.Recipient, msg.Denom)
	if !found {
		recipientBalance = types.NewCollection(msg.Denom, types.NewNFTs(nftTranfered))
	} else {
		recipientBalance.AddNFT(nftTranfered)
	}

	// save new NFT in the collection and balance
	k.SetBalance(ctx, msg.Recipient, recipientBalance)
	err = k.SetNFT(ctx, msg.Denom, nftTranfered)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: sdk.NewTags(
			tags.Category, tags.TxCategory,
			tags.Sender, msg.Sender.String(),
			tags.Recipient, msg.Recipient.String(),
			tags.Denom, string(msg.Denom),
			tags.NFTID, uint64(msg.ID),
		),
	}
}

// HandleMsgEditNFTMetadata handler for MsgEditNFTMetadata
func HandleMsgEditNFTMetadata(ctx sdk.Context, msg types.MsgEditNFTMetadata, k keeper.Keeper,
) sdk.Result {

	nft, err := k.GetNFT(ctx, msg.Denom, msg.ID)
	if err != nil {
		return err.Result()
	}

	// check if msg sender is the Owner of the NFT
	if !nft.GetOwner().Equals(msg.Owner) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("%s is not the owner of NFT #%d", msg.Owner.String(), msg.ID)).Result()
	}

	// edit NFT
	nftEdited := nft.EditMetadata(msg.Name, msg.Description, msg.Image, msg.TokenURI)
	err = k.SetNFT(ctx, msg.Denom, nftEdited)
	if err != nil {
		return err.Result()
	}

	// update original NFT on the owner's balance
	balance, found := k.GetBalance(ctx, msg.Owner, msg.Denom)
	if !found {
		// safety check
		panic(fmt.Sprintf("NFT #%d is not registered in it's original owner's balance (%s)", nftEdited.GetID(), nftEdited.GetOwner()))
	}

	err = balance.UpdateNFT(nftEdited)
	if err != nil {
		return err.Result()
	}

	// save new NFT in the collection and balance
	k.SetBalance(ctx, msg.Owner, balance)
	err = k.SetNFT(ctx, msg.Denom, nftEdited)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: sdk.NewTags(
			tags.Category, tags.TxCategory,
			tags.Sender, msg.Owner.String(),
			tags.Denom, string(msg.Denom),
			tags.NFTID, msg.ID,
		),
	}
}
