package client

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
)

// HandleMsgCreateClient defines the sdk.Handler for MsgCreateClient
func HandleMsgCreateClient(ctx sdk.Context, k Keeper, msg exported.MsgCreateClient) (*sdk.Result, error) {
	clientType := exported.ClientTypeFromString(msg.GetClientType())

	var (
		clientState     exported.ClientState
		consensusHeight uint64
	)

	switch clientType {
	case exported.Localhost:
		// msg client id is always "localhost"
		clientState = localhosttypes.NewClientState(ctx.ChainID(), ctx.BlockHeight())
		consensusHeight = uint64(ctx.BlockHeight())
	default:
		var err error
		clientState, err = msg.InitializeClientState()
		if err != nil {
			return nil, err
		}
		consensusHeight = msg.GetConsensusState().GetHeight()
	}

	_, err := k.CreateClient(
		ctx, clientState, msg.GetConsensusState(),
	)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeCreateClient,
			sdk.NewAttribute(AttributeKeyClientID, msg.GetClientID()),
			sdk.NewAttribute(AttributeKeyClientType, msg.GetClientType()),
			sdk.NewAttribute(types.AttributeKeyConsensusHeight, fmt.Sprintf("%d", consensusHeight)),
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

// HandleMsgUpdateClient defines the sdk.Handler for MsgUpdateClient
func HandleMsgUpdateClient(ctx sdk.Context, k Keeper, msg exported.MsgUpdateClient) (*sdk.Result, error) {
	_, err := k.UpdateClient(ctx, msg.GetClientID(), msg.GetHeader())
	if err != nil {
		return nil, err
	}

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandlerClientMisbehaviour defines the Evidence module handler for submitting a
// light client misbehaviour.
func HandlerClientMisbehaviour(k Keeper) evidencetypes.Handler {
	return func(ctx sdk.Context, evidence evidenceexported.Evidence) error {
		misbehaviour, ok := evidence.(exported.Misbehaviour)
		if !ok {
			return sdkerrors.Wrapf(types.ErrInvalidEvidence,
				"expected evidence to implement client Misbehaviour interface, got %T", evidence,
			)
		}

		return k.CheckMisbehaviourAndUpdateState(ctx, misbehaviour)
	}
}
