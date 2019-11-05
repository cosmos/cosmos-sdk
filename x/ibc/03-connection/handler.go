package connection

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// HandleMsgConnectionOpenInit defines the sdk.Handler for MsgConnectionOpenInit
func HandleMsgConnectionOpenInit(ctx sdk.Context, k keeper.Keeper, msg types.MsgConnectionOpenInit) sdk.Result {
	err := k.ConnOpenInit(ctx, msg.ConnectionID, msg.ClientID, msg.Counterparty)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenInit,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionID),
			sdk.NewAttribute(types.AttributeKeyCounterpartyClientID, msg.Counterparty.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

// HandleMsgConnectionOpenTry defines the sdk.Handler for MsgConnectionOpenTry
func HandleMsgConnectionOpenTry(ctx sdk.Context, k keeper.Keeper, msg types.MsgConnectionOpenTry) sdk.Result {
	err := k.ConnOpenTry(
		ctx, msg.ConnectionID, msg.Counterparty, msg.ClientID,
		msg.CounterpartyVersions, msg.ProofInit, msg.ProofConsensus, msg.ProofHeight, msg.ConsensusHeight)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenTry,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionID),
			sdk.NewAttribute(types.AttributeKeyCounterpartyClientID, msg.Counterparty.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

// HandleMsgConnectionOpenAck defines the sdk.Handler for MsgConnectionOpenAck
func HandleMsgConnectionOpenAck(ctx sdk.Context, k keeper.Keeper, msg types.MsgConnectionOpenAck) sdk.Result {
	err := k.ConnOpenAck(
		ctx, msg.ConnectionID, msg.Version, msg.ProofTry, msg.ProofConsensus,
		msg.ProofHeight, msg.ConsensusHeight,
	)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenAck,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}

// HandleMsgConnectionOpenConfirm defines the sdk.Handler for MsgConnectionOpenConfirm
func HandleMsgConnectionOpenConfirm(ctx sdk.Context, k keeper.Keeper, msg types.MsgConnectionOpenConfirm) sdk.Result {
	err := k.ConnOpenConfirm(ctx, msg.ConnectionID, msg.ProofAck, msg.ProofHeight)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenConfirm,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}
