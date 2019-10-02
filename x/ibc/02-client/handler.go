package ics02

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// NewHandler creates a new Handler instance for IBC client
// transactions
func NewHandler(manager types.Manager) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgCreateClient:
			return handleMsgCreateClient(ctx, manager, msg)

		case types.MsgUpdateClient:
			return handleMsgUpdateClient(ctx, manager, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized IBC Client message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateClient(ctx sdk.Context, manager types.Manager, msg types.MsgCreateClient) sdk.Result {
	_, err := manager.Create(ctx, msg.ClientID, msg.ConsensusState)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(100), err.Error()).Result()
	}

	// TODO: events
	return sdk.Result{}
}

func handleMsgUpdateClient(ctx sdk.Context, manager types.Manager, msg types.MsgUpdateClient) sdk.Result {
	obj, err := manager.Query(ctx, msg.ClientID)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(200), err.Error()).Result()
	}
	err = obj.Update(ctx, msg.Header)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(300), err.Error()).Result()
	}

	// TODO: events
	return sdk.Result{}
}
