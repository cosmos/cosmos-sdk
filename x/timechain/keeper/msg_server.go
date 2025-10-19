// SPDX-License-Identifier: MPL-2.0
// Copyright © 2025 Timechain-Arweave-LunCoSim Contributors

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/timechain/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) ProposeSlot(goCtx context.Context, msg *types.MsgProposeSlot) (*types.MsgProposeSlotResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.MsgProposeSlotResponse{}, k.Keeper.ProposeSlot(ctx, msg)
}

func (k msgServer) ConfirmSlot(goCtx context.Context, msg *types.MsgConfirmSlot) (*types.MsgConfirmSlotResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.MsgConfirmSlotResponse{}, k.Keeper.ConfirmSlot(ctx, msg)
}

func (k msgServer) RelayEvent(goCtx context.Context, msg *types.MsgRelayEvent) (*types.MsgRelayEventResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Implementation to be added in a future step
	return &types.MsgRelayEventResponse{}, nil
}
