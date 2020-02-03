package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
)

// CreateClient creates a new client state and populates it with a given consensus
// state as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func (k Keeper) CreateClient(
	ctx sdk.Context, clientID string,
	clientType exported.ClientType, consensusState exported.ConsensusState,
) (exported.ClientState, error) {
	_, found := k.GetClientState(ctx, clientID)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrClientExists, "cannot create client with ID %s", clientID)
	}

	_, found = k.GetClientType(ctx, clientID)
	if found {
		panic(fmt.Sprintf("client type is already defined for client %s", clientID))
	}

	clientState, err := k.initialize(ctx, clientID, clientType, consensusState)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "cannot create client with ID %s", clientID)
	}

	k.SetClientState(ctx, clientState)
	k.SetClientType(ctx, clientID, clientType)
	k.Logger(ctx).Info(fmt.Sprintf("client %s created at height %d", clientID, clientState.GetLatestHeight()))

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateClient,
			sdk.NewAttribute(types.AttributeKeyClientID, clientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return clientState, nil
}

// UpdateClient updates the consensus state and the state root from a provided header
func (k Keeper) UpdateClient(
	ctx sdk.Context, clientID string, oldHeader, newHeader exported.Header,
) error {
	clientType, found := k.GetClientType(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(types.ErrClientTypeNotFound, "cannot update client with ID %s", clientID)
	}

	// check that the headers client type matches the stored one
	if oldHeader.ClientType() != clientType {
		return sdkerrors.Wrapf(
			types.ErrInvalidClientType,
			"old header type ≠ client type (%s ≠ %s)", oldHeader.ClientType().String(), clientType.String(),
		)
	}

	if newHeader.ClientType() != clientType {
		return sdkerrors.Wrapf(
			types.ErrInvalidClientType,
			"new header type ≠ client type (%s ≠ %s)", oldHeader.ClientType().String(), clientType.String(),
		)
	}

	clientState, found := k.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(types.ErrClientNotFound, "cannot update client with ID %s", clientID)
	}

	// addittion to spec: prevent update if the client is frozen
	if clientState.IsFrozen() {
		return sdkerrors.Wrapf(types.ErrClientFrozen, "cannot update client with ID %s", clientID)
	}

	var (
		consensusState exported.ConsensusState
		err            error
	)

	switch clientType {
	case exported.Tendermint:
		clientState, consensusState, err = tendermint.CheckValidityAndUpdateState(
			clientState, oldHeader, newHeader, ctx.ChainID(), ctx.BlockTime(),
		)
	default:
		err = types.ErrInvalidClientType
	}

	if err != nil {
		return sdkerrors.Wrapf(err, "cannot update client with ID %s", clientID)
	}

	k.SetClientState(ctx, clientState)
	k.SetClientConsensusState(ctx, clientID, newHeader.GetHeight(), consensusState)
	k.Logger(ctx).Info(fmt.Sprintf("client %s updated to height %d", clientID, newHeader.GetHeight()))

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUpdateClient,
			sdk.NewAttribute(types.AttributeKeyClientID, clientID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return nil
}

// CheckMisbehaviourAndUpdateState checks for client misbehaviour and freezes the
// client if so.
func (k Keeper) CheckMisbehaviourAndUpdateState(ctx sdk.Context, misbehaviour exported.Misbehaviour) error {
	clientState, found := k.GetClientState(ctx, misbehaviour.GetClientID())
	if !found {
		return sdkerrors.Wrap(types.ErrClientNotFound, misbehaviour.GetClientID())
	}

	consensusState, found := k.GetClientConsensusState(ctx, misbehaviour.GetClientID(), uint64(misbehaviour.GetHeight()))
	if !found {
		return sdkerrors.Wrap(types.ErrConsensusStateNotFound, misbehaviour.GetClientID())
	}

	var err error
	switch e := misbehaviour.(type) {
	case tendermint.Evidence:
		clientState, err = tendermint.CheckMisbehaviourAndUpdateState(
			clientState, consensusState, misbehaviour, uint64(misbehaviour.GetHeight()), ctx.BlockTime(),
		)

	default:
		err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC client evidence type: %T", e)
	}

	if err != nil {
		return err
	}

	k.SetClientState(ctx, clientState)
	k.Logger(ctx).Info(fmt.Sprintf("client %s frozen due to misbehaviour", misbehaviour.GetClientID()))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitMisbehaviour,
			sdk.NewAttribute(types.AttributeKeyClientID, misbehaviour.GetClientID()),
			sdk.NewAttribute(types.AttributeKeyClientType, misbehaviour.ClientType().String()),
		),
	)

	return nil
}
