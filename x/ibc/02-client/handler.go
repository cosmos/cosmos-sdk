package client

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// HandleMsgCreateClient defines the sdk.Handler for MsgCreateClient
func HandleMsgCreateClient(ctx sdk.Context, k keeper.Keeper, msg *types.MsgCreateClient) (*sdk.Result, error) {
	clientState, err := types.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, err
	}

	consensusState, err := types.UnpackConsensusState(msg.ConsensusState)
	if err != nil {
		return nil, err
	}

	_, err = k.CreateClient(ctx, msg.ClientId, clientState, consensusState)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateClient,
			sdk.NewAttribute(types.AttributeKeyClientID, msg.ClientId),
			sdk.NewAttribute(types.AttributeKeyClientType, clientState.ClientType().String()),
			sdk.NewAttribute(types.AttributeKeyConsensusHeight, fmt.Sprintf("%d", consensusState.GetHeight())),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgUpdateClient defines the sdk.Handler for MsgUpdateClient
func HandleMsgUpdateClient(ctx sdk.Context, k keeper.Keeper, msg *types.MsgUpdateClient) (*sdk.Result, error) {
	header, err := types.UnpackHeader(msg.Header)
	if err != nil {
		return nil, err
	}

	_, err = k.UpdateClient(ctx, msg.ClientId, header)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgSubmitMisbehaviour defines the Evidence module handler for submitting a
// light client misbehaviour.
func HandleMsgSubmitMisbehaviour(ctx sdk.Context, k keeper.Keeper, msg *types.MsgSubmitMisbehaviour) (*sdk.Result, error) {
	misbehaviour, err := types.UnpackMisbehaviour(msg.Misbehaviour)
	if err != nil {
		return nil, err
	}

	if err := k.CheckMisbehaviourAndUpdateState(ctx, misbehaviour); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to process misbehaviour for IBC client")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitMisbehaviour,
			sdk.NewAttribute(types.AttributeKeyClientID, msg.ClientId),
			sdk.NewAttribute(types.AttributeKeyClientType, misbehaviour.ClientType().String()),
			sdk.NewAttribute(types.AttributeKeyConsensusHeight, fmt.Sprintf("%d", misbehaviour.GetHeight())),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}
