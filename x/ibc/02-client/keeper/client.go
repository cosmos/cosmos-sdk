package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// CreateClient creates a new client state and populates it with a given consensus
// state as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
//
// CONTRACT: ClientState was constructed correctly from given initial consensusState
func (k Keeper) CreateClient(
	ctx sdk.Context, clientID string, clientState exported.ClientState, consensusState exported.ConsensusState,
) error {
	_, found := k.GetClientState(ctx, clientID)
	if found {
		return sdkerrors.Wrapf(types.ErrClientExists, "cannot create client with ID %s", clientID)
	}

	if consensusState != nil {
		k.SetClientConsensusState(ctx, clientID, clientState.GetLatestHeight(), consensusState)
	}

	k.SetClientState(ctx, clientID, clientState)
	k.Logger(ctx).Info(fmt.Sprintf("client %s created at height %d", clientID, clientState.GetLatestHeight()))

	return nil
}

// UpdateClient updates the consensus state and the state root from a provided header.
func (k Keeper) UpdateClient(ctx sdk.Context, clientID string, header exported.Header) error {
	clientState, found := k.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(types.ErrClientNotFound, "cannot update client with ID %s", clientID)
	}

	// prevent update if the client is frozen before or at header height
	if clientState.IsFrozen() && clientState.GetFrozenHeight().LTE(header.GetHeight()) {
		return sdkerrors.Wrapf(types.ErrClientFrozen, "cannot update client with ID %s", clientID)
	}

	var (
		consensusState  exported.ConsensusState
		consensusHeight exported.Height
		err             error
	)

	clientState, consensusState, err = clientState.CheckHeaderAndUpdateState(ctx, k.cdc, k.ClientStore(ctx, clientID), header)

	if err != nil {
		return sdkerrors.Wrapf(err, "cannot update client with ID %s", clientID)
	}

	k.SetClientState(ctx, clientID, clientState)

	// we don't set consensus state for localhost client
	if header != nil && clientID != exported.Localhost {
		k.SetClientConsensusState(ctx, clientID, header.GetHeight(), consensusState)
		consensusHeight = header.GetHeight()
	} else {
		consensusHeight = types.GetSelfHeight(ctx)
	}

	k.Logger(ctx).Info(fmt.Sprintf("client %s updated height %d", clientID, consensusHeight))

	// emitting events in the keeper emits for both begin block and handler client updates
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateClient,
			sdk.NewAttribute(types.AttributeKeyClientID, clientID),
			sdk.NewAttribute(types.AttributeKeyClientType, clientState.ClientType()),
			sdk.NewAttribute(types.AttributeKeyConsensusHeight, consensusHeight.String()),
		),
	)

	return nil
}

// CheckMisbehaviourAndUpdateState checks for client misbehaviour and freezes the
// client if so.
func (k Keeper) CheckMisbehaviourAndUpdateState(ctx sdk.Context, misbehaviour exported.Misbehaviour) error {
	clientState, found := k.GetClientState(ctx, misbehaviour.GetClientID())
	if !found {
		return sdkerrors.Wrapf(types.ErrClientNotFound, "cannot check misbehaviour for client with ID %s", misbehaviour.GetClientID())
	}

	clientState, err := clientState.CheckMisbehaviourAndUpdateState(ctx, k.cdc, k.ClientStore(ctx, misbehaviour.GetClientID()), misbehaviour)
	if err != nil {
		return err
	}

	k.SetClientState(ctx, misbehaviour.GetClientID(), clientState)
	k.Logger(ctx).Info(fmt.Sprintf("client %s frozen due to misbehaviour", misbehaviour.GetClientID()))

	return nil
}
