package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// CreateClient creates a new client state and populates it with a given consensus
// state as defined in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func (k Keeper) CreateClient(
	ctx sdk.Context, clientID string,
	clientType exported.ClientType, consensusState exported.ConsensusState,
) (types.State, error) {
	_, found := k.GetClientState(ctx, clientID)
	if found {
		return types.State{}, types.ErrClientExists(k.codespace, clientID)
	}

	_, found = k.GetClientType(ctx, clientID)
	if found {
		panic(fmt.Sprintf("consensus type is already defined for client %s", clientID))
	}

	clientState := k.initialize(ctx, clientID, consensusState)
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
		return sdkerrors.Wrap(types.ErrClientTypeNotFound(k.codespace), "cannot update client")
	}

	// check that the header consensus matches the client one
	if header.ClientType() != clientType {
		return sdkerrors.Wrap(types.ErrInvalidConsensus(k.codespace), "cannot update client")
	}

	clientState, found := k.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(types.ErrClientNotFound(k.codespace, clientID), "cannot update client")
	}

	if clientState.Frozen {
		return sdkerrors.Wrap(types.ErrClientFrozen(k.codespace, clientID), "cannot update client")
	}

	consensusState, found := k.GetConsensusState(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(types.ErrConsensusStateNotFound(k.codespace), "cannot update client")
	}

	if header.GetHeight() < consensusState.GetHeight() {
		return sdkerrors.Wrap(
			sdk.ErrInvalidSequence(
				fmt.Sprintf("header height < consensus height (%d < %d)", header.GetHeight(), consensusState.GetHeight()),
			),
			"cannot update client",
		)
	}

	consensusState, err := consensusState.CheckValidityAndUpdateState(header)
	if err != nil {
		return sdkerrors.Wrap(err, "cannot update client")
	}

	k.SetConsensusState(ctx, clientID, consensusState)
	k.SetVerifiedRoot(ctx, clientID, consensusState.GetHeight(), consensusState.GetRoot())
	k.Logger(ctx).Info(fmt.Sprintf("client %s updated to height %d", clientID, consensusState.GetHeight()))
	return nil
}

// CheckMisbehaviourAndUpdateState checks for client misbehaviour and freezes the
// client if so.
func (k Keeper) CheckMisbehaviourAndUpdateState(ctx sdk.Context, clientID string, evidence exported.Evidence) error {
	clientState, found := k.GetClientState(ctx, clientID)
	if !found {
		sdk.ResultFromError(types.ErrClientNotFound(k.codespace, clientID))
	}

	err := k.checkMisbehaviour(ctx, evidence)
	if err != nil {
		return err
	}

	clientState, err = k.freeze(ctx, clientState)
	if err != nil {
		return err
	}

	k.SetClientState(ctx, clientState)
	k.Logger(ctx).Info(fmt.Sprintf("client %s frozen due to misbehaviour", clientID))
	return nil
}
