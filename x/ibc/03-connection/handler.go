package connection

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
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
			types.EventTypeConnectionOpenInit,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionID),
			sdk.NewAttribute(types.AttributeKeyClientID, msg.ClientID),
			sdk.NewAttribute(types.AttributeKeyCounterpartyClientID, msg.Counterparty.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgConnectionOpenTry defines the sdk.Handler for MsgConnectionOpenTry
func HandleMsgConnectionOpenTry(ctx sdk.Context, k Keeper, msg MsgConnectionOpenTry) (*sdk.Result, error) {
	proofInit, err := commitmenttypes.UnpackAnyProof(&msg.ProofInit)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof init")
	}

	proofConsensus, err := commitmenttypes.UnpackAnyProof(&msg.ProofConsensus)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof consensus")
	}

	if err := k.ConnOpenTry(
		ctx, msg.ConnectionID, msg.Counterparty, msg.ClientID,
		msg.CounterpartyVersions, proofInit, proofConsensus,
		msg.ProofHeight, msg.ConsensusHeight,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenTry,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionID),
			sdk.NewAttribute(types.AttributeKeyClientID, msg.ClientID),
			sdk.NewAttribute(types.AttributeKeyCounterpartyClientID, msg.Counterparty.ClientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgConnectionOpenAck defines the sdk.Handler for MsgConnectionOpenAck
func HandleMsgConnectionOpenAck(ctx sdk.Context, k Keeper, msg MsgConnectionOpenAck) (*sdk.Result, error) {
	proofTry, err := commitmenttypes.UnpackAnyProof(&msg.ProofTry)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof try")
	}

	proofConsensus, err := commitmenttypes.UnpackAnyProof(&msg.ProofConsensus)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof consensus")
	}

	if err := k.ConnOpenAck(
		ctx, msg.ConnectionID, msg.Version, proofTry, proofConsensus,
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
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgConnectionOpenConfirm defines the sdk.Handler for MsgConnectionOpenConfirm
func HandleMsgConnectionOpenConfirm(ctx sdk.Context, k Keeper, msg MsgConnectionOpenConfirm) (*sdk.Result, error) {
	proofAck, err := commitmenttypes.UnpackAnyProof(&msg.ProofAck)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof ack")
	}

	if err := k.ConnOpenConfirm(
		ctx, msg.ConnectionID, proofAck, msg.ProofHeight,
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
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}
