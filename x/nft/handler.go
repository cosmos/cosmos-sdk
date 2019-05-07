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
			return handleMsgTransferNFT(ctx, msg, k)
		case types.MsgEditNFTMetadata:
			return HandleMsgEditNFTMetadata(ctx, msg, k)
		default:
			errMsg := fmt.Sprintf("unrecognized nft message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgTransferNFT(ctx sdk.Context, msg types.MsgTransferNFT, k keeper.Keeper,
) sdk.Result {

	nft, err := k.GetNFT(ctx, msg.Denom, msg.ID)
	if err != nil {
		return err.Result()
	}

	if !nft.Owner.Equals(msg.Sender) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("%s is not the owner of NFT #%d", msg.Sender.String(), msg.ID)).Result()
	}

	// delete NFT from original owner balance

	// update NFT owner
	nft.Owner = msg.Recipient

	// save new NFT in the collection and balance
	err = k.SetNFT(ctx, msg.Denom, nft)
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

func HandleMsgEditNFTMetadata(ctx sdk.Context, msg types.MsgEditNFTMetadata, k keeper.Keeper,
) sdk.Result {

	nft, err := k.GetNFT(ctx, msg.Denom, msg.ID)
	if err != nil {
		return err.Result()
	}

	// check if msg sender is the Owner of the NFT
	if !nft.Owner.Equals(msg.Owner) {
		return sdk.ErrInvalidAddress(fmt.Sprintf("%s is not the owner of NFT #%d", msg.Owner.String(), msg.ID)).Result()
	}

	nft = nft.EditMetadata(msg.Name, msg.Description, msg.Image, msg.TokenURI)
	err = k.SetNFT(ctx, msg.Denom, msg.ID, nft)
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
