package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the nft MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// Issue implement Issue method of the types.MsgServer.
func (m msgServer) Issue(goCtx context.Context, msg *types.MsgIssue) (*types.MsgIssueResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if msg.Metadata == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidMetadata, "metadata is empty")
	}

	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.Issue(ctx, msg.Metadata.Type,
		msg.Metadata.Name,
		msg.Metadata.Symbol,
		msg.Metadata.Description,
		msg.Metadata.MintRestricted,
		msg.Metadata.EditRestricted,
		issuer); err != nil {
		return nil, err
	}
	return &types.MsgIssueResponse{}, nil
}

// Mint implement Mint method of the types.MsgServer.
func (m msgServer) Mint(goCtx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if msg.NFT == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidNFT, "nft is empty")
	}

	minter, err := sdk.AccAddressFromBech32(msg.Minter)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.MintNFT(ctx, msg.NFT.Type,
		msg.NFT.ID,
		msg.NFT.URI,
		msg.NFT.Data,
		minter); err != nil {
		return nil, err
	}
	return &types.MsgMintResponse{}, nil
}

// Edit implement Edit method of the types.MsgServer.
func (m msgServer) Edit(goCtx context.Context, msg *types.MsgEdit) (*types.MsgEditResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if msg.NFT == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidNFT, "nft is empty")
	}

	editor, err := sdk.AccAddressFromBech32(msg.Editor)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.EditNFT(ctx, msg.NFT.Type,
		msg.NFT.ID,
		msg.NFT.URI,
		msg.NFT.Data,
		editor); err != nil {
		return nil, err
	}
	return &types.MsgEditResponse{}, nil
}

// Send implement Send method of the types.MsgServer.
func (m msgServer) Send(goCtx context.Context, msg *types.MsgSend) (*types.MsgSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.SendNFT(ctx, msg.Type,
		msg.ID,
		sender,
		receiver); err != nil {
		return nil, err
	}
	return &types.MsgSendResponse{}, nil
}

// Burn implement Burn method of the types.MsgServer.
func (m msgServer) Burn(goCtx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	destroyer, err := sdk.AccAddressFromBech32(msg.Destroyer)
	if err != nil {
		return nil, err
	}

	if err := m.Keeper.BurnNFT(ctx, msg.Type,
		msg.ID,
		destroyer); err != nil {
		return nil, err
	}
	return &types.MsgBurnResponse{}, nil
}
