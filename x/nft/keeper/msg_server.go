package keeper

import (
	"context"

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

func (m msgServer) Issue(ctx context.Context, issue *types.MsgIssue) (*types.MsgIssueResponse, error) {
	panic("implement me")
}

func (m msgServer) Mint(ctx context.Context, mint *types.MsgMint) (*types.MsgMintResponse, error) {
	panic("implement me")
}

func (m msgServer) Edit(ctx context.Context, edit *types.MsgEdit) (*types.MsgEditResponse, error) {
	panic("implement me")
}

func (m msgServer) Send(ctx context.Context, send *types.MsgSend) (*types.MsgSendResponse, error) {
	panic("implement me")
}

func (m msgServer) Burn(ctx context.Context, burn *types.MsgBurn) (*types.MsgBurnResponse, error) {
	panic("implement me")
}
