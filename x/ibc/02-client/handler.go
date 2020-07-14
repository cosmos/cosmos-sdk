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
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
)

// HandleMsgCreateClient defines the sdk.Handler for MsgCreateClient
func HandleMsgCreateClient(ctx sdk.Context, k keeper.Keeper, msg exported.MsgCreateClient) (*sdk.Result, error) {
	clientType := exported.ClientTypeFromString(msg.GetClientType())

	var (
		clientState     exported.ClientState
		consensusHeight uint64
	)

	switch clientType {
	case exported.Tendermint:
		tmMsg, ok := msg.(ibctmtypes.MsgCreateClient)
		if !ok {
			return nil, sdkerrors.Wrapf(types.ErrInvalidClientType, "got %T, expected %T", msg, ibctmtypes.MsgCreateClient{})
		}
		var err error

		clientState, err = ibctmtypes.InitializeFromMsg(tmMsg)
		if err != nil {
			return nil, err
		}
		consensusHeight = msg.GetConsensusState().GetHeight()
	case exported.Localhost:
		// msg client id is always "localhost"
		clientState = localhosttypes.NewClientState(ctx.ChainID(), ctx.BlockHeight())
		consensusHeight = uint64(ctx.BlockHeight())
	default:
		return nil, sdkerrors.Wrapf(types.ErrInvalidClientType, "unsupported client type (%s)", msg.GetClientType())
	}

	_, err := k.CreateClient(
		ctx, clientState, msg.GetConsensusState(),
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

		return k.CheckMisbehaviourAndUpdateState(ctx, misbehaviour)
	}
}
