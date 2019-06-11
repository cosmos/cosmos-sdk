package nft

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/tags"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewHandler routes the messages to the handlers
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgTransferNFT:
			return HandleMsgTransferNFT(ctx, msg, k)
		case types.MsgEditNFTMetadata:
			return HandleMsgEditNFTMetadata(ctx, msg, k)
		case types.MsgMintNFT:
			return HandleMsgMintNFT(ctx, msg, k)
		case types.MsgBurnNFT:
			return HandleMsgBurnNFT(ctx, msg, k)
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
		return sdk.ErrInvalidAddress(fmt.Sprintf("%s is not the owner of NFT #%s", msg.Sender.String(), msg.ID)).Result()
	}

	// update NFT owner
	nftTranfered := nft.SetOwner(msg.Recipient)

	// update the NFT (owners are updated within the keeper)
	err = k.UpdateNFT(ctx, msg.Denom, nftTranfered)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: sdk.NewTags(
			tags.Category, tags.TxCategory,
			tags.Sender, msg.Sender.String(),
			tags.Recipient, msg.Recipient.String(),
			tags.Denom, msg.Denom,
			tags.NFTID, msg.ID,
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
		return sdk.ErrInvalidAddress(fmt.Sprintf("%s is not the owner of NFT #%s", msg.Owner.String(), msg.ID)).Result()
	}

	// update NFT
	nftEdited := nft.EditMetadata(msg.Name, msg.Description, msg.Image, msg.TokenURI)
	err = k.UpdateNFT(ctx, msg.Denom, nftEdited)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: sdk.NewTags(
			tags.Category, tags.TxCategory,
			tags.Sender, msg.Owner.String(),
			tags.Denom, msg.Denom,
			tags.NFTID, msg.ID,
		),
	}
}

// HandleMsgMintNFT handles MsgMintNFT
func HandleMsgMintNFT(ctx sdk.Context, msg types.MsgMintNFT, k keeper.Keeper,
) sdk.Result {

	nft := types.NewBaseNFT(msg.ID, msg.Recipient, msg.Name, msg.Description, msg.Image, msg.TokenURI)
	err := k.MintNFT(ctx, msg.Denom, nft)
	if err != nil {
		return err.Result()
	}

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

// HandleMsgBurnNFT handles MsgBurnNFT
func HandleMsgBurnNFT(ctx sdk.Context, msg types.MsgBurnNFT, k keeper.Keeper,
) sdk.Result {

	// remove  NFT
	err := k.DeleteNFT(ctx, msg.Denom, msg.ID)
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

// EndBlocker is run at the end of the block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) ([]abci.ValidatorUpdate, sdk.Tags) {
	return nil, nil
}
