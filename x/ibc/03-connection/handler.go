package connection

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HandleMsgConnectionOpenInit defines the sdk.Handler for MsgConnectionOpenInit
func HandleMsgConnectionOpenInit(ctx sdk.Context, k Keeper, msg MsgConnectionOpenInit) (*sdk.Result, error) {
	if err := k.ConnOpenInit(
		ctx, msg.ConnectionID, msg.ClientID, msg.Counterparty,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeConnectionOpenInit,
			sdk.NewAttribute(AttributeKeyConnectionID, msg.ConnectionID),
			sdk.NewAttribute(AttributeKeyCounterpartyClientID, msg.Counterparty.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

// HandleMsgConnectionOpenTry defines the sdk.Handler for MsgConnectionOpenTry
func HandleMsgConnectionOpenTry(ctx sdk.Context, k Keeper, msg MsgConnectionOpenTry) (*sdk.Result, error) {
	if err := k.ConnOpenTry(
		ctx, msg.ConnectionID, msg.Counterparty, msg.ClientID,
		msg.CounterpartyVersions, msg.ProofInit, msg.ProofConsensus,
		msg.ProofHeight, msg.ConsensusHeight,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeConnectionOpenTry,
			sdk.NewAttribute(AttributeKeyConnectionID, msg.ConnectionID),
			sdk.NewAttribute(AttributeKeyCounterpartyClientID, msg.Counterparty.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

// HandleMsgConnectionOpenAck defines the sdk.Handler for MsgConnectionOpenAck
func HandleMsgConnectionOpenAck(ctx sdk.Context, k Keeper, msg MsgConnectionOpenAck) (*sdk.Result, error) {
	if err := k.ConnOpenAck(
		ctx, msg.ConnectionID, msg.Version, msg.ProofTry, msg.ProofConsensus,
		msg.ProofHeight, msg.ConsensusHeight,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeConnectionOpenAck,
			sdk.NewAttribute(AttributeKeyConnectionID, msg.ConnectionID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

// HandleMsgConnectionOpenConfirm defines the sdk.Handler for MsgConnectionOpenConfirm
func HandleMsgConnectionOpenConfirm(ctx sdk.Context, k Keeper, msg MsgConnectionOpenConfirm) (*sdk.Result, error) {
	if err := k.ConnOpenConfirm(
		ctx, msg.ConnectionID, msg.ProofAck, msg.ProofHeight,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeConnectionOpenConfirm,
			sdk.NewAttribute(AttributeKeyConnectionID, msg.ConnectionID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}
