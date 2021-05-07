package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/poolx/types"
)

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

type msgServer struct {
	Keeper
}

func (msgServer) CreatePool(context.Context, *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	return nil, nil // TODO(levi) implement
}

func (msgServer) FundPool(context.Context, *types.MsgFundPool) (*types.MsgFundPoolResponse, error) {
	return nil, nil // TODO(levi) implement
}
