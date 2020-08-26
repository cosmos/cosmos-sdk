package client

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// HandleMsgCreateClient defines the sdk.Handler for MsgCreateClient
func HandleMsgCreateClient(ctx sdk.Context, k keeper.Keeper, msg *types.MsgCreateClient) (*sdk.Result, error) {
	clientState := types.MustUnpackClientState(msg.ClientState)
	consensusState := types.MustUnpackConsensusState(msg.ConsensusState)

	_, err := k.CreateClient(ctx, msg.ClientId, clientState, consensusState)
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
	header := types.MustUnpackHeader(msg.Header)

	_, err := k.UpdateClient(ctx, msg.ClientId, header)
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

// HandlerClientMisbehaviour defines the Evidence module handler for submitting a
// light client misbehaviour.
func HandlerClientMisbehaviour(k keeper.Keeper) evidencetypes.Handler {
	return func(ctx sdk.Context, evidence evidenceexported.Evidence) error {
		misbehaviour, ok := evidence.(exported.Misbehaviour)
		if !ok {
			return sdkerrors.Wrapf(types.ErrInvalidEvidence,
				"expected evidence to implement client Misbehaviour interface, got %T", evidence,
			)
		}

		if err := k.CheckMisbehaviourAndUpdateState(ctx, misbehaviour); err != nil {
			return sdkerrors.Wrap(err, "failed to process misbehaviour for IBC client")
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSubmitMisbehaviour,
				sdk.NewAttribute(types.AttributeKeyClientID, misbehaviour.GetClientID()),
				sdk.NewAttribute(types.AttributeKeyClientType, misbehaviour.ClientType().String()),
				sdk.NewAttribute(types.AttributeKeyConsensusHeight, fmt.Sprintf("%d", uint64(misbehaviour.GetHeight()))),
			),
		)
		return nil
	}
}
