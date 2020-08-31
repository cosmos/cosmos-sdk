package client

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
)

// HandleMsgCreateClient defines the sdk.Handler for MsgCreateClient
func HandleMsgCreateClient(ctx sdk.Context, k keeper.Keeper, msg exported.MsgCreateClient) (*sdk.Result, error) {
	var (
		consensusHeight uint64
		clientState     exported.ClientState
	)

	switch msg.(type) {
	// localhost is a special case that must initialize client state
	// from context and not from msg
	case *localhosttypes.MsgCreateClient:
		selfHeight := types.NewHeight(0, uint64(ctx.BlockHeight()))
		clientState = localhosttypes.NewClientState(ctx.ChainID(), selfHeight)
		// Localhost consensus height is chain's blockheight
		consensusHeight = uint64(ctx.BlockHeight())
	default:
		clientState = msg.InitializeClientState()
		if consState := msg.GetConsensusState(); consState != nil {
			consensusHeight = consState.GetHeight()
		}
	}

	_, err := k.CreateClient(
		ctx, msg.GetClientID(), clientState, msg.GetConsensusState(),
	)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateClient,
			sdk.NewAttribute(types.AttributeKeyClientID, msg.GetClientID()),
			sdk.NewAttribute(types.AttributeKeyClientType, msg.GetClientType()),
			sdk.NewAttribute(types.AttributeKeyConsensusHeight, fmt.Sprintf("%d", consensusHeight)),
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
func HandleMsgUpdateClient(ctx sdk.Context, k keeper.Keeper, msg exported.MsgUpdateClient) (*sdk.Result, error) {
	_, err := k.UpdateClient(ctx, msg.GetClientID(), msg.GetHeader())
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
func HandleMsgSubmitMisbehaviour(ctx sdk.Context, k keeper.Keeper, msg exported.MsgSubmitMisbehaviour) (*sdk.Result, error) {
	misbehaviour := msg.GetMisbehaviour()
	if misbehaviour == nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidMisbehaviour, "misbehaviour is nil")
	}

	if err := misbehaviour.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.CheckMisbehaviourAndUpdateState(ctx, misbehaviour); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to process misbehaviour for IBC client")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitMisbehaviour,
			sdk.NewAttribute(types.AttributeKeyClientID, misbehaviour.GetClientID()),
			sdk.NewAttribute(types.AttributeKeyClientType, misbehaviour.ClientType().String()),
			sdk.NewAttribute(types.AttributeKeyConsensusHeight, fmt.Sprintf("%d", misbehaviour.GetHeight())),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}
