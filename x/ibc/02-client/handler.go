package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// HandleMsgCreateClient defines the sdk.Handler for MsgCreateClient
func HandleMsgCreateClient(ctx sdk.Context, k Keeper, msg exported.MsgCreateClient) (*sdk.Result, error) {
	clientType := exported.ClientTypeFromString(msg.GetClientType())
	var clientState exported.ClientState
	switch clientType {
	case 0:
		return nil, sdkerrors.Wrap(ErrInvalidClientType, msg.GetClientType())
	case exported.Tendermint:
		tmMsg, ok := msg.(ibctmtypes.MsgCreateClient)
		if !ok {
			return nil, sdkerrors.Wrap(ErrInvalidClientType, "Msg is not a Tendermint CreateClient msg")
		}
		var err error
		clientState, err = ibctmtypes.InitializeFromMsg(tmMsg)
		if err != nil {
			return nil, err
		}
	default:
		return nil, sdkerrors.Wrap(ErrInvalidClientType, msg.GetClientType())
	}

	_, err := k.CreateClient(
		ctx, clientState, msg.GetConsensusState(),
	)
	if err != nil {
		return nil, err
	}

	attributes := make([]sdk.Attribute, len(msg.GetSigners())+1)
	attributes[0] = sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory)
	for i, signer := range msg.GetSigners() {
		attributes[i+1] = sdk.NewAttribute(sdk.AttributeKeySender, signer.String())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeCreateClient,
			sdk.NewAttribute(AttributeKeyClientID, msg.GetClientID()),
			sdk.NewAttribute(AttrbuteKeyClientType, msg.GetClientType()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			attributes...,
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgUpdateClient defines the sdk.Handler for MsgUpdateClient
func HandleMsgUpdateClient(ctx sdk.Context, k Keeper, msg exported.MsgUpdateClient) (*sdk.Result, error) {
	if err := k.UpdateClient(ctx, msg.GetClientID(), msg.GetHeader()); err != nil {
		return nil, err
	}

	attributes := make([]sdk.Attribute, len(msg.GetSigners())+1)
	attributes[0] = sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory)
	for i, signer := range msg.GetSigners() {
		attributes[i+1] = sdk.NewAttribute(sdk.AttributeKeySender, signer.String())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeUpdateClient,
			sdk.NewAttribute(AttributeKeyClientID, msg.GetClientID()),
			sdk.NewAttribute(AttrbuteKeyClientType, msg.GetHeader().ClientType().String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			attributes...,
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandlerClientMisbehaviour defines the Evidence module handler for submitting a
// light client misbehaviour.
func HandlerClientMisbehaviour(k Keeper) evidence.Handler {
	return func(ctx sdk.Context, evidence evidenceexported.Evidence) error {
		misbehaviour, ok := evidence.(exported.Misbehaviour)
		if !ok {
			return types.ErrInvalidEvidence
		}

		return k.CheckMisbehaviourAndUpdateState(ctx, misbehaviour)
	}
}
