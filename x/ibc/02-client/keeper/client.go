package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
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
		return nil, sdkerrors.Wrapf(errors.ErrClientExists, "cannot create client with ID %s", clientID)
	}

	_, found = k.GetClientType(ctx, clientID)
	if found {
		panic(fmt.Sprintf("consensus type is already defined for client %s", clientID))
	}

	clientState, err := k.initialize(ctx, clientID, clientType, consensusState)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "cannot create client with ID %s", clientID)
	}

	k.SetCommitter(ctx, clientID, consensusState.GetHeight(), consensusState.GetCommitter())
	k.SetVerifiedRoot(ctx, clientID, consensusState.GetHeight(), consensusState.GetRoot())
	k.SetClientState(ctx, clientState)
	k.SetClientType(ctx, clientID, clientType)
	k.Logger(ctx).Info(fmt.Sprintf("client %s created at height %d", clientID, consensusState.GetHeight()))
	return clientState, nil
}

// UpdateClient updates the consensus state and the state root from a provided header
func (k Keeper) UpdateClient(ctx sdk.Context, clientID string, header exported.Header) error {
	clientType, found := k.GetClientType(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(errors.ErrClientTypeNotFound, "cannot update client with ID %s", clientID)
	}

	// check that the header consensus matches the client one
	if header.ClientType() != clientType {
		return sdkerrors.Wrapf(errors.ErrInvalidConsensus, "cannot update client with ID %s", clientID)
	}

	clientState, found := k.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(errors.ErrClientNotFound, "cannot update client with ID %s", clientID)
	}

	if clientState.IsFrozen() {
		return sdkerrors.Wrapf(errors.ErrClientFrozen, "cannot update client with ID %s", clientID)
	}

	consensusState, found := k.GetConsensusState(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(errors.ErrConsensusStateNotFound, "cannot update client with ID %s", clientID)
	}

	consensusState, err := consensusState.CheckValidityAndUpdateState(header)
	if err != nil {
		return sdkerrors.Wrapf(err, "cannot update client with ID %s", clientID)
	}

	k.SetConsensusState(ctx, clientID, consensusState)
	k.SetCommitter(ctx, clientID, consensusState.GetHeight(), consensusState.GetCommitter())
	k.SetVerifiedRoot(ctx, clientID, consensusState.GetHeight(), consensusState.GetRoot())
	k.Logger(ctx).Info(fmt.Sprintf("client %s updated to height %d", clientID, consensusState.GetHeight()))
	return nil
}

// CheckMisbehaviourAndUpdateState checks for client misbehaviour and freezes the
// client if so.
func (k Keeper) CheckMisbehaviourAndUpdateState(ctx sdk.Context, misbehaviour exported.Misbehaviour) error {
	clientState, found := k.GetClientState(ctx, misbehaviour.GetClientID())
	if !found {
		return sdkerrors.Wrap(errors.ErrClientNotFound, misbehaviour.GetClientID())
	}

	committer, found := k.GetCommitter(ctx, misbehaviour.GetClientID(), uint64(misbehaviour.GetHeight()))
	if !found {
		return errors.ErrCommitterNotFound
	}

	var err error
	switch e := misbehaviour.(type) {
	case tendermint.Evidence:
		clientState, err = tendermint.CheckMisbehaviourAndUpdateState(clientState, committer, misbehaviour)

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
