package ics02

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// NewHandler creates a new Handler instance for IBC client
// transactions
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgCreateClient:
			return handleMsgCreateClient(ctx, k, msg)

		case types.MsgUpdateClient:
			return handleMsgUpdateClient(ctx, k, msg)

		case types.MsgSubmitMisbehaviour:
			return handleMsgSubmitMisbehaviour(ctx, k, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized IBC Client message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateClient(ctx sdk.Context, k keeper.Keeper, msg types.MsgCreateClient) sdk.Result {
	_, err := k.CreateClient(ctx, msg.ClientID, msg.ConsensusState)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(100), err.Error()).Result()
	}

	// TODO: events
	return sdk.Result{}
}

func handleMsgUpdateClient(ctx sdk.Context, k keeper.Keeper, msg types.MsgUpdateClient) sdk.Result {
	state, err := k.Query(ctx, msg.ClientID)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(200), err.Error()).Result()
	}

	err = state.Update(ctx, msg.Header)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(300), err.Error()).Result()
	}

	// TODO: events
	return sdk.Result{}
}

func handleMsgSubmitMisbehaviour(ctx sdk.Context, k keeper.Keeper, msg types.MsgSubmitMisbehaviour) sdk.Result {
	state, err := k.Query(ctx, msg.ClientID)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(200), err.Error()).Result()
	}

	err = k.CheckMisbehaviourAndUpdateState(ctx, state, msg.Evidence)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType("ibc"), sdk.CodeType(200), err.Error()).Result()
	}

	// TODO: events
	return sdk.Result{}
}
